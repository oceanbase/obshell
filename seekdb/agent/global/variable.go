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

package global

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"runtime"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/lib/process"
)

var (
	HomePath         string
	Uid              uint32
	Gid              uint32
	Pid              = os.Getpid()
	StartAt          = time.Now().UnixNano()
	Protocol         = "http"
	CaCertPool       *x509.CertPool
	SkipVerify       bool
	EnableHTTPS      bool
	EnableClientAuth bool // require and verify client certificates (mTLS) when true
	// OutboundMTLSKeyPair is loaded in init() when EnableClientAuth is true; used by HTTP clients.
	OutboundMTLSKeyPair tls.Certificate
	EnableTelemetry  bool
	Architecture     string
	Os               string
)

var (
	architectureMap = map[string]string{
		"amd64":   "x86_64",
		"x86_64":  "x86_64",
		"arm64":   "aarch64",
		"aarch64": "aarch64",
	}
)

func initArchitecture() {
	arch := runtime.GOARCH
	if _, ok := architectureMap[arch]; !ok {
		Architecture = arch
	} else {
		Architecture = architectureMap[arch]
	}
	log.Info("architecture is ", Architecture)
}

func initTelemetry() {
	EnableTelemetry = true // default enable telemetry
	envValue := os.Getenv(constant.ENV_OBSHELL_TELEMETRY_ENABLED)
	if envValue != "" {
		// if env value is not empty, only enable telemetry when env value is 1
		EnableTelemetry = envValue == "1"
	}
}

func InitGlobalVariable() {
	HomePath = path.AgentDir()
	Uid = process.Uid()
	Gid = process.Gid()
	log.Info("homePath is ", HomePath)
	log.Info("binDir is ", path.BinDir())
	initArchitecture()
	initTelemetry()
	Os = runtime.GOOS
}

func init() {
	keyFile, certFile := path.ObshellCertificateAndKeyPaths()
	if keyFile == "" || certFile == "" {
		return
	}
	pool := x509.NewCertPool()
	for _, cert := range path.ObshellCertificatePaths() {
		caCert, err := os.ReadFile(cert)
		if err != nil {
			log.WithError(err).Warn("read ca cert file failed")
			return
		}
		pool.AppendCertsFromPEM(caCert)
	}
	CaCertPool = pool
	Protocol = "https"
	EnableHTTPS = true
	_, SkipVerify = syscall.Getenv(constant.SKIP_VERIFY)
	_, EnableClientAuth = syscall.Getenv(constant.ENABLE_CLIENT_AUTH)
	if EnableClientAuth {
		pair, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.WithError(err).Fatal("enable_client_auth requires a valid TLS key/cert pair for outbound mTLS")
		}
		OutboundMTLSKeyPair = pair
	}
}
