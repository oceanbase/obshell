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

import "fmt"

func NewIO() *IO {
	return std.NewSubIO()
}

func Verbose(msg string) {
	std.Verbose(msg)
}

func Verbosef(format string, a ...any) {
	std.Verbosef(format, a...)
}

func Print(msg string) {
	std.Print(msg)
}

func Printf(format string, a ...any) {
	std.Printf(format, a...)
}

func PrintfWithoutNewline(format string, a ...any) {
	std.PrintfWithoutNewline(format, a...)
}

func Success(msg string) {
	std.Success(msg)
}

func Successf(format string, a ...interface{}) {
	std.Successf(format, a...)
}

func Info(msg string) {
	std.Info(msg)
}

func Infof(format string, a ...any) {
	std.Infof(format, a...)
}

func Warn(msg string) {
	std.Warn(msg)
}

func Warnf(format string, a ...any) {
	std.Warnf(format, a...)
}

func Error(msg string) {
	std.Error(msg)
}

func Errorf(format string, a ...any) {
	std.Errorf(format, a...)
}

func Failed(msg string) {
	std.Failed(msg)
}

func Failedf(format string, a ...interface{}) {
	std.Failedf(format, a...)
}

func Confirm(msg string) (bool, error) {
	return std.Confirm(msg)
}

func Confirmf(format string, a ...any) (bool, error) {
	return std.Confirmf(format, a...)
}

func StartLoading(msg string) {
	std.StartLoading(msg)
}

func StartLoadingf(format string, a ...any) {
	std.StartLoading(fmt.Sprintf(format, a...))
}

func UpdateLoading(msg string) {
	std.UpdateLoading(msg)
}

func UpdateLoadingf(format string, a ...any) {
	std.UpdateLoading(fmt.Sprintf(format, a...))
}

func StartOrUpdateLoading(msg string) {
	if std.spinner == nil {
		StartLoading(msg)
		return
	}
	UpdateLoading(msg)
}

func StopLoading() {
	std.StopLoading()
}

func LoadInfo(msg string) {
	std.LoadInfo(msg)
}

func LoadInfof(format string, a ...any) {
	std.LoadInfof(format, a...)
}

func LoadSuccess(msg string) {
	std.LoadSuccess(msg)
}

func LoadSuccessf(format string, a ...any) {
	std.LoadSuccessf(format, a...)
}

func LoadError(msg string) {
	std.LoadError(msg)
}

func LoadErrorf(format string, a ...any) {
	std.LoadErrorf(format, a...)
}

func LoadFailed(msg string) {
	std.LoadFailed(msg)
}

func LoadFailedf(format string, a ...any) {
	std.LoadFailedf(format, a...)
}

func LoadStageSuccess(msg string) {
	std.LoadStageSuccess(msg)
}

func LoadStageSuccessf(format string, a ...interface{}) {
	std.LoadStageSuccessf(format, a...)
}

func StartProcessBar(msg string) {
	std.StartProcessBar(msg)
}

func UpdateProcessBar(i int) {
	std.UpdateProcessBar(i)
}

func IncProcessBar() {
	std.IncProcessBar()
}

func FinishProcessBar() {
	std.FinishProcessBar()
}

func ExitProcessBar() {
	std.ExitProcessBar()
}

func IsBusy() bool {
	return std.IsBusy()
}
