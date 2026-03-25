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

package param

import (
	"fmt"
	"regexp"
	"strings"

	oberrors "github.com/oceanbase/obshell/ob/agent/errors"
)

type BackupStorageType string

const (
	Cos BackupStorageType = "cos"
	Oss BackupStorageType = "oss"
	S3  BackupStorageType = "s3"
)

type ObjectStoragePath struct {
	StorageType BackupStorageType `json:"storage_type"`
	BucketName  string            `json:"bucket_name"`
	ObjectName  string            `json:"object_name"`
}

var storageURIRegexp = regexp.MustCompile(`^(?:oss://|cos://|s3://)?([^/]+)(?:/(.*?)/?)?$`)

func ParseStorageUri(uri string) (*ObjectStoragePath, error) {
	if uri == "" {
		return nil, oberrors.Occur(oberrors.ErrCommonIllegalArgument, "uri cannot be empty")
	}

	matches := storageURIRegexp.FindStringSubmatch(uri)
	if matches == nil {
		return nil, oberrors.Occur(oberrors.ErrObStorageURIInvalid, uri)
	}

	bucketName := strings.Trim(matches[1], "/")
	var objectName string
	if len(matches) > 2 && matches[2] != "" {
		objectName = strings.Trim(matches[2], "/")
	}

	if bucketName == "" {
		return nil, oberrors.Occur(oberrors.ErrObStorageURIInvalid, "bucket name cannot be empty")
	}

	if !strings.Contains(uri, "://") {
		return nil, oberrors.Occur(oberrors.ErrObStorageURIInvalid, "URI must contain storage type prefix")
	}

	prefixParts := strings.SplitN(uri, "://", 2)
	storageType := BackupStorageType(prefixParts[0])

	switch storageType {
	case Oss, Cos, S3:
	default:
		return nil, oberrors.Occur(oberrors.ErrObStorageURIInvalid, fmt.Sprintf("unsupported storage type: %s", storageType))
	}

	return &ObjectStoragePath{
		StorageType: storageType,
		BucketName:  bucketName,
		ObjectName:  objectName,
	}, nil
}

type ValidateSharedStorageKeyParam struct {
	AccessKey string `json:"access_key" binding:"required"`
	SecretKey string `json:"secret_key" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Endpoint  string `json:"endpoint" binding:"required"`
}

type SaveSharedStorageKeyParam struct {
	AccessKey string `json:"access_key" binding:"required"`
	SecretKey string `json:"secret_key" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Endpoint  string `json:"endpoint" binding:"required"`
}
