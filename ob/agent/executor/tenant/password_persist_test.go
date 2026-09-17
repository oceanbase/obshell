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
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	oberrors "github.com/oceanbase/obshell/ob/agent/errors"
	libhttp "github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/meta"
	tenantservice "github.com/oceanbase/obshell/ob/agent/service/tenant"
)

func TestPasswordPersisterRetriesAgainstChangedMaintainer(t *testing.T) {
	self := meta.NewAgentInfo("127.0.0.1", 2881)
	staleMaintainer := meta.NewAgentInfo("127.0.0.2", 2882)
	currentMaintainer := meta.NewAgentInfo("127.0.0.3", 2883)
	transientErr := stderrors.New("stale maintainer rejected request")

	resolveCalls := 0
	forwardedTo := make([]string, 0, 2)
	persister := passwordPersister{
		self:  self,
		cache: &tenantservice.PasswordMap{},
		resolve: func(context.Context) (meta.AgentInfoInterface, error) {
			resolveCalls++
			if resolveCalls == 1 {
				return staleMaintainer, nil
			}
			return currentMaintainer, nil
		},
		check: func(context.Context, string, string) (bool, error) {
			t.Fatal("a forwarding sender must not validate the tenant password locally")
			return false, nil
		},
		forward: func(_ context.Context, maintainer meta.AgentInfoInterface, _, _ string) error {
			forwardedTo = append(forwardedTo, maintainer.String())
			if samePasswordAgent(maintainer, staleMaintainer) {
				return transientErr
			}
			return nil
		},
		delay: 0,
	}

	if err := persister.persist(context.Background(), "tenant1", "password", true); err != nil {
		t.Fatalf("persist after maintainer switch returned error: %v", err)
	}
	if resolveCalls != 2 {
		t.Fatalf("resolve call count = %d, want 2", resolveCalls)
	}
	if len(forwardedTo) != 2 {
		t.Fatalf("forward call count = %d, want 2", len(forwardedTo))
	}
	if forwardedTo[0] != staleMaintainer.String() || forwardedTo[1] != currentMaintainer.String() {
		t.Fatalf("forward targets = %v, want [%s %s]", forwardedTo, staleMaintainer.String(), currentMaintainer.String())
	}
}

func TestPasswordPersisterForwardedReceiverDoesNotForwardOrRetry(t *testing.T) {
	self := meta.NewAgentInfo("127.0.0.1", 2881)
	otherMaintainer := meta.NewAgentInfo("127.0.0.2", 2882)
	resolveCalls := 0
	forwardCalls := 0
	persister := passwordPersister{
		self:  self,
		cache: &tenantservice.PasswordMap{},
		resolve: func(context.Context) (meta.AgentInfoInterface, error) {
			resolveCalls++
			return otherMaintainer, nil
		},
		check: func(context.Context, string, string) (bool, error) {
			t.Fatal("a non-authoritative receiver must not validate the password")
			return false, nil
		},
		forward: func(context.Context, meta.AgentInfoInterface, string, string) error {
			forwardCalls++
			return nil
		},
		delay: 0,
	}

	err := persister.persist(context.Background(), "tenant1", "password", false)
	assertAgentErrorCode(t, err, oberrors.ErrAgentMaintainerNotActive.Code)
	if resolveCalls != 1 {
		t.Fatalf("resolve call count = %d, want 1", resolveCalls)
	}
	if forwardCalls != 0 {
		t.Fatalf("forward call count = %d, want 0", forwardCalls)
	}
}

