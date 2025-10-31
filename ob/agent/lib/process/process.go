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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/net"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
)

var (
	obServerPidPath = filepath.Join(path.RunDir(), "observer.pid")
)

func getPidStr(pidPath string) (string, error) {
	if _, err := os.Stat(pidPath); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	content, err := os.ReadFile(pidPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func getPid(pidPath string) (int32, error) {
	if _, err := os.Stat(pidPath); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	content, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(content)))
	return int32(pid), err
}

func checkProcessDir(cwdPath string) (bool, error) {
	workDir, err := filepath.EvalSymlinks(cwdPath)
	if err != nil {
		return false, err
	}
	agentDir, err := filepath.EvalSymlinks(path.RunDir())
	if err != nil {
		return false, err
	}

	absOB := filepath.Clean(workDir)
	absAgent := filepath.Clean(agentDir)
	return absOB != absAgent, nil
}

type ProcessInfo struct {
	pid      string
	procPath string
}

func getObserverProcess() (*ProcessInfo, error) {
	pid, err := getPid(obServerPidPath)
	if err != nil {
		return nil, err
	}
	return &ProcessInfo{
		pid:      fmt.Sprint(pid),
		procPath: fmt.Sprintf("/proc/%d", pid),
	}, nil
}

func getObproxyProcess() (*ProcessInfo, error) {
	pid, err := getPid(path.ObproxyPidPath())
	if err != nil {
		return nil, err
	}
	return &ProcessInfo{
		pid:      fmt.Sprint(pid),
		procPath: fmt.Sprintf("/proc/%d", pid),
	}, nil
}

func getObproxydProcess() (*ProcessInfo, error) {
	pid, err := getPid(path.ObproxydPidPath())
	if err != nil {
		return nil, err
	}
	return &ProcessInfo{
		pid:      fmt.Sprint(pid),
		procPath: fmt.Sprintf("/proc/%d", pid),
	}, nil
}

func (p *ProcessInfo) exist() (bool, error) {
	if p.pid == "" {
		return false, nil
	}
	if _, err := os.Stat(p.procPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CheckPortSocketInode(port string) (bool, error) {
	log.Infof("check port '%s' socket inode", port)
	value, err := strconv.Atoi(port)
	if err != nil {
		return false, err
	}
	portHex := strings.ToUpper(fmt.Sprintf("%04x", value))
	cmd := fmt.Sprintf("cat /proc/net/{tcp*,udp*}| awk -F' ' '{print $2,$10}' | grep '00000000:%s' | awk -F' ' '{print $2}' | uniq", portHex)
	res, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		return false, err
	}
	if len(res) > 0 {
		return true, nil
	}
	return false, nil
}

func (p *ProcessInfo) checkDir() (bool, error) {
	cwdPath := fmt.Sprintf("/proc/%s/cwd", p.pid)
	return checkProcessDir(cwdPath)
}

func (p *ProcessInfo) Exist() (bool, error) {
	if exist, err := p.exist(); err != nil || !exist {
		return false, err
	}
	return p.checkDir()
}

func (p *ProcessInfo) Pid() (string, error) {
	if exist, err := p.Exist(); err != nil || !exist {
		return "", err
	}
	return p.pid, nil
}

func CheckObserverProcess() (bool, error) {
	process, err := getObserverProcess()
	if err != nil {
		return false, err
	}
	return process.Exist()
}

func CheckObproxyProcess() (bool, error) {
	process, err := getObproxyProcess()
	if err != nil {
		return false, err
	}
	return process.Exist()
}

func CheckObproxydProcess() (bool, error) {
	process, err := getObproxydProcess()
	if err != nil {
		return false, err
	}
	return process.Exist()
}

func CheckProcessExist(pid int32) (bool, error) {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		log.Infof("process %d not exist", pid)
		return false, err
	}
	if err = proc.Signal(syscall.Signal(0)); err != nil {
		log.Printf("Process %d is dead!", pid)
		return false, nil
	} else {
		return true, nil
	}
}

func GetObserverPid() (string, error) {
	process, err := getObserverProcess()
	if err != nil {
		return "", err
	}
	return process.Pid()
}

func GetDaemonPid() (int32, error) {
	return getPid(path.DaemonPidPath())
}

func GetObshellPid() (int32, error) {
	return getPid(path.ObshellPidPath())
}

func GetObshellPidStr() (string, error) {
	return getPidStr(path.ObshellPidPath())
}

func GetPid(path string) (int32, error) {
	return getPid(path)
}

func ExecuteBinary(binaryPath string, inputs []string) (err error) {
	cmd := exec.Command(binaryPath, inputs...)

	// Get the standard output stream of the command.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	// Start a goroutine to process the output stream in real time.
	go func() {
		defer stdout.Close()

		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("read stdout error: %v\n", err)
				break
			}
			if err == io.EOF || n == 0 {
				break
			}

			fmt.Print(string(buf[:n]))
		}
	}()

	// Wait for command execution to complete.
	return cmd.Wait()
}

// for obproxy
func GetObproxyPid() (string, error) {
	process, err := getObproxyProcess()
	if err != nil {
		return "", err
	}
	return process.Pid()
}

// for obproxy
func GetObproxydPid() (string, error) {
	process, err := getObproxydProcess()
	if err != nil {
		return "", err
	}
	return process.Pid()
}

// writePid writes the pid to the specified path atomically.
// If the file already exists, an error is returned.
func WritePid(path string, pid int) (err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC|syscall.O_CLOEXEC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprint(f, pid)
	if err != nil {
		return err
	}
	return nil
}

// writePid writes the pid to the specified path atomically.
func WritePidForce(path string, pid int) (err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_SYNC|syscall.O_CLOEXEC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprint(f, pid)
	if err != nil {
		return err
	}
	return nil
}

func FindPIDByPort(port uint32) (int32, error) {
	// NOTICE: use inet6 to support ipv6
	connections, err := net.Connections("inet")
	if err != nil {
		return 0, err
	}

	for _, conn := range connections {
		if conn.Laddr.Port == port {
			return conn.Pid, nil
		}
	}
	return 0, errors.Occurf(errors.ErrCommonUnexpected, "no process found on port %d", port)
}
