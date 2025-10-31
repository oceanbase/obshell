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
	"io"

	pb "github.com/schollz/progressbar/v3"
)

const total = 100

type ProcessBar struct {
	pb  *pb.ProgressBar
	msg string
	Animate
}

type PbConfigForDag struct {
	Msg          string
	MaxStage     int
	CurrentStage int
	Writer       io.Writer

	width       int
	enableColor bool
	theme       *pb.Theme
}

func newPbConfig(stream io.Writer, msg string) *PbConfigForDag {
	conf := &PbConfigForDag{
		Msg:         msg,
		width:       30,
		enableColor: true,
		Writer:      stream,
		theme: &pb.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		},
	}
	return conf
}

func newProcessBar(io *IO, msg string) *ProcessBar {
	conf := newPbConfig(io.GetOutputStream(), msg)
	bar := pb.NewOptions(
		total,
		pb.OptionSetWriter(conf.Writer),
		pb.OptionEnableColorCodes(conf.enableColor),
		pb.OptionSetWidth(conf.width),
		pb.OptionSetDescription(msg),
		pb.OptionSetTheme(*conf.theme),
	)
	return &ProcessBar{
		pb:      bar,
		msg:     msg,
		Animate: *newAnimate(io),
	}
}

func (pb *ProcessBar) Start() error {
	pb.isRunning = true
	return pb.pb.Add(1)
}

func (pb *ProcessBar) Increment() error {
	return pb.pb.Add(1)
}

func (pb *ProcessBar) Set(i int) error {
	return pb.pb.Set(i)
}

func (pb *ProcessBar) Finish() error {
	pb.isRunning = false
	defer pb.stop()
	return pb.pb.Finish()
}

func (pb *ProcessBar) Exit() error {
	pb.isRunning = false
	defer pb.stop()
	return pb.pb.Exit()
}

func (pb *ProcessBar) stop() {
	pb.io.outputStream.Write([]byte("\n"))
	pb.isRunning = false
}

func (pb *ProcessBar) updateStream(stream io.Writer) {
	return
}
