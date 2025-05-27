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
	"errors"
	"fmt"
	"net"
)

type AgentMode = string

const (
	DebugMode   AgentMode = "debug"
	ReleaseMode AgentMode = "release"
)

type ServerConfig struct {
	Ip          string
	Port        int
	Address     string
	RunDir      string
	UpgradeMode bool
}

func NewServerConfig(ip string, port int, runDir string, UpgradeMode bool) (*ServerConfig, error) {
	address, err := generateAddress(ip, port)
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		Ip:          ip,
		Port:        port,
		Address:     address,
		RunDir:      runDir,
		UpgradeMode: UpgradeMode,
	}, nil
}

func generateAddress(ip string, port int) (string, error) {
	ipParsed := net.ParseIP(ip)
	if ipParsed == nil {
		return "", errors.New("invalid ip")
	}
	if ipParsed.To4() != nil {
		return fmt.Sprint("0.0.0.0:", port), nil
	}
	return fmt.Sprint("[::]:", port), nil
}
