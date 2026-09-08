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

package standby

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/seekdb/param"
)

type fakeSwitchoverToPrimaryService struct {
	statuses     []param.LocalStandbyStatus
	statusIndex  int
	clearErr     error
	flipErr      error
	operations   []string
	flippedHost  string
	flippedPort  int
	flippedDir   string
	upstreamPeer sqlite.SeekdbStandbyPeer
}

func (f *fakeSwitchoverToPrimaryService) GetLocalStatus() (param.LocalStandbyStatus, error) {
	f.operations = append(f.operations, "status")
	if len(f.statuses) == 0 {
		return param.LocalStandbyStatus{}, errors.New("no fake status configured")
	}
	index := f.statusIndex
	if index >= len(f.statuses) {
		index = len(f.statuses) - 1
	}
	f.statusIndex++
	return f.statuses[index], nil
}

func (f *fakeSwitchoverToPrimaryService) SwitchoverToPrimary() error {
	f.operations = append(f.operations, "promote")
	return nil
}

func (f *fakeSwitchoverToPrimaryService) ClearLogRestoreSource() error {
	f.operations = append(f.operations, "clear")
	return f.clearErr
}

func (f *fakeSwitchoverToPrimaryService) GetUpstreamPeerForCaller(callerPort int) (*sqlite.SeekdbStandbyPeer, error) {
	f.operations = append(f.operations, "resolve")
	if f.upstreamPeer.PeerObshellPort == 0 {
		f.upstreamPeer = sqlite.SeekdbStandbyPeer{PeerHost: "10.0.0.1", PeerObshellPort: callerPort}
	}
	return &f.upstreamPeer, nil
}

func (f *fakeSwitchoverToPrimaryService) FlipDirection(host string, port int, direction string) error {
	f.operations = append(f.operations, "flip")
	f.flippedHost = host
	f.flippedPort = port
	f.flippedDir = direction
	return f.flipErr
}

func stableStatus(role string) param.LocalStandbyStatus {
	return param.LocalStandbyStatus{
		Role:             role,
		PendingRole:      "INVALID",
		SwitchoverStatus: "NORMAL",
	}
}

func TestExecuteSwitchoverToPrimaryClearsRestoreSourceBeforeFlippingPeer(t *testing.T) {
	service := &fakeSwitchoverToPrimaryService{
		statuses: []param.LocalStandbyStatus{stableStatus("STANDBY"), stableStatus("PRIMARY")},
	}
	p := param.RpcSwitchoverToPrimaryParam{CallerHost: "10.0.0.1", CallerObshellPort: 2886}

	if err := executeSwitchoverToPrimary(service, p); err != nil {
		t.Fatalf("execute switchover to primary: %v", err)
	}
	wantOperations := []string{"status", "promote", "status", "clear", "status", "resolve", "flip"}
	if !reflect.DeepEqual(service.operations, wantOperations) {
		t.Fatalf("unexpected operation order: got %v, want %v", service.operations, wantOperations)
	}
	if service.flippedHost != p.CallerHost || service.flippedPort != p.CallerObshellPort || service.flippedDir != constant.STANDBY_DIRECTION_DOWNSTREAM {
		t.Fatalf("unexpected peer update: %s:%d %s", service.flippedHost, service.flippedPort, service.flippedDir)
	}
}

func TestExecuteSwitchoverToPrimaryRetriesCleanupWithoutPromotingAgain(t *testing.T) {
	service := &fakeSwitchoverToPrimaryService{
		statuses: []param.LocalStandbyStatus{stableStatus("PRIMARY")},
	}
	p := param.RpcSwitchoverToPrimaryParam{CallerHost: "10.0.0.1", CallerObshellPort: 2886}

	if err := executeSwitchoverToPrimary(service, p); err != nil {
		t.Fatalf("retry switchover cleanup: %v", err)
	}
	wantOperations := []string{"status", "clear", "status", "resolve", "flip"}
	if !reflect.DeepEqual(service.operations, wantOperations) {
		t.Fatalf("unexpected retry operations: got %v, want %v", service.operations, wantOperations)
	}
}

func TestExecuteSwitchoverToPrimaryDoesNotPublishPeerWhenClearFails(t *testing.T) {
	service := &fakeSwitchoverToPrimaryService{
		statuses: []param.LocalStandbyStatus{stableStatus("PRIMARY")},
		clearErr: errors.New("clear failed"),
	}

	err := executeSwitchoverToPrimary(service, param.RpcSwitchoverToPrimaryParam{})
	if err == nil || !strings.Contains(err.Error(), "failed to clear log restore source") {
		t.Fatalf("expected clear failure, got %v", err)
	}
	wantOperations := []string{"status", "clear"}
	if !reflect.DeepEqual(service.operations, wantOperations) {
		t.Fatalf("unexpected failure operations: got %v, want %v", service.operations, wantOperations)
	}
}

func TestExecuteSwitchoverToPrimaryDoesNotPublishPeerWhenClearIsNotVisible(t *testing.T) {
	notCleared := stableStatus("PRIMARY")
	notCleared.LogRestoreSource = "10.0.0.1:2882"
	service := &fakeSwitchoverToPrimaryService{
		statuses: []param.LocalStandbyStatus{stableStatus("PRIMARY"), notCleared},
	}

	err := executeSwitchoverToPrimary(service, param.RpcSwitchoverToPrimaryParam{CallerObshellPort: 2886})
	if err == nil || !strings.Contains(err.Error(), "after clearing it") {
		t.Fatalf("expected readback failure, got %v", err)
	}
	wantOperations := []string{"status", "clear", "status"}
	if !reflect.DeepEqual(service.operations, wantOperations) {
		t.Fatalf("unexpected readback failure operations: got %v, want %v", service.operations, wantOperations)
	}
}

func TestExecuteSwitchoverToPrimaryUsesCanonicalUpstreamAddress(t *testing.T) {
	service := &fakeSwitchoverToPrimaryService{
		statuses:     []param.LocalStandbyStatus{stableStatus("PRIMARY")},
		upstreamPeer: sqlite.SeekdbStandbyPeer{PeerHost: "canonical-primary", PeerObshellPort: 2886},
	}
	p := param.RpcSwitchoverToPrimaryParam{CallerHost: "primary-alias", CallerObshellPort: 2886}

	if err := executeSwitchoverToPrimary(service, p); err != nil {
		t.Fatalf("execute switchover with caller alias: %v", err)
	}
	if service.flippedHost != "canonical-primary" || service.flippedPort != 2886 {
		t.Fatalf("unexpected canonical peer update: %s:%d", service.flippedHost, service.flippedPort)
	}
}

func TestExecuteSwitchoverToPrimaryRejectsUnstableRole(t *testing.T) {
	service := &fakeSwitchoverToPrimaryService{
		statuses: []param.LocalStandbyStatus{{
			Role:             "PRIMARY",
			PendingRole:      "STANDBY",
			SwitchoverStatus: "PREPARING",
		}},
	}

	err := executeSwitchoverToPrimary(service, param.RpcSwitchoverToPrimaryParam{})
	if err == nil || !strings.Contains(err.Error(), "did not reach a stable primary state") {
		t.Fatalf("expected unstable-state failure, got %v", err)
	}
	wantOperations := []string{"status"}
	if !reflect.DeepEqual(service.operations, wantOperations) {
		t.Fatalf("unexpected unstable-state operations: got %v, want %v", service.operations, wantOperations)
	}
}
