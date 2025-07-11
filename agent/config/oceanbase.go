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
	"fmt"
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
)

func NewObDataSourceConfig() *ObDataSourceConfig {
	return &ObDataSourceConfig{
		username:        constant.DB_USERNAME,
		ip:              meta.OCS_AGENT.GetLocalIp(),
		dBName:          constant.DB_OCS,
		charset:         constant.DB_DEFAULT_CHARSET,
		parseTime:       true,
		location:        constant.DB_DEFAULT_LOCATION,
		maxIdleConns:    constant.DB_DEFAULT_MAX_IDLE_CONNS,
		maxOpenConns:    constant.DB_DEFAULT_MAX_OPEN_CONNS,
		connMaxLifetime: constant.DB_DEFAULT_CONN_MAX_LIFETIME,
	}
}

func NewObproxyDataSourceConfig() *ObDataSourceConfig {
	return &ObDataSourceConfig{
		username:        constant.DB_PROXYSYS_USERNAME,
		ip:              constant.LOCAL_IP,
		charset:         constant.DB_DEFAULT_CHARSET,
		parseTime:       true,
		location:        constant.DB_DEFAULT_LOCATION,
		maxIdleConns:    constant.DB_DEFAULT_MAX_IDLE_CONNS,
		maxOpenConns:    constant.DB_DEFAULT_MAX_OPEN_CONNS,
		connMaxLifetime: constant.DB_DEFAULT_CONN_MAX_LIFETIME,
	}
}

type ObDataSourceConfig struct {
	// dsn config
	username          string
	password          string
	ip                string
	port              int
	dBName            string
	charset           string
	parseTime         bool
	location          string
	interpolateParams bool

	// pool config
	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime int

	// connection config
	timeout      int  // second
	tryTimes     int  // second, if tryTime <= 0, try forever
	skipPwdCheck bool // skip password check

	// logger config
	LoggerConfig
}

func (config *ObDataSourceConfig) SetUsername(username string) *ObDataSourceConfig {
	config.username = username
	return config
}

func (config *ObDataSourceConfig) SetPassword(password string) *ObDataSourceConfig {
	config.password = password
	return config
}

func (config *ObDataSourceConfig) SetIp(ip string) *ObDataSourceConfig {
	config.ip = ip
	return config
}

func (config *ObDataSourceConfig) SetPort(port int) *ObDataSourceConfig {
	config.port = port
	return config
}

func (config *ObDataSourceConfig) SetDBName(dBName string) *ObDataSourceConfig {
	config.dBName = dBName
	return config
}

func (config *ObDataSourceConfig) SetTimeout(timeout int) *ObDataSourceConfig {
	config.timeout = timeout
	return config
}

func (config *ObDataSourceConfig) SetTryTimes(tryTimes int) *ObDataSourceConfig {
	config.tryTimes = tryTimes
	return config
}

func (config *ObDataSourceConfig) SetSkipPwdCheck(skip bool) *ObDataSourceConfig {
	config.skipPwdCheck = skip
	return config
}

func (config *ObDataSourceConfig) SetCharset(charset string) *ObDataSourceConfig {
	config.charset = charset
	return config
}

func (config *ObDataSourceConfig) SetParseTime(parseTime bool) *ObDataSourceConfig {
	config.parseTime = parseTime
	return config
}

func (config *ObDataSourceConfig) SetLocation(location string) *ObDataSourceConfig {
	config.location = location
	return config
}

func (config *ObDataSourceConfig) SetInterpolateParams(interpolateParams bool) *ObDataSourceConfig {
	config.interpolateParams = interpolateParams
	return config
}

func (config *ObDataSourceConfig) SetMaxIdleConns(maxIdleConns int) *ObDataSourceConfig {
	config.maxIdleConns = maxIdleConns
	return config
}

func (config *ObDataSourceConfig) SetMaxOpenConns(maxOpenConns int) *ObDataSourceConfig {
	config.maxOpenConns = maxOpenConns
	return config
}

func (config *ObDataSourceConfig) SetConnMaxLifetime(connMaxLifetime int) *ObDataSourceConfig {
	config.connMaxLifetime = connMaxLifetime
	return config
}

func (config *ObDataSourceConfig) GetUsername() string {
	return config.username
}

func (config *ObDataSourceConfig) GetPassword() string {
	return config.password
}

func (config *ObDataSourceConfig) GetIp() string {
	return config.ip
}

func (config *ObDataSourceConfig) GetPort() int {
	return config.port
}

func (config *ObDataSourceConfig) GetDBName() string {
	return config.dBName
}

func (config *ObDataSourceConfig) GetTryTimes() int {
	if config.tryTimes <= 0 {
		return -1
	}
	return config.tryTimes
}

func (config *ObDataSourceConfig) GetMaxIdleConns() int {
	return config.maxIdleConns
}

func (config *ObDataSourceConfig) GetMaxOpenConns() int {
	return config.maxOpenConns
}

func (config *ObDataSourceConfig) GetConnMaxLifetime() int {
	return config.connMaxLifetime
}

func (config *ObDataSourceConfig) GetSkipPwdCheck() bool {
	return config.skipPwdCheck
}

func (config *ObDataSourceConfig) GetDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", config.username, config.password, meta.NewAgentInfo(config.ip, config.port).String())
	if config.dBName != "" {
		dsn += config.dBName
	}

	params := make([]string, 0)
	if config.timeout > 0 {
		params = append(params, fmt.Sprintf("timeout=%ds", config.timeout))
	}
	if config.charset != "" {
		params = append(params, fmt.Sprintf("charset=%s", config.charset))
	}
	if config.location != "" {
		params = append(params, fmt.Sprintf("loc=%s", config.location))
	}
	if config.interpolateParams {
		params = append(params, "interpolateParams=true")
	}
	if config.parseTime {
		params = append(params, "parseTime=true")
	}
	params = append(params, "ob_query_timeout=60000000")
	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}
	return dsn
}
