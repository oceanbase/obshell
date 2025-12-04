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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
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
	obType               modelob.OBType
}

func UpgradePkgUpload(input multipart.File) (*oceanbase.UpgradePkgInfo, error) {
	r := &upgradeRpmPkgInfo{
		rpmFile: input,
	}
	obType, err := obclusterService.GetOBType()
	if err != nil {
		return nil, err
	}
	r.obType = obType

	if err := r.CheckUpgradePkg(true); err != nil {
		return nil, err
	}

	record, err := obclusterService.DumpUpgradePkgInfoAndChunkTx(r.rpmPkg, r.rpmFile, r.upgradeDepYml)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// TODO: Skip this check during upgrade since it has already been checked during upload and this method is quite time-consuming.
func (r *upgradeRpmPkgInfo) CheckUpgradePkg(forUpload bool) (err error) {
	if r.rpmPkg, err = ReadRpm(r.rpmFile); err != nil {
		return
	}
	r.version = r.rpmPkg.Version()
	switch r.rpmPkg.Name() {
	case constant.PKG_OBSHELL:
		r.isAgentPkg = true
		err = r.fileCheck(forUpload)
	case constant.PKG_OCEANBASE_CE:
		if r.obType == modelob.OBTypeCommunity {
			err = r.fileCheck(forUpload)
		} else {
			err = errors.Occur(errors.ErrObPackageNameNotSupport, r.rpmPkg.Name(), strings.Join(constant.SUPPORT_PKG_NAMES_MAP[r.obType], ", "))
		}
		log.Info("rpm dep yml is ", r.upgradeDepYml)
	case constant.PKG_OCEANBASE:
		if r.obType == modelob.OBTypeBusiness {
			err = r.fileCheck(forUpload)
		} else {
			err = errors.Occur(errors.ErrObPackageNameNotSupport, r.rpmPkg.Name(), strings.Join(constant.SUPPORT_PKG_NAMES_MAP[r.obType], ", "))
		}
		log.Info("rpm dep yml is ", r.upgradeDepYml)
	case constant.PKG_OCEANBASE_STANDALONE:
		if r.obType == modelob.OBTypeStandalone {
			err = r.fileCheck(forUpload)
		} else {
			err = errors.Occur(errors.ErrObPackageNameNotSupport, r.rpmPkg.Name(), strings.Join(constant.SUPPORT_PKG_NAMES_MAP[r.obType], ", "))
		}
		log.Info("rpm dep yml is ", r.upgradeDepYml)
	case constant.PKG_OCEANBASE_CE_LIBS:
		if r.obType == modelob.OBTypeCommunity {
			err = r.dirCheckForLibs()
		} else {
			err = errors.Occur(errors.ErrObPackageNameNotSupport, r.rpmPkg.Name(), strings.Join(constant.SUPPORT_PKG_NAMES_MAP[r.obType], ", "))
		}
	default:
		err = errors.Occur(errors.ErrObPackageNameNotSupport, r.rpmPkg.Name(), strings.Join(constant.SUPPORT_PKG_NAMES_MAP[r.obType], ", "))
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

func (r *upgradeRpmPkgInfo) fileCheck(forUpload bool) (err error) {
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
	if r.needGetUpgardeDepYml && forUpload {
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
	log.Infof("obType is %s", r.obType)
	var expected []string
	if r.obType == modelob.OBTypeBusiness {
		expected = defalutFilesForOB[:]
	} else {
		expected = append(defalutFilesForOB, defaultFilesForAgent...)
	}
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

	start := time.Now()
	var bufferedReader *bufio.Reader

	// Get payload start position
	// Since rpm.Read() positions the reader at payload start, we need to re-read the headers
	// to get the current position, or we can seek to start and read again
	if _, err := r.rpmFile.Seek(0, io.SeekStart); err != nil {
		return errors.Wrapf(err, "seek to start")
	}

	// Re-read headers to position at payload start (without MD5Check for performance)
	_, err = rpm.Read(r.rpmFile)
	if err != nil {
		return errors.Wrapf(err, "read rpm headers")
	}

	// Get current position as payload start (after rpm.Read(), position is at payload start)
	payloadStart, err := r.rpmFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return errors.Wrapf(err, "get payload position")
	}

	if reader, cleanup, err := pkg.NewXzSystemReader(r.rpmFile, payloadStart); err == nil {
		defer cleanup()
		bufferedReader = bufio.NewReaderSize(reader, 256*1024)
	} else {
		// Fallback to pure Go xz implementation
		// Reset file position since NewXzSystemReader may have changed it
		if _, err := r.rpmFile.Seek(payloadStart, io.SeekStart); err != nil {
			return errors.Wrapf(err, "seek to payload")
		}
		reader, err := xz.NewReader(r.rpmFile)
		if err != nil {
			return errors.Wrapf(err, "create xz reader")
		}
		bufferedReader = bufio.NewReaderSize(reader, 256*1024)
	}
	cpioReader := cpio.NewReader(bufferedReader)
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
			break
		}
	}
	elapsed := time.Since(start)
	log.Infof("time taken to get upgrade dep yml: %s", elapsed)
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
