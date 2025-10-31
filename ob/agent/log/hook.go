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

package log

import (
	"errors"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type CallerHook struct{}

func (hook *CallerHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 4)
	cnt := runtime.Callers(8, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !isIgnorePackages(name) {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data[logrus.FieldKeyFile] = getPackage(file)
			entry.Data[FieldKeyLine] = line
			entry.Data[logrus.FieldKeyFunc] = name[strings.LastIndex(name, ".")+1:]
			break
		}
	}
	return nil
}

func (hook *CallerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func isIgnorePackages(name string) bool {
	return strings.Contains(name, "github.com/sirupsen/logrus") ||
		strings.Contains(name, "github.com/go-kit/log") ||
		strings.Contains(name, "github.com/gin-gonic/gin")
}

type CostDurationHook struct{}

func (hook *CostDurationHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}
	startTime := entry.Context.Value(StartTimeKey)
	if startTime == nil {
		return nil
	}
	start, ok := startTime.(time.Time)
	if !ok {
		return errors.New("startTime is no time.Time")
	}

	duration := time.Now().Sub(start)
	entry.Data[FieldKeyDuration] = duration
	return nil
}

func (hook *CostDurationHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
