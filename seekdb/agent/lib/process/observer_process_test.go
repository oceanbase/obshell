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

package process

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
)

func TestObserverProcessBelongsToBaseDir(t *testing.T) {
	for _, scenario := range []string{"base", "symlink", "other", "run", "missing", "zero", "negative", "overflow"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			base := filepath.Join(root, "base")
			if err := os.MkdirAll(filepath.Join(base, "run"), 0755); err != nil {
				t.Fatal(err)
			}
			cwd := base
			baseArg := base
			switch scenario {
			case "symlink":
				baseArg = filepath.Join(root, "alias")
				if err := os.Symlink(base, baseArg); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			case "other":
				cwd = root
			case "run":
				cwd = filepath.Join(base, "run")
			}

			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(executable, "-test.run=^TestObserverProcessBaseDirHelper$")
			cmd.Dir = cwd
			cmd.Env = append(os.Environ(), "OBSHELL_PROCESS_TEST_BASE="+baseArg, "OBSHELL_PROCESS_TEST_SCENARIO="+scenario)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("process identity check failed: %v\n%s", err, output)
			}
		})
	}
}

func TestObserverProcessBaseDirHelper(t *testing.T) {
	base := os.Getenv("OBSHELL_PROCESS_TEST_BASE")
	if base == "" {
		return
	}
	// Resolve path state only in the child so other package tests retain their arguments and base directory.
	os.Args = []string{os.Args[0], "agent", "start", "--base-dir", base}
	scenario := os.Getenv("OBSHELL_PROCESS_TEST_SCENARIO")
	pid := strconv.Itoa(os.Getpid())
	switch scenario {
	case "zero":
		pid = "0"
	case "negative":
		pid = "-1"
	case "overflow":
		pid = strconv.FormatInt(int64(os.Getpid())+1<<32, 10)
	}
	if scenario != "missing" {
		if err := os.WriteFile(filepath.Join(path.RunDir(), "seekdb.pid"), []byte(pid), 0600); err != nil {
			t.Fatal(err)
		}
	}
	want := scenario == "base" || scenario == "symlink"
	got, err := CheckObserverProcess()
	if scenario == "overflow" {
		if got || err == nil {
			t.Fatalf("CheckObserverProcess() = %v, %v for an overflowing PID", got, err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("CheckObserverProcess() = %v, want %v for %s", got, want, scenario)
	}
	gotPID, err := GetObserverPid()
	if err != nil {
		t.Fatal(err)
	}
	if want && gotPID != pid {
		t.Fatalf("GetObserverPid() = %q, want %q", gotPID, pid)
	}
	if !want && gotPID != "" {
		t.Fatalf("GetObserverPid() = %q for %s", gotPID, scenario)
	}
}

func TestCheckProcessExistRejectsNonPositivePID(t *testing.T) {
	for _, pid := range []int32{0, -1} {
		got, err := CheckProcessExist(pid)
		if err != nil {
			t.Fatal(err)
		}
		if got {
			t.Errorf("CheckProcessExist(%d) = true", pid)
		}
	}
}
