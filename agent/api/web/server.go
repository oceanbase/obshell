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

package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	api "github.com/oceanbase/obshell/agent/api"
	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/global"
	http2 "github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/rpc"
)

type Server struct {
	Config config.ServerConfig
	state  *http2.State

	LocalServer
	TcpServer
}

type LocalServer struct {
	LocalRouter     *gin.Engine
	LocalHttpServer *http.Server
	UnixListener    *net.UnixListener
}

type TcpServer struct {
	Router      *gin.Engine
	HttpServer  *http.Server
	TcpListener *net.TCPListener
}

// NewServer initializes gin mode, register api and rpc routers for
// engine instances, and returns a new server instance.
func NewServer(mode config.AgentMode, conf config.ServerConfig) *Server {
	router := gin.New()
	localRouter := gin.New()

	if mode == config.DebugMode {
		gin.SetMode(gin.DebugMode)
		router.Use(gin.Logger())
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ret := &Server{
		TcpServer: TcpServer{
			Router: router,
		},
		LocalServer: LocalServer{
			LocalRouter: localRouter,
		},
		Config: conf,
		state:  http2.NewState(constant.STATE_STARTING),
	}
	api.InitOcsAgentRoutes(ret.state, router, false)
	api.InitOcsAgentRoutes(ret.state, localRouter, true)
	rpc.InitOcsAgentRpcRoutes(ret.state, router, false)
	rpc.InitOcsAgentRpcRoutes(ret.state, localRouter, true)
	return ret
}

// NewServerOnlyLocal initializes gin mode, register api and rpc routers for
// local router engine instances, and returns a new server instance.
func NewServerOnlyLocal(mode config.AgentMode, conf config.ServerConfig) *Server {
	if mode == config.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	localRouter := gin.New()
	ret := &Server{
		LocalServer: LocalServer{
			LocalRouter: localRouter,
		},
		Config: conf,
		state:  http2.NewState(constant.STATE_STARTING),
	}
	api.InitOcsAgentRoutes(ret.state, localRouter, true)
	rpc.InitOcsAgentRpcRoutes(ret.state, localRouter, true)
	return ret
}

// NewUnxiListener creates a new unix socket listener,
// and returns the listener and error.
func (s *Server) NewUnixListener() (*net.UnixListener, error) {
	s.LocalHttpServer = &http.Server{
		Handler:      s.LocalRouter,
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, common.UNIX_CONNECT, c)
		},
	}
	socketPath := s.SocketPath()
	log.Infof("listen unix socket on %s", socketPath)
	return http2.NewSocketListener(socketPath)
}

// ListenUnixSocket listens on unix socket,
// and update the symbolic link of unix socket file.
func (s *Server) ListenUnixSocket() (err error) {
	s.UnixListener, err = s.NewUnixListener()
	if err != nil {
		log.Error(err)
		process.ExitWithFailure(constant.EXIT_CODE_ERROR_SERVER_LISTEN, fmt.Sprintf("create socket listerner failed: %v\n", err))
	}
	return nil
}

// ListenTcpSocket listens on tcp socket,
// and address is configured in ServerConfig.
func (s *Server) ListenTcpSocket() (err error) {
	s.HttpServer = &http.Server{
		Handler:      s.Router,
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
	}
	log.Infof("listen tcp socket on %s", s.Config.Address)
	s.TcpListener, err = http2.NewTcpListener(s.Config.Address)
	if err != nil {
		log.Error(err)
		process.ExitWithFailure(constant.EXIT_CODE_ERROR_SERVER_LISTEN, fmt.Sprintf("create tcp listerner failed: %v\n", err))
	}
	return nil
}

// RunLocalServer creates services based on the unix socket
func (s *Server) RunLocalServer() {
	go func() {
		err := s.LocalHttpServer.Serve(s.UnixListener)
		if err != nil {
			log.WithError(err).Error("obshell serve on unix listener failed")
			if s.SocketPath() == path.ObshellTmpSocketPath() {
				return
			}
			if s.IsStarting() {
				process.ExitWithFailure(constant.EXIT_CODE_ERROR_SERVER_LISTEN, fmt.Sprintf("serve on unix listener failed: %v\n", err))
			}
		}
	}()

	// for upgrade mode, we need to set state to running after socket file is created
	if s.SocketPath() == path.ObshellTmpSocketPath() {
		s.setState(constant.STATE_RUNNING)
	}
}

// RunTcpServer creates services based on the tcp socket
func (s *Server) RunTcpServer() {
	log.Info("run tcp server")
	go func() {
		var err error
		if global.EnableHTTPS {
			keyFile, certFile := path.ObshellCertificateAndKeyPaths()
			log.Infof("listen tcp socket with tls on %s", s.Config.Address)
			err = s.HttpServer.ServeTLS(s.TcpListener, certFile, keyFile)
		} else {
			log.Infof("listen tcp socket on %s", s.Config.Address)
			err = s.HttpServer.Serve(s.TcpListener)
		}
		if err != nil {
			log.WithError(err).Error("serve on tcp listener failed")
			if s.IsStarting() {
				process.ExitWithFailure(constant.EXIT_CODE_ERROR_SERVER_LISTEN, fmt.Sprintf("serve on tcp listener failed: %v\n", err))
			}
		}
	}()
	s.setState(constant.STATE_RUNNING)
}

func (s *Server) SocketPath() string {
	if s.Config.UpgradeMode {
		return path.ObshellTmpSocketPath()
	} else {
		return path.ObshellSocketPath()
	}
}

func (s *Server) Stop() {
	s.setState(constant.STATE_STOPPING)
	if s.TcpListener != nil && s.HttpServer != nil {
		if err := s.HttpServer.Shutdown(context.Background()); err != nil {
			log.WithError(err).Error("stop http server got error")
		}
	}

	if s.UnixListener != nil && s.LocalHttpServer != nil {
		if err := s.LocalHttpServer.Shutdown(context.Background()); err != nil {
			log.WithError(err).Error("stop local http server got error")
		}
	}
	s.setState(constant.STATE_STOPPED)
}

func (s *Server) GetState() int32 {
	return s.state.GetState()
}

func (s *Server) setState(state int32) {
	log.Infof("set web server state to %d", state)
	s.state.SetState(state)
}

func (s *Server) IsStarting() bool {
	return s.state.IsStarting()
}
