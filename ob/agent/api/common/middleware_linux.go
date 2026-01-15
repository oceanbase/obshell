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

//go:build linux
// +build linux

package common

import (
	"net"
	"net/http"
	"syscall"
)

// ucredWrapper wraps syscall.Ucred to implement PeerCred interface
type ucredWrapper struct {
	*syscall.Ucred
}

func (u *ucredWrapper) Uid() uint32 {
	return u.Ucred.Uid
}

func (u *ucredWrapper) Gid() uint32 {
	return u.Ucred.Gid
}

func (u *ucredWrapper) Pid() int32 {
	return int32(u.Ucred.Pid)
}

// getPeerCred retrieves the Unix credentials (UID, GID, PID) of the peer process of a Unix Domain Socket connection.
// It uses the 'SO_PEERCRED' socket option to get the credentials from an HTTP request object.
func getPeerCred(r *http.Request) interface{} {
	var ucred *syscall.Ucred
	iconn := r.Context().Value(UNIX_CONNECT)
	if iconn == nil {
		return nil
	}

	conn, ok := iconn.(net.Conn)
	if !ok {
		return nil
	}

	rawConn, err := conn.(*net.UnixConn).SyscallConn()
	if err != nil {
		return nil
	}

	err = rawConn.Control(func(fd uintptr) {
		var err error
		ucred, err = syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if err != nil {
			return
		}
	})
	if err != nil {
		return nil
	}
	if ucred == nil {
		return nil
	}
	return &ucredWrapper{Ucred: ucred}
}
