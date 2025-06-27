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
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/pkg"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
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

var defalutFilesForOB = []string{
	"/home/admin/oceanbase/bin/observer",
}

var defaultFilesForOBUpgrade = []string{
	"/home/admin/oceanbase/etc/upgrade_checker.py",
	"/home/admin/oceanbase/etc/upgrade_pre.py",
	"/home/admin/oceanbase/etc/upgrade_health_checker.py",
	"/home/admin/oceanbase/etc/upgrade_post.py",
	"/home/admin/oceanbase/etc/oceanbase_upgrade_dep.yml",
}

type upgradeRpmPkgInfo struct {
	rpmFile              multipart.File
	rpmPkg               *rpm.Package
	needGetUpgardeDepYml bool
	upgradeDepYml        string
	isAgentPkg           bool
	version              string
	release              string
	distribution         string
}

func UpgradePkgUpload(input multipart.File) (*oceanbase.UpgradePkgInfo, error) {
	r := &upgradeRpmPkgInfo{
		rpmFile: input,
	}

	if err := r.CheckUpgradePkg(true); err != nil {
		return nil, err
	}

	record, err := obclusterService.DumpUpgradePkgInfoAndChunkTx(r.rpmPkg, r.rpmFile, r.upgradeDepYml)
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
		r.isAgentPkg = true
		err = r.fileCheck()
	case constant.PKG_OCEANBASE_CE:
		err = r.fileCheck()
		log.Info("rpm dep yml is ", r.upgradeDepYml)
	case constant.PKG_OCEANBASE_CE_LIBS:
		err = r.dirCheckForLibs()
	default:
		err = errors.Occur(errors.ErrObPackageNameNotSupport, r.rpmPkg.Name(), strings.Join(constant.SUPPORT_PKG_NAMES, ", "))
	}
	if err != nil {
		return
	}
	return nil
}

func (r *upgradeRpmPkgInfo) dirCheckForLibs() (err error) {
	if err = r.checkVersion(); err != nil {
		return errors.Wrap(err, "failed to check version and release")
	}
	files := r.rpmPkg.Files()
	for _, actual := range files {
		if actual.Name() == "/home/admin/oceanbase/lib" && actual.IsDir() {
			return nil
		}
	}
	return errors.Occur(errors.ErrObPackageMissingFile, r.rpmPkg.Name(), "/home/admin/oceanbase/lib")
}

func (r *upgradeRpmPkgInfo) fileCheck() (err error) {
	// Check for the necessary files required for the agent upgrade process.
	if err = r.checkVersion(); err != nil {
		return errors.Wrap(err, "failed to check version and release")
	}
	if r.isAgentPkg {
		return r.findAllExpectedFiles(defaultFilesForAgent)
	}

	if err = r.checkFiles(); err != nil {
		return
	}
	if r.needGetUpgardeDepYml {
		log.Info("need to get upgrade dep yml")
		return r.GetUpgradeDepYml()
	}
	return nil
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

func (r *upgradeRpmPkgInfo) checkFiles() (err error) {
	log.Infof("rpm pkg name is %s", r.rpmPkg.Name())
	expected := append(defalutFilesForOB, defaultFilesForAgent...)
	obBuildVersion, err := obclusterService.GetObBuildVersion()
	if err != nil {
		return errors.Wrap(err, "get ob version failed")
	}
	obVerRel := strings.Split(obBuildVersion, "_")
	obVer := obVerRel[0]
	obRel := obVerRel[1]
	log.Infof("obBuildVersion is %s-%s, rpm version is %s-%s", obVer, obRel, r.version, r.release)
	if pkg.CompareVersion(fmt.Sprintf("%s-%s", r.version, r.release), fmt.Sprintf("%s-%s", obVer, obRel)) > 0 && r.version != obVer {
		expected = append(expected, defaultFilesForOBUpgrade...)
		r.needGetUpgardeDepYml = true
	}
	log.Infof("expected files are %v", expected)
	return r.findAllExpectedFiles(expected)
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

func (r *upgradeRpmPkgInfo) GetUpgradeDepYml() (err error) {
	log.Info("start to get upgrade dep yml")
	if err = pkg.CheckCompressAndFormat(r.rpmPkg); err != nil {
		return
	}
	xzReader, err := xz.NewReader(r.rpmFile)
	if err != nil {
		return
	}
	cpioReader := cpio.NewReader(xzReader)
	buffer := new(bytes.Buffer)
	for {
		hdr, err := cpioReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if !hdr.Mode.IsRegular() {
			continue
		}
		if hdr.Name == fmt.Sprintf(".%s", OCEANBASE_UPGRADE_DEP_YAML) {
			if _, err = io.Copy(buffer, cpioReader); err != nil {
				return err
			}
			r.upgradeDepYml = buffer.String()
		}
	}
	return
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