func TestPasswordPersisterPasswordMismatchFailsImmediately(t *testing.T) {
	self := meta.NewAgentInfo("127.0.0.1", 2881)
	cache := &tenantservice.PasswordMap{}
	resolveCalls := 0
	checkCalls := 0
	persister := passwordPersister{
		self:  self,
		cache: cache,
		resolve: func(context.Context) (meta.AgentInfoInterface, error) {
			resolveCalls++
			return self, nil
		},
		check: func(context.Context, string, string) (bool, error) {
			checkCalls++
			return false, nil
		},
		forward: func(context.Context, meta.AgentInfoInterface, string, string) error {
			t.Fatal("the local maintainer must not forward a password mismatch")
			return nil
		},
		delay: 0,
	}

	err := persister.persist(context.Background(), "tenant1", "wrong", true)
	assertAgentErrorCode(t, err, oberrors.ErrObTenantRootPasswordIncorrect.Code)
	if resolveCalls != 1 {
		t.Fatalf("resolve call count = %d, want 1", resolveCalls)
	}
	if checkCalls != 1 {
		t.Fatalf("check call count = %d, want 1", checkCalls)
	}
	if password, ok := cache.Get("tenant1"); ok {
		t.Fatalf("cache unexpectedly contains password %q", password)
	}
}

func TestPasswordPersisterCachesValidatedPassword(t *testing.T) {
	self := meta.NewAgentInfo("127.0.0.1", 2881)
	cache := &tenantservice.PasswordMap{}
	resolveCalls := 0
	checkCalls := 0
	persister := passwordPersister{
		self:  self,
		cache: cache,
		resolve: func(context.Context) (meta.AgentInfoInterface, error) {
			resolveCalls++
			return self, nil
		},
		check: func(_ context.Context, name, password string) (bool, error) {
			checkCalls++
			if name != "tenant1" || password != "password" {
				t.Fatalf("check got name=%q password=%q", name, password)
			}
			return true, nil
		},
		forward: func(context.Context, meta.AgentInfoInterface, string, string) error {
			t.Fatal("the current maintainer must not forward a successful local store")
			return nil
		},
		delay: 0,
	}

	if err := persister.persist(context.Background(), "tenant1", "password", true); err != nil {
		t.Fatalf("persist returned error: %v", err)
	}
	if resolveCalls != 3 {
		t.Fatalf("resolve call count = %d, want 3", resolveCalls)
	}
	if checkCalls != 1 {
		t.Fatalf("check call count = %d, want 1", checkCalls)
	}
	password, ok := cache.Get("tenant1")
	if !ok {
		t.Fatal("cache does not contain tenant1")
	}
	if password != "password" {
		t.Fatalf("cached password = %q, want %q", password, "password")
	}
}

