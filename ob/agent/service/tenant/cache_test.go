/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tenant

import (
	"context"
	stderrors "errors"
	"testing"
	"time"
)

func TestPasswordMapWaitingUpdateCanBeCanceled(t *testing.T) {
	passwords := &PasswordMap{}
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	// This test owns the holder; cleanup releases it and joins its result.
	go func() {
		done <- passwords.SetValidated(context.Background(), "tenant1", "valid", func() error {
			close(entered)
			<-release
			return nil
		})
	}()
	t.Cleanup(func() {
		close(release)
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("lock holder failed: %v", err)
			}
		case <-time.After(time.Second):
			t.Error("lock holder did not exit")
		}
	})
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("lock holder did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := passwords.SetValidated(ctx, "tenant1", "replacement", func() error {
		t.Error("expired waiter must not validate or store")
		return nil
	})
	if !stderrors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting update error = %v", err)
	}
}

func TestPasswordMapSetValidatedDoesNotStoreWhenValidationFails(t *testing.T) {
	passwords := &PasswordMap{}
	passwords.Set("tenant1", "existing")
	validationErr := stderrors.New("validation failed")

	err := passwords.SetValidated(context.Background(), "tenant1", "replacement", func() error {
		return validationErr
	})
	if !stderrors.Is(err, validationErr) {
		t.Fatalf("SetValidated error = %v, want %v", err, validationErr)
	}
	password, ok := passwords.Get("tenant1")
	if !ok {
		t.Fatal("existing password was removed after failed validation")
	}
	if password != "existing" {
		t.Fatalf("password = %q, want %q", password, "existing")
	}
}

func TestPasswordMapSetValidatedSerializesSameTenantValidationAndStore(t *testing.T) {
	passwords := &PasswordMap{}
	oldValidationEntered := make(chan struct{})
	releaseOldValidation := make(chan struct{})
	newCallStarted := make(chan struct{})
	newValidationEntered := make(chan struct{})
	results := make(chan error, 2)

	go func() {
		err := passwords.SetValidated(context.Background(), "tenant1", "old", func() error {
			close(oldValidationEntered)
			<-releaseOldValidation
			return nil
		})
		results <- err
	}()

	select {
	case <-oldValidationEntered:
	case <-time.After(time.Second):
		close(releaseOldValidation)
		t.Fatal("old validation did not start")
	}

	go func() {
		close(newCallStarted)
		err := passwords.SetValidated(context.Background(), "tenant1", "new", func() error {
			close(newValidationEntered)
			return nil
		})
		results <- err
	}()

	select {
	case <-newCallStarted:
	case <-time.After(time.Second):
		close(releaseOldValidation)
		t.Fatal("new SetValidated call did not start")
	}

	newValidationRanEarly := false
	select {
	case <-newValidationEntered:
		newValidationRanEarly = true
	case <-time.After(50 * time.Millisecond):
	}
	close(releaseOldValidation)

	for i := 0; i < 2; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Fatalf("SetValidated returned error: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("SetValidated goroutine did not finish")
		}
	}
	if newValidationRanEarly {
		t.Fatal("new validation ran before the older validation and store completed")
	}
	select {
	case <-newValidationEntered:
	default:
		t.Fatal("new validation never ran after the old update completed")
	}

	password, ok := passwords.Get("tenant1")
	if !ok {
		t.Fatal("password was not stored")
	}
	if password != "new" {
		t.Fatalf("password = %q, want %q; an older update overwrote the newer update", password, "new")
	}
}
