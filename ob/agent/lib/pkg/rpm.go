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
	"bufio"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"

	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	"github.com/oceanbase/obshell/ob/agent/errors"
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
	installPath := filepath.Dir(path)
	return InstallRpmPkgToTargetDir(path, installPath)
}

func InstallRpmPkgToTargetDir(path string, installPath string) (err error) {
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

	// Get payload start position
	// After rpm.Read(), the file position is already at the payload start
	// Get current position as payload start
	payloadStart, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return errors.Wrapf(err, "get payload position")
	}

	var bufferedReader *bufio.Reader
	// Try to use system command for xz decompression (faster)
	if reader, cleanup, err := NewXzSystemReader(f, payloadStart); err == nil {
		defer cleanup()
		bufferedReader = bufio.NewReaderSize(reader, 256*1024)
	} else {
		// Fallback to pure Go xz implementation
		// Reset file position since NewXzSystemReader may have changed it
		if _, err := f.Seek(payloadStart, io.SeekStart); err != nil {
			return errors.Wrapf(err, "seek to payload")
		}
		reader, err := xz.NewReader(f)
		if err != nil {
			return errors.Wrapf(err, "create xz reader")
		}
		bufferedReader = bufio.NewReaderSize(reader, 256*1024)
	}

	cpioReader := cpio.NewReader(bufferedReader)

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

// newXzSystemReader tries to use system xzcat/unxz command for faster decompression
func NewXzSystemReader(rpmFile multipart.File, payloadStart int64) (io.Reader, func(), error) {
	// Try xzcat first, then unxz -c
	var cmd *exec.Cmd
	var cmdName string

	if _, err := exec.LookPath("xzcat"); err == nil {
		cmdName = "xzcat"
	} else if _, err := exec.LookPath("unxz"); err == nil {
		cmdName = "unxz"
	} else {
		return nil, nil, errors.Occur(errors.ErrEmpty, "xzcat/unxz not available")
	}

	// Check if multipart.File is *os.File to get the file path
	// If not, we cannot use system command and should return error to fallback
	osFile, ok := rpmFile.(*os.File)
	if !ok {
		return nil, nil, errors.Occur(errors.ErrEmpty, "multipart.File is not *os.File, cannot use system command")
	}

	// Get the file path from *os.File
	rpmPath := osFile.Name()

	// Open a new file handle for the command (system commands need a file path)
	// We need a separate file handle because the command reads from stdin
	f, err := os.Open(rpmPath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "open rpm")
	}

	// Seek to payload start position
	if _, err := f.Seek(payloadStart, io.SeekStart); err != nil {
		f.Close()
		return nil, nil, errors.Wrapf(err, "seek to payload")
	}

	if cmdName == "xzcat" {
		cmd = exec.Command("xzcat")
	} else {
		cmd = exec.Command("unxz", "-c")
	}

	cmd.Stdin = f
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		f.Close()
		return nil, nil, errors.Wrapf(err, "create stdout pipe")
	}

	if err := cmd.Start(); err != nil {
		f.Close()
		stdout.Close()
		return nil, nil, errors.Wrapf(err, "start %s", cmdName)
	}
	return bufio.NewReaderSize(stdout, 256*1024), func() {
		stdout.Close()
		cmd.Process.Kill()
		cmd.Wait()
		f.Close()
	}, nil
}
