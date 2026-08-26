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
	"path/filepath"
	"testing"
	"time"
)

func TestProcessWaitForExit(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	proc := NewProcess(ProcessConfig{
		Program:     executable,
		Args:        []string{"-test.run=TestProcessWaitForExitHelper", "--", "process-wait-helper"},
		LogFilePath: filepath.Join(t.TempDir(), "process.log"),
	})
	if err := proc.Start(); err != nil {
		t.Fatal(err)
	}
	if proc.WaitForExit(50 * time.Millisecond) {
		t.Fatal("process exited before receiving a signal")
	}
	if err := proc.Stop(); err != nil {
		t.Fatal(err)
	}
	if !proc.WaitForExit(5 * time.Second) {
		proc.Kill()
		t.Fatal("process did not exit after receiving SIGTERM")
	}
	if proc.IsRunning() {
		t.Fatal("process is still marked as running after exit")
	}
}

func TestProcessWaitForExitHelper(t *testing.T) {
	if len(os.Args) == 0 || os.Args[len(os.Args)-1] != "process-wait-helper" {
		return
	}
	select {}
}
