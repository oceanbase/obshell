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

package http

import (
	"context"
	"net"
	"net/http"
	"syscall"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type Listener struct {
	tcpListener  *net.TCPListener
	unixListener *net.UnixListener
	mux          *http.ServeMux
	srv          *http.Server
}

func NewListener() *Listener {
	mux := http.NewServeMux()
	return &Listener{
		mux: mux,
		srv: &http.Server{Handler: mux},
	}
}

func (l *Listener) AddHandler(path string, h http.Handler) {
	l.mux.Handle(path, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Connection", "close")
		h.ServeHTTP(writer, request)
	}))
}

func NewTcpListener(addr string) (*net.TCPListener, error) {
	cfg := net.ListenConfig{
		Control: reusePort,
	}
	listener, err := cfg.Listen(context.Background(), "tcp", addr)
	if err != nil {
		log.WithError(err).Errorf("create tcp listener on %s", addr)
		return nil, err
	}
	return listener.(*net.TCPListener), nil
}

func reusePort(network, address string, c syscall.RawConn) error {
	var err2 error
	err := c.Control(func(fd uintptr) {
		err2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if err2 != nil {
			return
		}
		err2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 0)
	})
	if err2 != nil {
		return err2
	}
	return err
}

func NewSocketListener(path string) (*net.UnixListener, error) {
	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return nil, err
	}
	return net.ListenUnix("unix", addr)
}
