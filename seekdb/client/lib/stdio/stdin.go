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

package stdio

import (
	"bufio"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var input *SysStdin

type SysStdin struct {
	isTTY    bool
	reader   *bufio.Reader
	nonBlock bool
}

func init() {
	input = NewSysStdin()
}

func ReadLine(blocked bool) (string, error) {
	return input.ReadLine(blocked)
}

func ReadLines(blocked bool) ([]string, error) {
	return input.ReadLines(blocked)
}

func IsTTY() bool {
	return input.IsTTY()
}

func NewSysStdin() *SysStdin {
	return &SysStdin{
		isTTY:    isTerminal(os.Stdin.Fd()),
		reader:   bufio.NewReader(os.Stdin),
		nonBlock: false,
	}
}

// IsTTY checks if Stdin is a terminal.
func (s *SysStdin) IsTTY() bool {
	return s.isTTY
}

// ReadLine reads a line from Stdin, potentially in a non-blocking manner.
func (s *SysStdin) ReadLine(blocked bool) (string, error) {
	if !blocked {
		// Set up a channel to receive input without blocking.
		inputChan := make(chan string)
		go func() {
			line, _, _ := s.reader.ReadLine()
			inputChan <- string(line)
		}()

		select {
		case line := <-inputChan:
			return line, nil
		default:
			// If no input is ready, return an empty string.
			return "", nil
		}
	}

	// For blocking mode, simply use ReadLine.
	res, err := s.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res), nil
}

// ReadLines reads all available lines, potentially in a non-blocking manner.
func (s *SysStdin) ReadLines(blocked bool) ([]string, error) {
	var lines []string
	for {
		line, err := s.ReadLine(blocked)
		if err != nil {
			break
		}
		if line == "" {
			break
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// isTerminal returns true if the given file descriptor is a terminal.
func isTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&termios)))
	return err == 0
}
