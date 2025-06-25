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

package pkg

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/ulikunitz/xz"
)

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

func SplitRelease(release string) (buildNumber, distribution string, err error) {
	releaseSplit := strings.Split(release, ".")
	if len(releaseSplit) < 2 {
		return "", "", errors.Occur(errors.ErrPackageReleaseFormatInvalid, release)
	}
	buildNumber = releaseSplit[0]
	distribution = releaseSplit[len(releaseSplit)-1]
	return
}

func InstallRpmPkgInPlace(path string) (err error) {
	log.Infof("InstallRpmPkg: %s", path)
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	pkg, err := rpm.Read(f)
	if err != nil {
		return
	}
	if err = CheckCompressAndFormat(pkg); err != nil {
		return
	}

	xzReader, err := xz.NewReader(f)
	if err != nil {
		return
	}
	installPath := filepath.Dir(path)
	cpioReader := cpio.NewReader(xzReader)

	for {
		hdr, err := cpioReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		m := hdr.Mode
		if m.IsDir() {
			dest := filepath.Join(installPath, hdr.Name)
			log.Infof("%s is a directory, creating %s", hdr.Name, dest)
			if err := os.MkdirAll(dest, 0755); err != nil {
				return errors.Wrapf(err, "mkdir failed %s", hdr.Name)
			}

		} else if m.IsRegular() {
			if err := handleRegularFile(hdr, cpioReader, installPath); err != nil {
				return err
			}

		} else if hdr.Linkname != "" {
			if err := handleSymlink(hdr, installPath); err != nil {
				return err
			}
		} else {
			log.Infof("Skipping unsupported file %s type: %v", hdr.Name, m)
		}
	}

	return nil
}

func handleRegularFile(hdr *cpio.Header, cpioReader *cpio.Reader, installPath string) error {
	dest := filepath.Join(installPath, hdr.Name)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		log.WithError(err).Error("mkdir failed")
		return err
	}

	outFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer outFile.Close()

	log.Infof("Extracting %s", hdr.Name)
	if _, err := io.Copy(outFile, cpioReader); err != nil {
		return err
	}
	return nil
}

func handleSymlink(hdr *cpio.Header, installPath string) error {
	dest := filepath.Join(installPath, hdr.Name)
	if err := os.Symlink(hdr.Linkname, dest); err != nil {
		return errors.Wrapf(err, "create symlink failed %s -> %s", dest, hdr.Linkname)
	}
	log.Infof("Creating symlink %s -> %s", dest, hdr.Linkname)
	return nil
}

func CheckCompressAndFormat(pkg *rpm.Package) error {
	if pkg.PayloadCompression() != "xz" {
		return errors.Occur(errors.ErrPackageCompressionNotSupported, pkg.PayloadCompression())
	}
	if pkg.PayloadFormat() != "cpio" {
		return errors.Occur(errors.ErrPackageFormatInvalid, pkg.PayloadFormat())
	}
	return nil
}
