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
	"fmt"
	"io"
	"reflect"
)

var (
	TypeSpinner    = reflect.TypeOf(&Spinner{})
	TYpeProcessBar = reflect.TypeOf(&ProcessBar{})
)

type Animater interface {
	IsRunning() bool
	updateStream(io.Writer)
}

type Animate struct {
	io        *IO
	isTTY     bool
	isRunning bool
}

func newAnimate(io *IO) *Animate {
	return &Animate{
		io:    io,
		isTTY: io.outputIsTTY,
	}
}

func (s *Animate) termHandler(message string) {
	s.io.log(NORM, message)
}

func (s *Animate) IsRunning() bool {
	return s.isRunning
}

func (io *IO) IsBusy() bool {
	return io.animater != nil && io.animater.IsRunning()
}

func (io *IO) startAnimate(animater Animater) bool {
	if io.IsBusy() {
		io.waitQueue = append(io.waitQueue, animater)
		return true
	}
	if !io.setAnimate(animater) {
		return false
	}

	if io.rootIO != nil {
		if !io.rootIO.startAnimate(animater) {
			io.unsetAnimate()
			return false
		}
	} else {
		io.currOutStream = NewBufferIO(true)
	}
	return true
}

func (io *IO) setAnimate(animater Animater) bool {
	switch reflect.TypeOf(animater) {
	case TypeSpinner:
		io.spinner = animater.(*Spinner)
	case TYpeProcessBar:
		io.bar = animater.(*ProcessBar)
	default:
		return false
	}
	io.animater = animater
	return true
}

func (io *IO) stopAnimate(animater Animater) bool {
	if io.animater != animater {
		return false
	}
	if io.rootIO != nil {
		if !io.rootIO.stopAnimate(animater) {
			return false
		}
	} else {
		io.unsetAnimate()
		if io.handleWaitQueue() {
			buffer := io.currOutStream.(*BufferIO).String()
			if buffer != "" {
				fmt.Fprintln(io.outputStream, buffer)
			}
			io.currOutStream = io.outputStream
		}
	}
	return true
}

func (io *IO) unsetAnimate() {
	io.animater = nil
	io.spinner = nil
}

// if has wait queue, handle it and return false, else return true
func (io *IO) handleWaitQueue() bool {
	if len(io.waitQueue) == 0 {
		return true
	}

	for i, animater := range io.waitQueue {
		if !animater.IsRunning() || !io.setAnimate(animater) {
			continue
		}
		io.waitQueue = io.waitQueue[i+1:]
		animater.updateStream(io.outputStream)
		return false
	}
	return true
}

func (io *IO) StartLoading(message string) {
	if io.IsBusy() {
		return
	}

	spinner := newSpinner(io)
	if !io.startAnimate(spinner) {
		return
	}

	io.spinner = spinner
	io.spinner.Start(message)
}

func (io *IO) StartLoadingf(format string, a ...interface{}) {
	io.StartLoading(fmt.Sprintf(format, a...))
}

func (io *IO) UpdateLoading(message string) {
	if !io.IsBusy() || io.spinner == nil {
		return
	}
	io.spinner.Update(message)
}

func (io *IO) UpdateLoadingf(format string, a ...interface{}) {
	io.UpdateLoading(fmt.Sprintf(format, a...))
}

func (io *IO) stopLoading(symbol *FormattedText, text string) {
	if !io.IsBusy() || io.spinner == nil {
		return
	}
	if symbol != nil {
		io.spinner.stopAndPersist(symbol, text)
	} else {
		io.spinner.Stop()
	}
	io.stopAnimate(io.animater)
}

func (io *IO) StopLoading() {
	io.stopLoading(nil, "")
}

func (io *IO) LoadSuccess(message string) {
	io.stopLoading(SuccessSymbol, message)
}

func (io *IO) LoadSuccessf(format string, a ...interface{}) {
	io.LoadSuccess(fmt.Sprintf(format, a...))
}

func (io *IO) LoadStageSuccess(message string) {
	if !io.IsBusy() || io.spinner == nil {
		return
	}
	if io.rootIO != nil {
		if io.rootIO.animater != io.animater && io.rootIO.spinner != nil {
			io.rootIO.LoadStageSuccess(message)
			return
		}
	}
	io.spinner.LoadStageSuccess(message)
}

func (io *IO) LoadStageSuccessf(format string, a ...interface{}) {
	io.LoadStageSuccess(fmt.Sprintf(format, a...))
}

func (io *IO) LoadFailed(message string) {
	io.stopLoading(FailedSymbol, message)
}

func (io *IO) LoadFailedf(format string, a ...interface{}) {
	io.LoadFailed(fmt.Sprintf(format, a...))
}

func (io *IO) LoadError(message string) {
	io.stopLoading(ErrorSymbol, message)
}

func (io *IO) LoadErrorf(format string, a ...interface{}) {
	io.LoadError(fmt.Sprintf(format, a...))
}

func (io *IO) LoadWarning(message string) {
	io.stopLoading(WarningSymbol, message)
}

func (io *IO) LoadWarningf(format string, a ...interface{}) {
	io.LoadWarning(fmt.Sprintf(format, a...))
}

func (io *IO) LoadInfo(message string) {
	io.stopLoading(InfoSymbol, message)
}

func (io *IO) LoadInfof(format string, a ...interface{}) {
	io.LoadInfo(fmt.Sprintf(format, a...))
}

func (io *IO) StartProcessBar(message string) {
	if io.IsBusy() {
		return
	}

	bar := newProcessBar(io, message)
	if !io.startAnimate(bar) {
		return
	}

	io.bar = bar
	io.bar.Start()
}

func (io *IO) IncProcessBar() {
	if !io.IsBusy() || io.bar == nil {
		return
	}
	io.bar.Increment()
}

func (io *IO) UpdateProcessBar(i int) {
	if !io.IsBusy() || io.bar == nil {
		return
	}
	io.bar.Set(i)
}

func (io *IO) FinishProcessBar() {
	if !io.IsBusy() || io.bar == nil {
		return
	}
	io.bar.Finish()
}

func (io *IO) ExitProcessBar() {
	if !io.IsBusy() || io.bar == nil {
		return
	}
	io.bar.Exit()
}

func PrintTable(header []string, data [][]string) {
	std.PrintTable(header, data)
}

func PrintTableWithTitle(title string, header []string, data [][]string) {
	std.PrintTableWithTitle(title, header, data)
}
