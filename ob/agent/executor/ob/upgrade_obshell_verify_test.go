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

package ob

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/meta"
)

func TestRestoreLegacyObshellRpmIdentity(t *testing.T) {
	const (
		version      = "4.5.0.1"
		release      = "12026081800.el8"
		architecture = "aarch64"
	)
	originalArchitecture := global.Architecture
	global.Architecture = architecture
	t.Cleanup(func() {
		global.Architecture = originalArchitecture
	})

	ctx := task.NewTaskContext().
		SetParam(PARAM_VERSION, version).
		SetParam(PARAM_RELEASE_DISTRIBUTION, release)
	legacy := rpmPacakgeInstallInfo{RpmName: constant.PKG_OBSHELL}

	identity, err := restoreObshellRpmIdentity(ctx, &legacy)
	if err != nil {
		t.Fatalf("restore legacy OBShell RPM identity: %v", err)
	}
	if identity.Name != constant.PKG_OBSHELL || identity.Version != version || identity.Release != release || identity.Architecture != architecture {
		t.Fatalf("unexpected restored identity: %+v", identity)
	}
	if legacy.RpmVersion != version || legacy.RpmRelease != release || legacy.RpmArchitecture != architecture {
		t.Fatalf("legacy task metadata was not restored: %+v", legacy)
	}
}

func TestRestoreObshellRpmIdentityRejectsPersistedMismatch(t *testing.T) {
	originalArchitecture := global.Architecture
	global.Architecture = "x86_64"
	t.Cleanup(func() {
		global.Architecture = originalArchitecture
	})

	ctx := task.NewTaskContext().
		SetParam(PARAM_VERSION, "4.5.0.1").
		SetParam(PARAM_RELEASE_DISTRIBUTION, "12026081800.el7")
	info := rpmPacakgeInstallInfo{
		RpmName:    constant.PKG_OBSHELL,
		RpmVersion: "4.5.0.2",
	}

	if _, err := restoreObshellRpmIdentity(ctx, &info); err == nil {
		t.Fatal("expected conflicting persisted RPM identity to be rejected")
	}
}

func TestInstallNewAgentGetParamsRestoresLegacyTask(t *testing.T) {
	sourcePath := os.Getenv("OBSHELL_SIGNED_RPM_TEST_PATH")
	if sourcePath == "" {
		t.Skip("OBSHELL_SIGNED_RPM_TEST_PATH is not set")
	}

	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read OBShell release RPM: %v", err)
	}
	rpmPath := filepath.Join(t.TempDir(), "obshell.rpm")
	if err := os.WriteFile(rpmPath, contents, 0600); err != nil {
		t.Fatalf("copy OBShell release RPM: %v", err)
	}

	input, err := os.Open(rpmPath)
	if err != nil {
		t.Fatalf("open OBShell release RPM: %v", err)
	}
	rpmPackage, err := pkg.ReadRpm(input)
	if err != nil {
		if closeErr := input.Close(); closeErr != nil {
			t.Errorf("close OBShell RPM after read failure: %v", closeErr)
		}
		t.Fatalf("read OBShell RPM identity: %v", err)
	}
	if err := input.Close(); err != nil {
		t.Fatalf("close OBShell release RPM: %v", err)
	}

	originalArchitecture := global.Architecture
	global.Architecture = rpmPackage.Architecture()
	t.Cleanup(func() {
		global.Architecture = originalArchitecture
	})
	ctx := task.NewTaskContext().
		SetParam(PARAM_VERSION, rpmPackage.Version()).
		SetParam(PARAM_RELEASE_DISTRIBUTION, rpmPackage.Release()).
		SetParam(PARAM_ONLY_FOR_AGENT, true)
	buildNumber, _, err := pkg.SplitRelease(rpmPackage.Release())
	if err != nil {
		t.Fatalf("split OBShell release: %v", err)
	}
	buildVersion := rpmPackage.Version() + "-" + buildNumber
	ctx.SetParam(PARAM_AGENT_UPGRADE_ROUTE, []RouteNode{{
		Version:      rpmPackage.Version(),
		Release:      buildNumber,
		BuildVersion: buildVersion,
	}})
	legacy := rpmPacakgeInstallInfo{
		RpmName:    constant.PKG_OBSHELL,
		RpmPkgPath: rpmPath,
	}
	execAgent := meta.AgentInfo{Ip: "127.0.0.1", Port: 2886}
	ctx.SetAgentData(&execAgent, buildVersion, legacy)
	installTask := newInstallNewAgentTask()
	installTask.SetContext(ctx)
	installTask.realExecAgent = execAgent
	installTask.SetLogChannel(make(chan task.TaskExecuteLogDTO, 1))

	if err := installTask.getParams(); err != nil {
		t.Fatalf("restore legacy OBShell task through InstallNewAgent: %v", err)
	}
	if installTask.rpmPkgInfo.ObshellSHA256 == "" {
		t.Fatal("legacy OBShell task did not receive the signed binary digest")
	}
	var persisted rpmPacakgeInstallInfo
	if err := ctx.GetAgentDataByAgentKeyWithValue(execAgent.String(), buildVersion, &persisted); err != nil {
		t.Fatalf("read restored OBShell task metadata: %v", err)
	}
	if persisted.RpmVersion != rpmPackage.Version() || persisted.RpmRelease != rpmPackage.Release() || persisted.RpmArchitecture != rpmPackage.Architecture() {
		t.Fatalf("restored RPM identity was not persisted: %+v", persisted)
	}
	if persisted.ObshellSHA256 != installTask.rpmPkgInfo.ObshellSHA256 {
		t.Fatalf("persisted digest %q does not match task digest %q", persisted.ObshellSHA256, installTask.rpmPkgInfo.ObshellSHA256)
	}
	extractedBinary := filepath.Join(filepath.Dir(rpmPath), "home", "admin", "oceanbase", "bin", "obshell")
	if _, err := os.Stat(extractedBinary); err != nil {
		t.Fatalf("stat OBShell binary extracted during legacy recovery: %v", err)
	}
}
