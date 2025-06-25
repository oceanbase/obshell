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

package obproxy

import (
	"fmt"
	"mime/multipart"

	"github.com/cavaliergopher/rpm"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/pkg"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

var defaultFileFormatForObproxy = "/home/admin/obproxy-%s/bin/obproxy"

type upgradeRpmPkgInfo struct {
	rpmFile      multipart.File
	rpmPkg       *rpm.Package
	version      string
	release      string
	distribution string
}

func UpgradePkgUpload(input multipart.File) (*sqlite.UpgradePkgInfo, error) {
	r := &upgradeRpmPkgInfo{
		rpmFile: input,
	}

	if err := r.CheckUpgradePkg(); err != nil {
		return nil, err
	}

	pkgInfo, err := obproxyService.DumpUpgradePkgInfoAndChunkTx(r.rpmPkg, r.rpmFile)
	if err != nil {
		return nil, err
	}
	return pkgInfo, nil
}

func (r *upgradeRpmPkgInfo) CheckUpgradePkg() (err error) {
	if r.rpmPkg, err = pkg.ReadRpm(r.rpmFile); err != nil {
		return
	}
	r.version = r.rpmPkg.Version()

	if r.rpmPkg.Name() != constant.PKG_OBPROXY_CE {
		return errors.Occur(errors.ErrOBProxyPackageNameInvalid, r.rpmPkg.Name(), constant.PKG_OBPROXY_CE)
	}
	return r.fileCheck()
}

func (r *upgradeRpmPkgInfo) fileCheck() (err error) {
	// Check for the necessary files required for the agent upgrade process.
	if err = r.checkVersion(); err != nil {
		return errors.Wrap(err, "failed to check version and release")
	}
	return r.findAllExpectedFiles([]string{defaultFileFormatForObproxy})
}

func (r *upgradeRpmPkgInfo) checkVersion() (err error) {
	log.Info("version is ", r.version)
	r.release, r.distribution, err = pkg.SplitRelease(r.rpmPkg.Release())
	if err != nil {
		return
	}
	if pkg.CompareVersion(r.rpmPkg.Version(), constant.OBPROXY_MIN_VERSION_SUPPORT) < 0 {
		return errors.Occur(errors.ErrOBProxyVersionNotSupported, r.rpmPkg.Version(), constant.SUPPORT_MIN_VERSION)
	}
	return nil
}

func (r *upgradeRpmPkgInfo) findAllExpectedFiles(expected []string) (err error) {
	succeed := true
	missingFiles := make([]string, 0)
	for _, expect := range expected {
		expect = fmt.Sprintf(expect, r.version)
		var found bool
		for _, actual := range r.rpmPkg.Files() {
			if actual.Name() == expect {
				log.Info("found file: ", expect)
				found = true
				break
			}
		}
		if !found {
			log.Errorf("file '%s' not 	found", expect)
			missingFiles = append(missingFiles, expect)
			succeed = false
		}
	}
	if !succeed {
		return errors.Occur(errors.ErrOBProxyPackageMissingFile, missingFiles)
	}
	return nil
}