func TestPasswordPersisterStopsWhenContextIsCanceled(t *testing.T) {
	self := meta.NewAgentInfo("127.0.0.1", 2881)
	resolveStarted := make(chan struct{})
	result := make(chan error, 1)
	var resolveCalls atomic.Int32
	retryErr := stderrors.New("retryable authority lookup error")
	persister := passwordPersister{
		self:  self,
		cache: &tenantservice.PasswordMap{},
		resolve: func(context.Context) (meta.AgentInfoInterface, error) {
			if resolveCalls.Add(1) == 1 {
				close(resolveStarted)
			}
			return nil, retryErr
		},
		check: func(context.Context, string, string) (bool, error) {
			t.Fatal("check must not run when authority resolution fails")
			return false, nil
		},
		forward: func(context.Context, meta.AgentInfoInterface, string, string) error {
			t.Fatal("forward must not run when authority resolution fails")
			return nil
		},
		delay: time.Hour,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		result <- persister.persist(ctx, "tenant1", "password", true)
	}()

	select {
	case <-resolveStarted:
		cancel()
	case <-time.After(time.Second):
		cancel()
		t.Fatal("persister did not start authority resolution")
	}

	select {
	case err := <-result:
		if !stderrors.Is(err, context.Canceled) {
			t.Fatalf("persist error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("persister goroutine did not exit after cancellation")
	}
	if calls := resolveCalls.Load(); calls != 1 {
		t.Fatalf("resolve call count = %d, want 1", calls)
	}
}

func TestPasswordPersisterReturnsLastErrorAfterRetryBudget(t *testing.T) {
	self := meta.NewAgentInfo("127.0.0.1", 2881)
	lastErr := stderrors.New("authority unavailable")
	resolveCalls := 0
	persister := passwordPersister{
		self:  self,
		cache: &tenantservice.PasswordMap{},
		resolve: func(context.Context) (meta.AgentInfoInterface, error) {
			resolveCalls++
			return nil, lastErr
		},
		check: func(context.Context, string, string) (bool, error) {
			t.Fatal("check must not run when authority resolution fails")
			return false, nil
		},
		forward: func(context.Context, meta.AgentInfoInterface, string, string) error {
			t.Fatal("forward must not run when authority resolution fails")
			return nil
		},
		delay: 0,
	}

	err := persister.persist(context.Background(), "tenant1", "password", true)
	if !stderrors.Is(err, lastErr) {
		t.Fatalf("persist error = %v, want %v", err, lastErr)
	}
	if resolveCalls != passwordPersistAttempts {
		t.Fatalf("resolve call count = %d, want %d", resolveCalls, passwordPersistAttempts)
	}
}

func TestPasswordPersisterRetriesNewOwnerWhenAuthorityChangesAfterPrecheck(t *testing.T) {
	self := meta.NewAgentInfo("127.0.0.1", 2881)
	newMaintainer := meta.NewAgentInfo("127.0.0.2", 2882)
	cache := &tenantservice.PasswordMap{}
	resolveCalls := 0
	forwardCalls := 0
	persister := passwordPersister{
		self:  self,
		cache: cache,
		resolve: func(context.Context) (meta.AgentInfoInterface, error) {
			resolveCalls++
			if resolveCalls == 1 {
				return self, nil
			}
			return newMaintainer, nil
		},
		check: func(context.Context, string, string) (bool, error) {
			return true, nil
		},
		forward: func(_ context.Context, maintainer meta.AgentInfoInterface, name, password string) error {
			forwardCalls++
			if !samePasswordAgent(maintainer, newMaintainer) {
				t.Fatalf("forward target = %s, want %s", maintainer.String(), newMaintainer.String())
			}
			if name != "tenant1" || password != "password" {
				t.Fatalf("forward got name=%q password=%q", name, password)
			}
			return nil
		},
		delay: 0,
	}

	if err := persister.persist(context.Background(), "tenant1", "password", true); err != nil {
		t.Fatalf("persist returned error: %v", err)
	}
	if resolveCalls != 3 {
		t.Fatalf("resolve call count = %d, want 3", resolveCalls)
	}
	if forwardCalls != 1 {
		t.Fatalf("forward call count = %d, want 1", forwardCalls)
	}
	if password, ok := cache.Get("tenant1"); ok {
		t.Fatalf("old maintainer cache unexpectedly contains password %q", password)
	}
}

func TestPasswordPersisterReplicatesToNewOwnerWhenAuthorityChangesAfterStore(t *testing.T) {
	self := meta.NewAgentInfo("127.0.0.1", 2881)
	newMaintainer := meta.NewAgentInfo("127.0.0.2", 2882)
	cache := &tenantservice.PasswordMap{}
	resolveCalls := 0
	forwardCalls := 0
	persister := passwordPersister{
		self:  self,
		cache: cache,
		resolve: func(context.Context) (meta.AgentInfoInterface, error) {
			resolveCalls++
			switch resolveCalls {
			case 1, 2:
				return self, nil
			default:
				return newMaintainer, nil
			}
		},
		check: func(context.Context, string, string) (bool, error) {
			return true, nil
		},
		forward: func(_ context.Context, maintainer meta.AgentInfoInterface, _, _ string) error {
			forwardCalls++
			if !samePasswordAgent(maintainer, newMaintainer) {
				t.Fatalf("forward target = %s, want %s", maintainer.String(), newMaintainer.String())
			}
			return nil
		},
		delay: 0,
	}

	if err := persister.persist(context.Background(), "tenant1", "password", true); err != nil {
		t.Fatalf("persist returned error: %v", err)
	}
	if resolveCalls != 4 {
		t.Fatalf("resolve call count = %d, want 4", resolveCalls)
	}
	if forwardCalls != 1 {
		t.Fatalf("forward call count = %d, want 1", forwardCalls)
	}
	password, ok := cache.Get("tenant1")
	if !ok {
		t.Fatal("old maintainer cache does not contain the validated password")
	}
	if password != "password" {
		t.Fatalf("cached password = %q, want %q", password, "password")
	}
}

func TestIsTenantPasswordIncorrectRecognizesTypedAPIError(t *testing.T) {
	err := &libhttp.ApiError{
		ErrCode: oberrors.ErrObTenantRootPasswordIncorrect.Code,
		Message: "incorrect password",
	}
	if !isTenantPasswordIncorrect(err) {
		t.Fatal("typed API password error was not recognized")
	}

	otherErr := &libhttp.ApiError{
		ErrCode: oberrors.ErrAgentMaintainerNotActive.Code,
		Message: "maintainer unavailable",
	}
	if isTenantPasswordIncorrect(otherErr) {
		t.Fatal("non-password API error was recognized as a password mismatch")
	}
}

func TestForwardedPasswordErrorSurvivesResponseSerialization(t *testing.T) {
	peerError := &passwordAPIError{
		ApiError: &libhttp.ApiError{ErrCode: oberrors.ErrObTenantRootPasswordIncorrect.Code, Message: "incorrect password"},
		status:   http.StatusBadRequest,
	}
	response := libhttp.BuildResponse(nil, peerError)
	if response.Successful || response.Status != http.StatusBadRequest || response.Error == nil {
		t.Fatalf("incorrect response: %+v", response)
	}
	if response.Error.ErrCode != oberrors.ErrObTenantRootPasswordIncorrect.Code || response.Error.Message != peerError.Message {
		t.Fatalf("forwarded error changed: %+v", response.Error)
	}
	if !isTenantPasswordIncorrect(peerError) {
		t.Fatal("forwarded mismatch must not be retried")
	}
	partial := oberrors.WrapOverride(oberrors.ErrObTenantPasswordCacheSyncFailed, peerError, "/api/v1/tenant/example/password/persist")
	if got := libhttp.BuildResponse(nil, partial); got.Error.ErrCode != oberrors.ErrObTenantPasswordCacheSyncFailed.Code {
		t.Fatalf("partial-success code lost: %+v", got.Error)
	}
}

func TestPreparedPasswordRequestAllowsExistingPrecheckRetryTime(t *testing.T) {
	var calls atomic.Int32
	// httptest owns the handler goroutine; Close waits for its completion.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		timer := time.NewTimer(3200 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-r.Context().Done():
			return
		case <-timer.C:
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"successful": true}); err != nil {
			t.Errorf("write precheck response: %v", err)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sendPreparedPasswordRequest(ctx, http.MethodGet, server.URL, nil, nil, nil); err != nil {
		t.Fatalf("precheck longer than three seconds must use the shared budget: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("precheck calls = %d, want one", calls.Load())
	}
}

func TestPreparedPasswordRequestHonorsCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// The server cancels the caller after receipt. Its handler exits when the
	// request context closes, and server.Close joins that handler.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cancel()
		<-r.Context().Done()
	}))
	defer server.Close()
	err := sendPreparedPasswordRequest(ctx, http.MethodGet, server.URL, nil, nil, nil)
	if err == nil || ctx.Err() != context.Canceled {
		t.Fatalf("canceled request returned %v, context error %v", err, ctx.Err())
	}
}

