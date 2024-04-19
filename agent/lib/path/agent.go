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

package path

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/oceanbase/obshell/agent/constant"
)

// The following functions provide the full paths to various log files located in the log directory.
// This includes the daemon log, client log, obshell log, and obshell standard output log.

func DaemonLogPath() string {
	return filepath.Join(LogDir(), constant.PROC_OBSHELL_DAEMON+".log")
}

func ClientLogPath() string {
	return filepath.Join(LogDir(), constant.PROC_OBSHELL_CLIENT+".log")
}

func ObshellLogPath() string {
	return filepath.Join(LogDir(), constant.PROC_OBSHELL+".log")
}

func ObshellStdPath() string {
	return filepath.Join(LogDir(), constant.PROC_OBSHELL+".out.log")
}

func DaemonBinPath() string {
	return ObshellBinPath()
}

func DaemonSocketPath() string {
	return filepath.Join(RunDir(), constant.PROC_OBSHELL_DAEMON+".sock")
}

func DaemonSocketBakPath() string {
	return filepath.Join(RunDir(), constant.PROC_OBSHELL_DAEMON+".sock.bak")
}

func DaemonSocketTmpPath() string {
	return filepath.Join(RunDir(), constant.PROC_OBSHELL_DAEMON+".sock.tmp")
}

func DaemonPidPath() string {
	return filepath.Join(RunDir(), constant.PROC_OBSHELL_DAEMON+".pid")
}

func DaemonPidBakPath() string {
	return filepath.Join(RunDir(), constant.PROC_OBSHELL_DAEMON+".pid.bak")
}

func ObshellBinPath() string {
	return filepath.Join(BinDir(), constant.PROC_OBSHELL)
}

func ObshellBinBackupPath() string {
	return filepath.Join(EtcDir(), constant.PROC_OBSHELL)
}

func ObshellPidPath() string {
	return filepath.Join(RunDir(), constant.PROC_OBSHELL+".pid")
}

func ObshellPidBakPath() string {
	return filepath.Join(RunDir(), constant.PROC_OBSHELL+".pid.bak")
}

func ObshellMetaPath() string {
	return filepath.Join(AgentDir(), constant.OCSAGENT_META_PATH)
}

func ObshellSocketPath() string {
	return filepath.Join(RunDir(), fmt.Sprintf("%s.sock", constant.PROC_OBSHELL))
}

func ObshellSocketBakPath() string {
	return filepath.Join(RunDir(), fmt.Sprintf("%s.sock.bak", constant.PROC_OBSHELL))
}

func ObshellTmpSocketPath() string {
	return filepath.Join(RunDir(), fmt.Sprintf("%s.sock.tmp", constant.PROC_OBSHELL))
}

func ObshellCertificateAndKeyPaths() (key string, cert string) {
	keys := scanFiles(filepath.Join(CertificateDir(), "*.key"))
	for _, key := range keys {
		name := filepath.Base(key)
		nameWithoutExt := name[:len(name)-len(filepath.Ext(name))]
		cert = filepath.Join(CertificateDir(), nameWithoutExt+".crt")
		if _, err := os.Stat(cert); err == nil {
			return key, cert
		}
	}

	maps := make(map[string]string)
	files := scanFiles(filepath.Join(CertificateDir(), "*.pem"))
	for _, file := range files {
		name := filepath.Base(file)
		nameWithoutExt := name[:len(name)-len(filepath.Ext(name))]
		maps[nameWithoutExt] = file
	}

	for name, path := range maps {
		keyNmae := name + ".key"
		if _, ok := maps[keyNmae]; ok {
			return maps[keyNmae], path
		}
		keyNmae = name + "-key"
		if _, ok := maps[keyNmae]; ok {
			return maps[keyNmae], path
		}
	}
	return "", ""
}

func ObshellCertificatePaths() (certs []string) {
	certs = scanFiles(filepath.Join(CertificateDir(), "*.crt"))
	pem := scanFiles(filepath.Join(CertificateDir(), "*.pem"))
	certs = append(certs, pem...)
	return
}

func scanFiles(regex string) []string {
	files, err := filepath.Glob(regex)
	if err != nil {
		return []string{}
	}
	return files
}
