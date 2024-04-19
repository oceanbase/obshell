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
	"errors"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

type IOer interface {
	Verbose(msg string)
	Verbosef(format string, a ...any)
	Print(msg string)
	Printf(format string, a ...any)
	Info(msg string)
	Infof(format string, a ...any)
	Warn(msg string)
	Warnf(format string, a ...any)
	Error(msg string)
	Errorf(format string, a ...any)
	Confirm(msg string) (bool, error)
	Confirmf(format string, a ...any) (bool, error)
	StartLoading(message string) *IO
	UpdateLoading(message string)
	StopLoading()
	StartProcessBar(message string)
	UpdateProcessBar(message string)
	IncProcessBar()
	StopProcessBar()
}

const (
	DEBUG = log.DebugLevel
	NORM  = log.Level(10)
	INFO  = log.InfoLevel
	WARN  = log.WarnLevel
	ERROR = log.ErrorLevel
	FATAL = log.FatalLevel
)

var std *IO

type IO struct {
	inputStream   io.Reader
	outputStream  io.Writer
	currOutStream io.Writer
	inputIsTTY    bool
	outputIsTTY   bool
	verboseMode   bool
	silenceMode   bool
	skipConfirm   bool
	logger        *log.Logger

	rootIO    *IO
	animater  Animater
	spinner   *Spinner
	bar       *ProcessBar
	waitQueue []Animater
}

func init() {
	std = newIO(os.Stdin, os.Stdout, log.StandardLogger())
	std.logger = nil
}

func newIO(input io.Reader, output io.Writer, logger *log.Logger) *IO {
	return &IO{
		inputStream:   input,
		outputStream:  output,
		currOutStream: output,
		outputIsTTY:   term.IsTerminal(int(output.(*os.File).Fd())),
		inputIsTTY:    term.IsTerminal(int(input.(*os.File).Fd())),
		logger:        logger,
	}
}

func (io *IO) GetCurrentStream() io.Writer {
	if io.rootIO != nil {
		return io.rootIO.GetCurrentStream()
	}
	return io.currOutStream
}

func (io *IO) GetOutputStream() io.Writer {
	if io.rootIO != nil {
		return io.rootIO.GetCurrentStream()
	}
	return io.outputStream
}

func (io *IO) log(level log.Level, msg string) {
	if io.logger == nil || io.silenceMode {
		return
	}
	io.logger.Log(level, msg)
}

func (io *IO) print(level log.Level, msg string, end ...string) {
	if io.rootIO != nil {
		io.rootIO.print(level, msg, end...)
		return
	}

	var symbol *FormattedText
	print := true
	switch level {
	case DEBUG:
		msg = "- " + msg
		print = io.verboseMode
	case INFO:
		symbol = InfoSymbol
	case NORM:
		level = INFO
	case WARN:
		symbol = WarningSymbol
	case ERROR:
		symbol = ErrorSymbol
	case FATAL:
		symbol = FailedSymbol
	}

	endf := "\n"
	if len(end) > 0 {
		endf = end[0]
	}

	if print {
		if symbol != nil {
			fmt.Fprint(io.currOutStream, symbol.Format(io.outputIsTTY)+" "+msg+endf)
		} else {
			fmt.Fprint(io.currOutStream, msg+endf)
		}
	}
	io.log(level, msg)
}

func (io *IO) Verbose(msg string) {
	io.print(DEBUG, msg)
}

func (io *IO) Verbosef(format string, a ...any) {
	io.Verbose(fmt.Sprintf(format, a...))
}

func (io *IO) Print(msg string) {
	io.print(NORM, msg)
}

func (io *IO) Printf(format string, a ...any) {
	io.Print(fmt.Sprintf(format, a...))
}

func (io *IO) Success(msg string) {
	std.Print(fmt.Sprintf("%s %s", SuccessSymbol.Format(std.outputIsTTY), msg))
}

func (io *IO) Successf(format string, a ...interface{}) {
	std.Success(fmt.Sprintf(format, a...))
}

func (io *IO) Info(msg string) {
	io.print(INFO, msg)
}

func (io *IO) Infof(format string, a ...any) {
	io.Info(fmt.Sprintf(format, a...))
}

func (io *IO) Warn(msg string) {
	io.print(WARN, msg)
}

func (io *IO) Warnf(format string, a ...any) {
	io.Warn(fmt.Sprintf(format, a...))
}

func (io *IO) Error(msg string) {
	io.print(ERROR, msg)
}

func (io *IO) Errorf(format string, a ...any) {
	io.Error(fmt.Sprintf(format, a...))
}

func (io *IO) Failed(msg string) {
	io.print(FATAL, msg)
}

func (io *IO) Failedf(format string, a ...any) {
	io.Failed(fmt.Sprintf(format, a...))
}

func (io *IO) Confirm(msg string) (bool, error) {
	if io.IsBusy() {
		return false, errors.New("stdio is busy")
	}

	io.print(NORM, msg+" [Y/N]: ", "")
	if io.skipConfirm {
		io.Print("Y (auto confirm)")
		return true, nil
	}
	if !io.inputIsTTY {
		io.Print("N (auto confirm)")
		return false, nil
	}

	for {
		var input string
		_, err := fmt.Fscanln(io.inputStream, &input)
		if err != nil {
			return false, err
		}
		switch input {
		case "Y", "y":
			return true, nil
		case "N", "n":
			return false, nil
		default:
			io.print(NORM, "Please input Y or N: ", "")
		}
	}
}

func (io *IO) Confirmf(format string, a ...any) (bool, error) {
	return io.Confirm(fmt.Sprintf(format, a...))
}

func (io *IO) NewSubIO() *IO {
	return &IO{
		rootIO: io,
	}
}
