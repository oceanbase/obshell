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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	rpmsignature "github.com/oceanbase/obshell/rpm/signature"
)

// RpmIdentity is the package identity selected by the upgrade task before the
// RPM is copied into the path-accessible upgrade directory. OBShell self-upgrade
// compares every field with the signed RPM header so another official OBShell
// release cannot be substituted after package selection.
type RpmIdentity struct {
	Name         string
	Version      string
	Release      string
	Architecture string
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

// VerifyObshellRpmSignature verifies only packages identified as OBShell RPMs.
// Keeping the package-name guard next to the cryptographic check makes it hard
// for callers to accidentally expand verification to OceanBase or other RPMs.
func VerifyObshellRpmSignature(input multipart.File, packageName string) error {
	if packageName != "obshell" {
		return nil
	}

	signer, err := rpmsignature.Verify(input)
	if err != nil {
		return errors.WrapOverride(errors.ErrObshellPackageSignatureInvalid, err)
	}
	log.Infof("Verified OBShell RPM signature from %s", signer)
	return nil
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
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = errors.Wrap(closeErr, "close RPM package")
		}
	}()
	_, err = installRpmFileToTargetDir(f, installPath, nil)
	return err
}

// InstallVerifiedObshellRpmInPlace is the dedicated standalone OBShell upgrade
// path. Keeping verification out of InstallRpmPkg* preserves the behavior of
// OceanBase, standalone, libs, and obdiag package extraction.
func InstallVerifiedObshellRpmInPlace(path string, expectedIdentity RpmIdentity) (obshellBinarySHA256 string, err error) {
	return installVerifiedObshellRpmToTargetDir(path, filepath.Dir(path), expectedIdentity)
}

func installVerifiedObshellRpmToTargetDir(path string, installPath string, expectedIdentity RpmIdentity) (obshellBinarySHA256 string, err error) {
	if expectedIdentity.Name != "obshell" {
		return "", errors.Occur(errors.ErrPackageNameMismatch, expectedIdentity.Name, "obshell")
	}
	log.Infof("Install verified OBShell RPM: %s", path)

	f, err := openRpmForInstall(path, expectedIdentity.Name)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = errors.Wrap(closeErr, "close OBShell RPM package")
		}
	}()
	return installRpmFileToTargetDir(f, installPath, &expectedIdentity)
}

func openRpmForInstall(path string, expectedPackageName string) (*os.File, error) {
	source, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if expectedPackageName != "obshell" {
		return source, nil
	}

	// OBShell self-upgrade uses an unlinked 0600 snapshot. Signature verification
	// and extraction both read this private file, so neither path replacement nor
	// an in-place write through another descriptor can change the verified bytes.
	snapshot, err := os.CreateTemp(filepath.Dir(path), ".obshell-rpm-snapshot-*")
	if err != nil {
		if closeErr := source.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close OBShell RPM after snapshot creation failure")
		}
		return nil, errors.Wrap(err, "create OBShell RPM snapshot")
	}
	if err := os.Remove(snapshot.Name()); err != nil {
		if closeErr := snapshot.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close OBShell RPM snapshot after unlink failure")
		}
		if closeErr := source.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close OBShell RPM after snapshot unlink failure")
		}
		return nil, errors.Wrap(err, "unlink OBShell RPM snapshot")
	}
	if _, err := io.Copy(snapshot, source); err != nil {
		if closeErr := snapshot.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close incomplete OBShell RPM snapshot")
		}
		if closeErr := source.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close OBShell RPM after snapshot copy failure")
		}
		return nil, errors.Wrap(err, "copy OBShell RPM to private snapshot")
	}
	if err := source.Close(); err != nil {
		if closeErr := snapshot.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close OBShell RPM snapshot after source close failure")
		}
		return nil, errors.Wrap(err, "close OBShell RPM after snapshot")
	}
	if _, err := snapshot.Seek(0, io.SeekStart); err != nil {
		if closeErr := snapshot.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close OBShell RPM snapshot after seek failure")
		}
		return nil, errors.Wrap(err, "rewind OBShell RPM snapshot")
	}
	return snapshot, nil
}

