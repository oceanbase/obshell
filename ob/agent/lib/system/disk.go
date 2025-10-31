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

package system

import (
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type DiskInfo struct {
	Fsid               string `json:"fsid"`
	Path               string `json:"path"`
	TotalSizeBytes     uint64 `json:"totalSizeBytes"`     // total size in bytes
	UsedSizeBytes      uint64 `json:"usedSizeBytes"`      // used size in bytes
	AvailableSizeBytes uint64 `json:"availableSizeBytes"` // available size in bytes
	FreeSizeBytes      uint64 `json:"freeSizeBytes"`      // free size in bytes
}

func GetFsId(path string) (string, syscall.Fsid, error) {
	var stat syscall.Statfs_t
	var err error
	for {
		err = syscall.Statfs(path, &stat)
		if err == nil {
			return path, stat.Fsid, nil
		} else {
			if path == "/" {
				log.Infoln("GetFsId: path is /")
				return "", syscall.Fsid{}, err
			}
			path = filepath.Dir(path)
		}
	}
}

func GetDiskInfo(path string) (*DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}
	return &DiskInfo{
		TotalSizeBytes:     (uint64(stat.Blocks) * uint64(stat.Bsize)),
		AvailableSizeBytes: (uint64(stat.Bavail) * uint64(stat.Bsize)),
		FreeSizeBytes:      (uint64(stat.Bfree) * uint64(stat.Bsize)),
	}, nil
}
