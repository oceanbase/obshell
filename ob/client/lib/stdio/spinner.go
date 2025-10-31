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
	"time"

	"github.com/briandowns/spinner"
)

type Spinner struct {
	spinner *spinner.Spinner
	Animate
}

func createSpinner(steam io.Writer) *spinner.Spinner {
	return spinner.New(
		spinner.CharSets[9],
		100*time.Millisecond,
		spinner.WithWriter(steam),
	)
}

func newSpinner(io *IO) *Spinner {
	return &Spinner{
		spinner: createSpinner(io.GetOutputStream()),
		Animate: *newAnimate(io),
	}
}

func (s *Spinner) Start(prefix string) {
	s.spinner.Prefix = prefix + " "
	if s.isTTY {
		s.spinner.Start()
	}
	s.isRunning = true
	s.termHandler(s.spinner.Prefix + "...")
}

func (s *Spinner) Update(prefix string) {
	s.spinner.Prefix = prefix + " "
	s.termHandler(s.spinner.Prefix)
}

func (s *Spinner) Stop() {
	s.spinner.FinalMSG = s.spinner.Prefix + "\n"
	s.termHandler(s.spinner.FinalMSG)
	s.spinner.Stop()
	s.isRunning = false
}

func (s *Spinner) stopAndPersist(symbol *FormattedText, text string) {
	s.spinner.FinalMSG = fmt.Sprintf("%s %s\n", symbol.Format(s.isTTY), text)
	s.termHandler(s.spinner.FinalMSG)
	s.spinner.Stop()
	s.isRunning = false
}

func (s *Spinner) LoadStageSuccess(text string) {
	s.spinner.FinalMSG = fmt.Sprintf("%s %s\n", SuccessSymbol.Format(s.isTTY), text)
	s.termHandler(s.spinner.FinalMSG)
	s.spinner.Stop()
	s.spinner.Start()
}

func (s *Spinner) updateStream(stream io.Writer) {
	newSpinner := createSpinner(stream)
	newSpinner.Prefix = s.spinner.Prefix
	s.isTTY = true
	s.spinner.Stop()
	s.spinner = newSpinner
	s.spinner.Start()
}