func installRpmFileToTargetDir(f *os.File, installPath string, expectedIdentity *RpmIdentity) (obshellBinarySHA256 string, err error) {
	pkg, err := rpm.Read(f)
	if err != nil {
		return
	}
	verifiedObshellUpgrade := expectedIdentity != nil
	if verifiedObshellUpgrade {
		if pkg.Name() != expectedIdentity.Name {
			return "", errors.Occur(errors.ErrPackageNameMismatch, pkg.Name(), expectedIdentity.Name)
		}
		if err = VerifyObshellRpmSignature(f, expectedIdentity.Name); err != nil {
			return
		}
		if err = verifyObshellRpmIdentity(pkg, *expectedIdentity); err != nil {
			return "", err
		}
	}
	if err = CheckCompressAndFormat(pkg); err != nil {
		return
	}

	// Get payload start position
	// After rpm.Read(), the file position is already at the payload start
	// Get current position as payload start
	payloadStart, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", errors.Wrapf(err, "get payload position")
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
			return "", errors.Wrapf(err, "seek to payload")
		}
		reader, err := xz.NewReader(f)
		if err != nil {
			return "", errors.Wrapf(err, "create xz reader")
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
			return "", err
		}

		m := hdr.Mode
		if m.IsDir() {
			dest := filepath.Join(installPath, hdr.Name)
			log.Infof("%s is a directory, creating %s", hdr.Name, dest)
			if err := os.MkdirAll(dest, 0755); err != nil {
				return "", errors.Wrapf(err, "mkdir failed %s", hdr.Name)
			}

		} else if m.IsRegular() {
			var digest string
			if digest, err = handleRegularFile(hdr, cpioReader, installPath, verifiedObshellUpgrade); err != nil {
				return "", err
			}
			if digest != "" {
				obshellBinarySHA256 = digest
			}

		} else if hdr.Linkname != "" {
			if err := handleSymlink(hdr, installPath); err != nil {
				return "", err
			}
		} else {
			log.Infof("Skipping unsupported file %s type: %v", hdr.Name, m)
		}
	}

	if verifiedObshellUpgrade && obshellBinarySHA256 == "" {
		return "", errors.Occur(errors.ErrObPackageMissingFile, expectedIdentity.Name, "/home/admin/oceanbase/bin/obshell")
	}
	return obshellBinarySHA256, nil
}

func verifyObshellRpmIdentity(rpmPkg *rpm.Package, expected RpmIdentity) error {
	if expected.Name != "obshell" {
		return nil
	}

	actual := RpmIdentity{
		Name:         rpmPkg.Name(),
		Version:      rpmPkg.Version(),
		Release:      rpmPkg.Release(),
		Architecture: rpmPkg.Architecture(),
	}
	if expected.Version == "" || expected.Release == "" || expected.Architecture == "" {
		return errors.Occur(errors.ErrObPackageCorrupted, expected.Name, "missing expected signed RPM identity")
	}
	if actual != expected {
		return errors.Occur(
			errors.ErrObPackageCorrupted,
			expected.Name,
			fmt.Sprintf("signed RPM identity %s-%s-%s.%s does not match expected %s-%s-%s.%s",
				actual.Name, actual.Version, actual.Release, actual.Architecture,
				expected.Name, expected.Version, expected.Release, expected.Architecture),
		)
	}
	return nil
}

func handleRegularFile(hdr *cpio.Header, cpioReader *cpio.Reader, installPath string, recordObshellDigest bool) (string, error) {
	dest := filepath.Join(installPath, hdr.Name)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		log.WithError(err).Error("mkdir failed")
		return "", err
	}

	outFile, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.WithError(err).Warn("close extracted RPM file")
		}
	}()

	log.Infof("Extracting %s", hdr.Name)
	if recordObshellDigest && filepath.Clean(hdr.Name) == "home/admin/oceanbase/bin/obshell" {
		digest := sha256.New()
		if _, err := io.Copy(io.MultiWriter(outFile, digest), cpioReader); err != nil {
			return "", err
		}
		return hex.EncodeToString(digest.Sum(nil)), nil
	}
	if _, err := io.Copy(outFile, cpioReader); err != nil {
		return "", err
	}
	return "", nil
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

	// Keep the decompressor bound to the already opened file descriptor. Reopening
	// osFile.Name() here would allow the path to be replaced after signature
	// verification but before extraction.
	osFile, ok := rpmFile.(*os.File)
	if !ok {
		return nil, nil, errors.Occur(errors.ErrEmpty, "multipart.File is not *os.File, cannot use system command")
	}
	payload := io.NewSectionReader(osFile, payloadStart, 1<<63-1-payloadStart)

	if cmdName == "xzcat" {
		cmd = exec.Command("xzcat")
	} else {
		cmd = exec.Command("unxz", "-c")
	}

	cmd.Stdin = payload
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "create stdout pipe")
	}

	if err := cmd.Start(); err != nil {
		if closeErr := stdout.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("close xz stdout pipe after start failure")
		}
		return nil, nil, errors.Wrapf(err, "start %s", cmdName)
	}
	return bufio.NewReaderSize(stdout, 256*1024), func() {
		if err := stdout.Close(); err != nil {
			log.WithError(err).Debug("close xz stdout pipe")
		}
		if err := cmd.Process.Kill(); err != nil {
			log.WithError(err).Debug("stop xz process")
		}
		if err := cmd.Wait(); err != nil {
			log.WithError(err).Debug("wait for xz process")
		}
	}, nil
}
