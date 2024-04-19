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

package config

import (
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/log"
)

const (
	LEVEL       = "info"
	MAX_SIZE    = 100
	MAX_AGE     = 30
	MAX_BACKUPS = 10
)

func DefaultAgentLoggerConifg() log.LoggerConfig {
	conf := defaultLoggerConfig()
	conf.Filename = path.ObshellLogPath()
	return *conf
}

func DefaultDaemonLoggerConifg() log.LoggerConfig {
	conf := defaultLoggerConfig()
	conf.Filename = path.DaemonLogPath()
	return *conf
}

func DefaultClientLoggerConifg() log.LoggerConfig {
	conf := defaultLoggerConfig()
	conf.Filename = path.ClientLogPath()
	return *conf
}

func defaultLoggerConfig() *log.LoggerConfig {
	return &log.LoggerConfig{
		Level:      LEVEL,
		MaxSize:    MAX_SIZE,
		MaxAge:     MAX_AGE,
		MaxBackups: MAX_BACKUPS,
		LocalTime:  true,
		Compress:   false,
	}
}
