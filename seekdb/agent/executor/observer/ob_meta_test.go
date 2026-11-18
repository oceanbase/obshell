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

package observer

import (
	"bytes"
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func TestRetryOBMetaAfterTransientFailures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var output bytes.Buffer
	logger := log.New()
	logger.SetOutput(&output)
	failures := []error{driver.ErrBadConn, sqlite3.Error{Code: sqlite3.ErrBusy}}
	attempts := 0
	err := retryOBMeta(ctx, time.Millisecond, func() error {
		attempts++
		if attempts <= len(failures) {
			return failures[attempts-1]
		}
		return nil
	}, log.NewEntry(logger))
	if err != nil {
		t.Fatalf("metadata synchronization stopped after %d attempts: %v", attempts, err)
	}
	if attempts != len(failures)+1 {
		t.Fatalf("synchronization attempted %d times, want %d", attempts, len(failures)+1)
	}
	for _, failure := range failures {
		if !strings.Contains(output.String(), failure.Error()) {
			t.Errorf("synchronization failure %q missing from log: %s", failure, &output)
		}
	}
}

func TestRetryOBMetaStopsWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := log.New()
	logger.SetOutput(io.Discard)
	attempted := make(chan struct{})
	finished := make(chan error, 1)
	attempts := 0
	go func() {
		finished <- retryOBMeta(ctx, time.Hour, func() error {
			attempts++
			close(attempted)
			return driver.ErrBadConn
		}, log.NewEntry(logger))
	}()
	select {
	case <-attempted:
	case <-time.After(time.Second):
		t.Fatal("synchronization was not attempted")
	}
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled synchronization returned %v", err)
		}
		if attempts != 1 {
			t.Fatalf("canceled synchronization attempted %d times, want 1", attempts)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not interrupt the retry wait")
	}
}

func TestRetryOBMetaDoesNotStartAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	logger := log.New()
	logger.SetOutput(io.Discard)
	attempted := false
	err := retryOBMeta(ctx, time.Hour, func() error {
		attempted = true
		return nil
	}, log.NewEntry(logger))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled synchronization returned %v", err)
	}
	if attempted {
		t.Fatal("synchronization started after cancellation")
	}
}