func TestSetRootPasswordTaskRecovery(t *testing.T) {
	mismatch := oberrors.Occur(oberrors.ErrObTenantRootPasswordIncorrect)
	leaseError := oberrors.Occur(oberrors.ErrAgentMaintainerNotActive)
	modifyError := stderrors.New("modify failed")
	for _, tc := range []struct {
		name        string
		executions  int
		continued   bool
		persistErr  error
		modifyErr   error
		wantPersist int
		wantModify  int
		wantError   bool
	}{
		{name: "first execution", executions: 1, wantModify: 1},
		{name: "retry after committed password", executions: 2, wantPersist: 1},
		{name: "continue after committed password", executions: 1, continued: true, wantPersist: 1},
		{name: "retry before password changed", executions: 2, persistErr: mismatch, wantPersist: 1, wantModify: 1},
		{name: "continue before password changed", executions: 1, continued: true, persistErr: mismatch, wantPersist: 1, wantModify: 1},
		{name: "retry with unavailable maintainer", executions: 2, persistErr: leaseError, wantPersist: 1, wantError: true},
		{name: "continue with unavailable maintainer", executions: 1, continued: true, persistErr: leaseError, wantPersist: 1, wantError: true},
		{name: "modify failure is propagated", executions: 1, modifyErr: modifyError, wantModify: 1, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pwdTask := &SetRootPwdTask{Task: *task.NewSubTask(TASK_NAME_SET_ROOT_PWD)}
			for i := 0; i < tc.executions; i++ {
				pwdTask.AddExecuteTimes()
			}
			if tc.continued {
				pwdTask.SetIsContinue()
			}
			persistCalls, modifyCalls := 0, 0
			err := pwdTask.executePasswordChange(func() error {
				persistCalls++
				return tc.persistErr
			}, func() error {
				modifyCalls++
				return tc.modifyErr
			})
			if (err != nil) != tc.wantError || persistCalls != tc.wantPersist || modifyCalls != tc.wantModify {
				t.Fatalf("error=%v, persist=%d, modify=%d; want error=%v, persist=%d, modify=%d", err, persistCalls, modifyCalls, tc.wantError, tc.wantPersist, tc.wantModify)
			}
		})
	}
}

