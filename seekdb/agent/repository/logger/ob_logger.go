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

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

var OBDefault = logger.New(New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
	SlowThreshold:             200 * time.Millisecond,
	LogLevel:                  logger.Warn,
	IgnoreRecordNotFoundError: false,
	Colorful:                  true,
})

var DefaultExcludes = []string{
	"access_id",
	"access_key",
}

type OBLogger struct {
	log.Logger
	excludes []string
}

func (o *OBLogger) Printf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	for _, exclude := range o.excludes {
		if strings.Contains(str, exclude) {
			return
		}
	}
	o.Logger.Printf(str)
}

func New(out io.Writer, prefix string, flag int) *OBLogger {
	l := new(OBLogger)
	l.SetOutput(out)
	l.SetPrefix(prefix)
	l.SetFlags(flag)
	l.excludes = DefaultExcludes
	return l
}
