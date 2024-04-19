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
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/oceanbase/obshell/utils"
)

const (
	StartTimeKey = "startTime"
)

const defaultTimestampFormat = "2006-01-02T15:04:05.000"

var textFormatter = &TextFormatter{
	TimestampFormat:        defaultTimestampFormat,
	FullTimestamp:          true,
	DisableLevelTruncation: true,
	PadLevelText:           true,
	FieldMap: map[string]string{
		"WARNING": "WARN", // Log level string, use WARN
	},
	// Log caller, filename:line callFunction.
	CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
		filename := getPackage(frame.File)
		name := frame.Function
		idx := strings.LastIndex(name, ".")
		return name[idx+1:], fmt.Sprintf("%s:%d", filename, frame.Line)
	},
}

// noErrWriter wraps a writer to ignore error to avoid bad write cause logrus logger always print error messages.
type noErrWriter struct {
	o sync.Once
	w io.WriteCloser
}

func (w *noErrWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if err != nil {
		w.o.Do(func() {
			// Only print error message once
			_, _ = fmt.Fprintf(os.Stderr, "write log failed %v\n", err)
		})
		return len(p), nil
	}
	return
}

func (w *noErrWriter) Close() error {
	return w.w.Close()
}

type LoggerConfig struct {
	Level      string `yaml:"level"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxsize"`
	MaxAge     int    `yaml:"maxage"`
	MaxBackups int    `yaml:"maxbackups"`
	LocalTime  bool   `yaml:"localtime"`
	Compress   bool   `yaml:"compress"`
}

func InitLogger(config LoggerConfig) *logrus.Logger {
	logger := logrus.StandardLogger()
	if curOut, ok := logger.Out.(*noErrWriter); ok {
		if l, ok := curOut.w.(*lumberjack.Logger); ok {
			l.Filename = config.Filename
			l.MaxSize = config.MaxSize
			l.MaxBackups = config.MaxBackups
			l.MaxAge = config.MaxAge
			l.Compress = config.Compress
			_ = l.Close()
		}
	} else {
		writer := utils.NewRotateFile(config.Filename, int64(config.MaxSize), config.MaxAge, config.MaxBackups)
		logger.SetOutput(&noErrWriter{w: writer})
		logger.SetFormatter(textFormatter)
		logger.SetReportCaller(false)
		logger.AddHook(new(CostDurationHook))
		logger.AddHook(new(CallerHook))
	}

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		panic(fmt.Sprintf("parse log level: %+v", err))
	}
	logger.SetLevel(level)
	return logger
}

const (
	// Silent is silent log level
	Silent = iota + 1
	// Error is error log level
	Error
	// Warn is warn log level
	Warn
	// Info is info log level
	Info
)

var (
	db_log_level int = Error
)

func SetDBLoggerLevel(logLevel int) {
	db_log_level = logLevel
}

func GetDBLoggerLevel() int {
	return db_log_level
}
