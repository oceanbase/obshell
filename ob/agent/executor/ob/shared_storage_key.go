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

package ob

import (
	"context"
	"fmt"
	"strings"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
	"github.com/oceanbase/obshell/ob/param"
)

func ValidateSharedStorageKey(params *param.ValidateSharedStorageKeyParam) error {
	isSharedStorage, err := obclusterService.IsSharedStorageMode()

	if err != nil {
		return errors.Occurf(errors.ErrSharedStorageKeyValidateFailed, err.Error())
	}

	if !isSharedStorage {
		return errors.Occur(errors.ErrSharedStorageNotSupported)
	}

	if err := validateSharedStorageKeyParams(params); err != nil {
		return err
	}

	storageURI, err := buildStorageURI(params)
	if err != nil {
		return err
	}

	storage, err := system.GetStorageInterfaceByURI(storageURI)
	if err != nil {
		return errors.Occurf(errors.ErrSharedStorageKeyValidateFailed, "failed to get storage interface: %v", err)
	}

	if err := storage.CheckWritePermission(); err != nil {
		return errors.Occurf(errors.ErrSharedStorageKeyValidateFailed, "check write permission failed: %v", err)
	}

	return nil
}

func SaveSharedStorageKey(ctx context.Context, params *param.SaveSharedStorageKeyParam) error {
	isSharedStorage, err := obclusterService.IsSharedStorageMode()

	if err != nil {
		return errors.Occurf(errors.ErrSharedStorageKeyValidateFailed, err.Error())
	}

	if !isSharedStorage {
		return errors.Occur(errors.ErrSharedStorageNotSupported)
	}

	if err := validateSaveSharedStorageKeyParams(params); err != nil {
		return err
	}

	fullPath := joinPathQuery(params.Path, params.Endpoint)
	accessInfo := fmt.Sprintf("access_id=%s&access_key=%s", params.AccessKey, params.SecretKey)
	if err := obclusterService.UpdateSharedStorageConfig(ctx, fullPath, accessInfo); err != nil {
		return errors.Occurf(errors.ErrCommonUnexpected, "failed to execute ALTER SYSTEM statement: %v", err)
	}

	return nil
}

func joinPathQuery(path, query string) string {
	if strings.Contains(path, "?") {
		return path + "&" + query
	}
	return path + "?" + query
}

func validateSharedStorageKeyParams(params *param.ValidateSharedStorageKeyParam) error {
	if _, err := param.ParseStorageUri(params.Path); err != nil {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "path", err.Error())
	}
	return nil
}

func validateSaveSharedStorageKeyParams(params *param.SaveSharedStorageKeyParam) error {
	if _, err := param.ParseStorageUri(params.Path); err != nil {
		return errors.Occur(errors.ErrCommonIllegalArgumentWithMessage, "path", err.Error())
	}
	return nil
}

func buildStorageURI(params *param.ValidateSharedStorageKeyParam) (string, error) {
	pathInfo, err := param.ParseStorageUri(params.Path)
	if err != nil {
		return "", err
	}

	var uri string
	switch pathInfo.StorageType {
	case param.Oss:
		uri = fmt.Sprintf("%s%s/%s?host=%s&access_id=%s&access_key=%s",
			constant.PREFIX_OSS, pathInfo.BucketName, pathInfo.ObjectName, params.Endpoint, params.AccessKey, params.SecretKey)
	case param.Cos:
		uri = fmt.Sprintf("%s%s/%s?host=%s&access_id=%s&access_key=%s",
			constant.PREFIX_COS, pathInfo.BucketName, pathInfo.ObjectName, params.Endpoint, params.AccessKey, params.SecretKey)
	case param.S3:
		// force_path_style=true for Path-Style access (e.g. MinIO-compatible endpoints)
		uri = fmt.Sprintf("%s%s/%s?host=%s&access_id=%s&access_key=%s&force_path_style=true",
			constant.PREFIX_S3, pathInfo.BucketName, pathInfo.ObjectName, params.Endpoint, params.AccessKey, params.SecretKey)
	default:
		return "", errors.Occur(errors.ErrCommonIllegalArgument, "unsupported storage type: %s", pathInfo.StorageType)
	}

	return uri, nil
}
