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

//go:build darwin
// +build darwin

package common

import (
	"net"
	"net/http"

	"golang.org/x/sys/unix"
)

// LOCAL_PEERPID is defined in sys/un.h on macOS
// It's used to retrieve the peer process ID from a Unix domain socket
const LOCAL_PEERPID = 0x002

// xucredWrapper wraps unix.Xucred to implement PeerCred interface
type xucredWrapper struct {
	*unix.Xucred
	pid int32
}

func (u *xucredWrapper) Uid() uint32 {
	return u.Xucred.Uid
}

func (u *xucredWrapper) Gid() uint32 {
	if u.Xucred.Ngroups > 0 {
		return u.Xucred.Groups[0]
	}
	return 0
}

func (u *xucredWrapper) Pid() int32 {
	return u.pid
}

// getPeerCred retrieves the Unix credentials (UID, GID, PID) of the peer process of a Unix Domain Socket connection.
// macOS version: Uses LOCAL_PEERCRED socket option to get peer credentials (UID, GID) and LOCAL_PEERPID for PID.
func getPeerCred(r *http.Request) interface{} {
	iconn := r.Context().Value(UNIX_CONNECT)
	if iconn == nil {
		return nil
	}

	conn, ok := iconn.(net.Conn)
	if !ok {
		return nil
	}

	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return nil
	}

	rawConn, err := unixConn.SyscallConn()
	if err != nil {
		return nil
	}

	var xucred *unix.Xucred
	var pid int32
	err = rawConn.Control(func(fd uintptr) {
		// Get peer credentials (UID, GID) using LOCAL_PEERCRED
		var getErr error
		xucred, getErr = unix.GetsockoptXucred(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
		if getErr != nil {
			return
		}

		// Get peer PID using LOCAL_PEERPID
		peerPid, getErr := unix.GetsockoptInt(int(fd), unix.SOL_LOCAL, LOCAL_PEERPID)
		if getErr == nil {
			pid = int32(peerPid)
		}
		// If getting PID fails, pid remains 0, which is acceptable
	})
	if err != nil || xucred == nil {
		return nil
	}

	return &xucredWrapper{
		Xucred: xucred,
		pid:    pid,
	}
}
