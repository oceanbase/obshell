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

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/errors"
)

// CheckPathValid validates the specified path according to several rules.
// It ensures that the path:
//  1. starts with a forward slash ("/"),
//  2. matches a specific pattern that includes alphanumeric characters, Chinese characters,
//     certain special characters (-_: @/.), and
//  3. does not lead to directory traversal issues when joined with a generated parent path.
func CheckPathValid(path string) error {
	if !strings.HasPrefix(path, "/") {
		return errors.Errorf("path '%s' should start with '/'", path)
	}
	pattern := "^[a-zA-Z0-9\u4e00-\u9fa5\\-_:@/\\.]*$"
	match, err := regexp.MatchString(pattern, path)
	if err != nil {
		return errors.Wrapf(err, "match pattern %s failed", pattern)
	}
	if !match {
		return fmt.Errorf("%s is not matched", path)
	}

	parentPath := fmt.Sprintf("/%s", uuid.New().String())
	absolutePath := filepath.Join(parentPath, path)
	normalizedPath := filepath.Clean(absolutePath)
	if !strings.HasPrefix(normalizedPath, parentPath) {
		log.Errorf("'%s' is not a valid path, absolutePath is %s, normalizedPath is %s", path, absolutePath, normalizedPath)
		return fmt.Errorf("%s is not a valid path", path)
	}
	return nil
}

// CheckPathExistAndValid checks if the provided filesystem path exists and is valid.
// It returns an error if the path does not exist or if the path is invalid according to the
// CheckPathValid function's criteria.
func CheckPathExistAndValid(path string) error {
	if _, err := os.Stat(path); err != nil {
		return errors.Wrapf(err, "path %s not exist", path)
	}
	return CheckPathValid(path)
}

// CheckDirExists checks if the provided filesystem path exists and is a dir.
// It returns an error if the path does not exist or if the path is not a dir.
func CheckDirExists(dir string) error {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return errors.Wrapf(err, "failed to stat path %s", dir)
	}

	if !fileInfo.IsDir() {
		return errors.Errorf("path '%s' is not a directory", dir)
	}
	return nil
}

func CheckDirEmpty(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed to open directory %s", path)
	}
	defer dir.Close()

	_, err = dir.Readdir(1)
	if err == io.EOF {
		return nil // Directory is empty
	}
	if err != nil {
		return errors.Wrapf(err, "failed to read directory %s", path)
	}
	return errors.Errorf("directory '%s' is not empty", path)
}
