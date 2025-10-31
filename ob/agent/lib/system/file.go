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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsFileExist(path string) bool {
	var err error
	if _, err = os.Stat(path); err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func CopyFile(src, dest string) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func GetBinaryVersion(path string) (string, error) {
	cmd := exec.Command(path, "version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func CopyDirs(src, dest string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return CopyFile(src, dest)
	}
	if err = os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}
	entries, err := in.Readdir(0)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		subSrc := filepath.Join(src, entry.Name())
		subDest := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			if err = CopyDirs(subSrc, subDest); err != nil {
				return err
			}
		} else {
			if err = CopyFile(subSrc, subDest); err != nil {
				return err
			}
		}
	}
	return nil
}
