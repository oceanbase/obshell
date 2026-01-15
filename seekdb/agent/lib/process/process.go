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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
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
	obServerPidPath := filepath.Join(path.RunDir(), "seekdb.pid")
	pid, err := getPid(obServerPidPath)
	if err != nil {
		return nil, err
	}
	return &ProcessInfo{
		pid:      fmt.Sprint(pid),
		procPath: fmt.Sprintf("/proc/%d", pid), // kept for Linux compatibility, not used on macOS
	}, nil
}

func (p *ProcessInfo) exist() (bool, error) {
	if p.pid == "" {
		return false, nil
	}
	pidInt, err := strconv.Atoi(p.pid)
	if err != nil {
		return false, err
	}
	// On macOS and other non-Linux platforms, use Signal(0) to check if process exists
	if runtime.GOOS != "linux" {
		proc, err := os.FindProcess(pidInt)
		if err != nil {
			return false, nil
		}
		// Signal(0) doesn't actually send a signal, it just checks if the process exists
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			return false, nil
		}
		return true, nil
	}
	// On Linux, use /proc for efficiency
	if _, err := os.Stat(p.procPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *ProcessInfo) checkDir() (bool, error) {
	pidInt, err := strconv.Atoi(p.pid)
	if err != nil {
		return false, err
	}

	var cwdPath string
	if runtime.GOOS == "linux" {
		// On Linux, use /proc
		cwdPath = fmt.Sprintf("/proc/%s/cwd", p.pid)
	} else {
		// On macOS and other platforms, use gopsutil to get process working directory
		proc, err := process.NewProcess(int32(pidInt))
		if err != nil {
			return false, err
		}
		cwd, err := proc.Cwd()
		if err != nil {
			return false, err
		}
		cwdPath = cwd
	}
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

func GetObserverProcess() (*ProcessInfo, error) {
	return getObserverProcess()
}

func GetObserverPid() (string, error) {
	process, err := getObserverProcess()
	if err != nil {
		return "", err
	}
	return process.Pid()
}

func GetObserverPidInt() (int32, error) {
	return getPid(filepath.Join(path.RunDir(), "seekdb.pid"))
}

func GetObserverBinPath() (string, error) {
	pid, err := GetObserverPid()
	if err != nil {
		return "", err
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "linux" {
		// On Linux, use /proc
		return os.Readlink(fmt.Sprintf("/proc/%s/exe", pid))
	}
	// On macOS and other platforms, use gopsutil to get process executable path
	proc, err := process.NewProcess(int32(pidInt))
	if err != nil {
		return "", err
	}
	exe, err := proc.Exe()
	if err != nil {
		return "", err
	}
	return exe, nil
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

func GetSystemdUnit(pid int) (string, error) {
	// systemd is Linux-specific, return empty on other platforms
	if runtime.GOOS != "linux" {
		return "", nil
	}

	ctx := context.Background()
	if conn, err := dbus.NewSystemdConnectionContext(ctx); err == nil && conn != nil {
		defer conn.Close()
		unitName, _ := conn.GetUnitNameByPID(ctx, uint32(pid))
		// ignore error
		if unitName != "" {
			return unitName, nil
		}
	}

	// get unit name from /proc/pid/cgroup
	cgroupPath := fmt.Sprintf("/proc/%d/cgroup", pid)
	file, err := os.Open(cgroupPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner != nil {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "seekdb.service") {
				return "seekdb.service", nil
			}
		}
	}

	return "", nil
}
