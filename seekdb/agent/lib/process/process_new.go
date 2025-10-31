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
	"sync"
	"syscall"
	"time"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
)

type Process struct {
	cmd          *exec.Cmd
	conf         ProcessConfig
	mu           sync.Mutex
	running      bool
	state        ProcState
	stderrBuffer *switchableBuffer
	stdoutBuffer *switchableBuffer
}

type ProcessConfig struct {
	Program     string
	Args        []string
	LogFilePath string
}

type ProcState struct {
	Pid      int
	StartAt  time.Time
	EndAt    time.Time
	Exited   bool
	ExitCode int
}

func NewProcess(conf ProcessConfig) *Process {
	return &Process{
		conf: conf,
	}
}

func newBuffers(path string) (stdout *switchableBuffer, stderr *switchableBuffer, err error) {
	if path == "" {
		return nil, nil, errors.Occur(errors.ErrCommonInvalidPath, "", "path is empty")
	}
	parentDir := filepath.Dir(path)
	if err = os.MkdirAll(parentDir, 0755); err != nil {
		return
	}
	fileBuffer, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE|syscall.O_CLOEXEC, 0644)
	if err != nil {
		return nil, nil, err
	}

	stdout, err = newSwitchableBuffer(fileBuffer, os.Stdout)
	if err != nil {
		return nil, nil, err
	}
	stderr, err = newSwitchableBuffer(fileBuffer, os.Stderr)
	if err != nil {
		return nil, nil, err
	}

	stdout.memModel = false
	stderr.memModel = true
	return stdout, stderr, nil
}

func newCmd(conf ProcessConfig) (cmd *exec.Cmd) {
	cmd = exec.Command(conf.Program, conf.Args...)
	return cmd
}

func (p *Process) Start() (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return errors.Occur(errors.ErrCommonUnexpected, "proc already running")
	}

	p.cmd = newCmd(p.conf)
	p.stdoutBuffer, p.stderrBuffer, err = newBuffers(p.conf.LogFilePath)
	if err != nil {
		return err
	}
	p.cmd.Stdout = p.stdoutBuffer
	p.cmd.Stderr = p.stderrBuffer
	if err := p.cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start proc")
	}

	p.switchToRunningState()
	go p.wait()
	return nil
}

func (p *Process) wait() {
	p.cmd.Wait()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.switchToExitedState()
}

func (p *Process) switchToExitedState() {
	endAt := time.Now()
	p.running = false
	prevState := p.state
	state := p.cmd.ProcessState
	p.state = ProcState{
		Pid:      prevState.Pid,
		Exited:   true,
		ExitCode: state.ExitCode(),
		StartAt:  prevState.StartAt,
		EndAt:    endAt,
	}
	p.cmd = nil
}

func (p *Process) switchToRunningState() {
	p.running = true
	p.state = ProcState{
		Pid:     p.cmd.Process.Pid,
		StartAt: time.Now(),
		Exited:  false,
	}
}

// GetState will return the current process state.
func (p *Process) GetState() ProcState {
	return p.state
}

func (p *Process) Stop() error {
	return p.signal(syscall.SIGTERM)
}

// Kill will send a KILL signal.
func (p *Process) Kill() error {
	return p.signal(syscall.SIGKILL)
}

func (p *Process) signal(s os.Signal) error {
	if p.cmd == nil {
		return errors.Occur(errors.ErrCommonUnexpected, "proc not exist")
	}
	process := p.cmd.Process
	if process == nil {
		return errors.Occur(errors.ErrCommonUnexpected, "proc not exist")
	}
	if err := process.Signal(s); err != nil {
		return err
	}
	return nil
}

func (p *Process) SwitchToLogMode() {
	if p.stderrBuffer != nil {
		p.stderrBuffer.memModel = false
		p.stderrBuffer.Flush()
	}
}

func (p *Process) IsRunning() bool {
	return p.running
}
