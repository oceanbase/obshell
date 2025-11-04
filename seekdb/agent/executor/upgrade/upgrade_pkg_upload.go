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

package upgrade

import (
	"mime/multipart"

	"github.com/cavaliergopher/rpm"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/pkg"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
)

const (
	RPM_XZ_COMPRESSION = "xz"
	RPM_CPIO_PLAYLOAD  = "cpio"
)

var (
	OCEANBASE_UPGRADE_DEP_YAML = "/home/admin/oceanbase/etc/oceanbase_upgrade_dep.yml"
)

var defaultFilesForAgent = []string{
	"/home/admin/oceanbase/bin/obshell",
}

type upgradeRpmPkgInfo struct {
	rpmFile      multipart.File
	rpmPkg       *rpm.Package
	version      string
	release      string
	distribution string
}

func UpgradePkgUpload(input multipart.File) (*oceanbase.UpgradePkgInfo, error) {
	r := &upgradeRpmPkgInfo{
		rpmFile: input,
	}

	if err := r.CheckUpgradePkg(true); err != nil {
		return nil, err
	}

	record, err := obclusterService.DumpUpgradePkgInfoAndChunkTx(r.rpmPkg, r.rpmFile)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *upgradeRpmPkgInfo) CheckUpgradePkg(forUpload bool) (err error) {
	if r.rpmPkg, err = ReadRpm(r.rpmFile); err != nil {
		return
	}
	r.version = r.rpmPkg.Version()
	switch r.rpmPkg.Name() {
	case constant.PKG_OBSHELL:
		err = r.fileCheck()
	case constant.PKG_OCEANBASE_CE:
		log.Warn("oceanbase-ce is not supported")
	default:
		err = errors.Occur(errors.ErrObPackageNameNotSupport, r.rpmPkg.Name(), constant.PKG_OBSHELL)
	}
	if err != nil {
		return
	}
	return nil
}

func (r *upgradeRpmPkgInfo) fileCheck() (err error) {
	// Check for the necessary files required for the agent upgrade process.
	if err = r.checkVersion(); err != nil {
		return errors.Wrap(err, "failed to check version and release")
	}
	return r.findAllExpectedFiles(defaultFilesForAgent)
}

func (r *upgradeRpmPkgInfo) checkVersion() (err error) {
	log.Info("version is ", r.version)
	r.release, r.distribution, err = pkg.SplitRelease(r.rpmPkg.Release())
	if err != nil {
		return
	}
	if pkg.CompareVersion(r.rpmPkg.Version(), constant.SUPPORT_MIN_VERSION) < 0 {
		return errors.Occur(errors.ErrAgentOBVersionNotSupported, r.rpmPkg.Version(), constant.SUPPORT_MIN_VERSION)
	}
	return nil
}

func (r *upgradeRpmPkgInfo) findAllExpectedFiles(expected []string) (err error) {
	succeed := true
	missingFiles := make([]string, 0)
	for _, expect := range expected {
		var found bool
		for _, actual := range r.rpmPkg.Files() {
			if actual.Name() == expect {
				log.Info("found file: ", expect)
				found = true
				break
			}
		}
		if !found {
			log.Errorf("file '%s' not found", expect)
			missingFiles = append(missingFiles, expect)
			succeed = false
		}
	}
	if !succeed {
		return errors.Occur(errors.ErrObPackageMissingFile, r.rpmPkg.Name(), missingFiles)
	}
	return nil
}

func ReadRpm(input multipart.File) (pkg *rpm.Package, err error) {
	if _, err = input.Seek(0, 0); err != nil {
		return
	}
	if err = rpm.MD5Check(input); err != nil {
		return
	}
	if _, err = input.Seek(0, 0); err != nil {
		return
	}
	return rpm.Read(input)
}
