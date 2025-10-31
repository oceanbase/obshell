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

package rpm

import (
	"path/filepath"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/lib/http"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
)

func CallUploadPkgAndPrint(pkgDir, fileName string) (err error) {
	var ret oceanbase.UpgradePkgInfo
	filePath := filepath.Join(pkgDir, fileName)

	stdio.StartLoadingf("Uploading package %s", fileName)
	uri := constant.URI_API_V1 + constant.URI_UPGRADE + constant.URI_PACKAGE
	if err = http.UploadFileViaUnixSocket(path.ObshellSocketPath(), uri, filePath, &ret); err != nil {
		stdio.LoadErrorf("Upload package %s failed: %s", fileName, err)
		return
	}

	stdio.LoadSuccessf("Upload package %s successfully!", fileName)
	return nil
}