func TestPasswordCompletionSurvivesClientDisconnect(t *testing.T) {
	type contextKey struct{}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "request-value"))
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, "/api/v1/tenant/example/password", nil)
	if err != nil {
		t.Fatal(err)
	}
	c := &gin.Context{Request: request}
	cancel()
	if passwordRequestContext(c).Err() != context.Canceled {
		t.Fatal("direct persistence must retain caller cancellation")
	}
	completion := passwordCompletionContext(c)
	if completion.Err() != nil || completion.Value(contextKey{}) != "request-value" {
		t.Fatal("committed password completion must retain values but not caller cancellation")
	}
	self := meta.NewAgentInfo("127.0.0.1", 2886)
	cache := &tenantservice.PasswordMap{}
	persister := passwordPersister{
		self: self, cache: cache,
		resolve: func(context.Context) (meta.AgentInfoInterface, error) { return self, nil },
		check: func(checkCtx context.Context, name, password string) (bool, error) {
			deadline, ok := checkCtx.Deadline()
			if !ok || time.Until(deadline) > passwordPersistTimeout {
				t.Error("completion must install its own bounded request budget")
			}
			return true, checkCtx.Err()
		},
	}
	if err := persister.persist(completion, "example", "new-password", true); err != nil {
		t.Fatalf("completion after client disconnect: %v", err)
	}
	if got, ok := cache.Get("example"); !ok || got != "new-password" {
		t.Fatal("committed password was not cached after client disconnect")
	}
}

func TestPasswordPersisterPassesCancellationToAuthorityLookup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	persister := passwordPersister{
		resolve: func(queryCtx context.Context) (meta.AgentInfoInterface, error) {
			cancel()
			<-queryCtx.Done()
			return nil, queryCtx.Err()
		},
	}
	if err := persister.persist(ctx, "example", "password", true); !stderrors.Is(err, context.Canceled) {
		t.Fatalf("authority lookup did not receive caller cancellation: %v", err)
	}
}

func assertAgentErrorCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error is nil, want code %s", want)
	}
	var agentErr oberrors.OcsAgentErrorInterface
	if !stderrors.As(err, &agentErr) {
		t.Fatalf("error type = %T, want OcsAgentErrorInterface: %v", err, err)
	}
	if got := agentErr.ErrorCode().Code; got != want {
		t.Fatalf("error code = %s, want %s", got, want)
	}
}
