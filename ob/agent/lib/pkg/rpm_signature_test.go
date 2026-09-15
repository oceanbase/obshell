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
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	agenterrors "github.com/oceanbase/obshell/ob/agent/errors"
)

//go:embed testdata/oceanbase-ce-libs-4.5.0.0.el7.x86_64.rpm.b64
var testSignedRPMBase64 []byte

func TestVerifyObshellRpmSignature(t *testing.T) {
	testSignedRPM := loadSignedTestRPM(t)

	t.Run("accepts fixture signed by trusted release key", func(t *testing.T) {
		input := &multipartFile{Reader: *bytes.NewReader(testSignedRPM)}
		if err := VerifyObshellRpmSignature(input, "obshell"); err != nil {
			t.Fatalf("verify trusted RPM: %v", err)
		}
	})

	t.Run("rejects tampered package", func(t *testing.T) {
		tampered := bytes.Clone(testSignedRPM)
		tampered[len(tampered)-1] ^= 0xff

		input := &multipartFile{Reader: *bytes.NewReader(tampered)}
		if err := VerifyObshellRpmSignature(input, "obshell"); err == nil {
			t.Fatal("expected tampered RPM to be rejected")
		}
	})

	t.Run("rejects package with damaged signature header", func(t *testing.T) {
		unsigned := bytes.Clone(testSignedRPM)
		for index := 104; index < 104+16; index++ {
			unsigned[index] = 0
		}

		input := &multipartFile{Reader: *bytes.NewReader(unsigned)}
		if err := VerifyObshellRpmSignature(input, "obshell"); err == nil {
			t.Fatal("expected RPM with a damaged signature header to be rejected")
		}
	})

	t.Run("rejects structurally valid RPM without a signature", func(t *testing.T) {
		unsigned := removeRpmGPGSignatureTags(t, testSignedRPM)
		input := &multipartFile{Reader: *bytes.NewReader(unsigned)}
		err := VerifyObshellRpmSignature(input, "obshell")
		if err == nil {
			t.Fatal("expected unsigned RPM to be rejected")
		}
		agentErr, ok := err.(agenterrors.OcsAgentErrorInterface)
		if !ok {
			t.Fatalf("unsigned RPM returned an untyped error: %T: %v", err, err)
		}
		if actual := agentErr.ErrorCode().Code; actual != agenterrors.ErrObshellPackageSignatureInvalid.Code {
			t.Fatalf("unsigned RPM error code is %q, want %q", actual, agenterrors.ErrObshellPackageSignatureInvalid.Code)
		}
	})

	t.Run("preserves reader position", func(t *testing.T) {
		input := &multipartFile{Reader: *bytes.NewReader(testSignedRPM)}
		const originalOffset int64 = 137
		if _, err := input.Seek(originalOffset, io.SeekStart); err != nil {
			t.Fatalf("set reader position: %v", err)
		}
		if err := VerifyObshellRpmSignature(input, "obshell"); err != nil {
			t.Fatalf("verify trusted RPM: %v", err)
		}
		offset, err := input.Seek(0, io.SeekCurrent)
		if err != nil {
			t.Fatalf("get reader position: %v", err)
		}
		if offset != originalOffset {
			t.Fatalf("reader position changed: got %d, want %d", offset, originalOffset)
		}
	})

	t.Run("does not verify OceanBase package", func(t *testing.T) {
		input := &multipartFile{Reader: *bytes.NewReader([]byte("not a signed RPM"))}
		if err := VerifyObshellRpmSignature(input, "oceanbase-ce"); err != nil {
			t.Fatalf("OceanBase RPM unexpectedly required a signature: %v", err)
		}
	})
}

func TestVerifyRealObshellReleaseRpm(t *testing.T) {
	path := os.Getenv("OBSHELL_SIGNED_RPM_TEST_PATH")
	if path == "" {
		t.Skip("OBSHELL_SIGNED_RPM_TEST_PATH is not set")
	}

	input, err := os.Open(path)
	if err != nil {
		t.Fatalf("open OBShell release RPM: %v", err)
	}
	defer func() {
		if err := input.Close(); err != nil {
			t.Errorf("close OBShell release RPM: %v", err)
		}
	}()

	rpmPkg, err := ReadRpm(input)
	if err != nil {
		t.Fatalf("read OBShell release RPM: %v", err)
	}
	if rpmPkg.Name() != "obshell" {
		t.Fatalf("fixture package name is %q, want obshell", rpmPkg.Name())
	}
	if err := VerifyObshellRpmSignature(input, rpmPkg.Name()); err != nil {
		t.Fatalf("verify OBShell release RPM: %v", err)
	}
}

func TestInstallRealObshellReleaseRpm(t *testing.T) {
	path := os.Getenv("OBSHELL_SIGNED_RPM_TEST_PATH")
	if path == "" {
		t.Skip("OBSHELL_SIGNED_RPM_TEST_PATH is not set")
	}

	expectedIdentity := readRpmIdentity(t, path)
	installPath := t.TempDir()
	if _, err := installVerifiedObshellRpmToTargetDir(path, installPath, expectedIdentity); err != nil {
		t.Fatalf("verify and unpack OBShell release RPM: %v", err)
	}
	obshellPath := filepath.Join(installPath, "home/admin/oceanbase/bin/obshell")
	if _, err := os.Stat(obshellPath); err != nil {
		t.Fatalf("stat unpacked OBShell binary: %v", err)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read OBShell release RPM: %v", err)
	}
	damaged := damageRpmGPGSignature(t, contents)
	damagedPath := filepath.Join(t.TempDir(), "obshell-with-damaged-signature.rpm")
	if err := os.WriteFile(damagedPath, damaged, 0600); err != nil {
		t.Fatalf("write damaged OBShell RPM: %v", err)
	}
	if _, err := installVerifiedObshellRpmToTargetDir(damagedPath, t.TempDir(), expectedIdentity); err == nil {
		t.Fatal("expected install to reject OBShell RPM with a damaged signature")
	}
}

func TestInstallRejectsUnexpectedObshellVersion(t *testing.T) {
	path := os.Getenv("OBSHELL_SIGNED_RPM_TEST_PATH")
	if path == "" {
		t.Skip("OBSHELL_SIGNED_RPM_TEST_PATH is not set")
	}

	expectedIdentity := readRpmIdentity(t, path)
	expectedIdentity.Version += ".unexpected"
	installPath := t.TempDir()
	if _, err := installVerifiedObshellRpmToTargetDir(path, installPath, expectedIdentity); err == nil {
		t.Fatal("expected install to reject a trusted OBShell RPM with an unexpected version")
	}
	obshellPath := filepath.Join(installPath, "home/admin/oceanbase/bin/obshell")
	if _, err := os.Stat(obshellPath); !os.IsNotExist(err) {
		t.Fatalf("unexpected-version OBShell RPM was extracted: %v", err)
	}
}

func TestInstallRealObshellSnapshotIgnoresSourceRewrite(t *testing.T) {
	path := os.Getenv("OBSHELL_SIGNED_RPM_TEST_PATH")
	if path == "" {
		t.Skip("OBSHELL_SIGNED_RPM_TEST_PATH is not set")
	}

	releaseRPM, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read OBShell release RPM: %v", err)
	}
	sourcePath := filepath.Join(t.TempDir(), "obshell.rpm")
	if err := os.WriteFile(sourcePath, releaseRPM, 0600); err != nil {
		t.Fatalf("write OBShell source RPM: %v", err)
	}
	expectedIdentity := readRpmIdentity(t, path)

	snapshot, err := openRpmForInstall(sourcePath, "obshell")
	if err != nil {
		t.Fatalf("create private OBShell RPM snapshot: %v", err)
	}
	defer func() {
		if err := snapshot.Close(); err != nil {
			t.Errorf("close private OBShell RPM snapshot: %v", err)
		}
	}()
	if _, err := os.Stat(snapshot.Name()); !os.IsNotExist(err) {
		t.Fatalf("private OBShell RPM snapshot is still path-accessible: %v", err)
	}

	if err := os.WriteFile(sourcePath, loadSignedTestRPM(t), 0600); err != nil {
		t.Fatalf("rewrite original OBShell RPM path: %v", err)
	}
	installPath := t.TempDir()
	digest, err := installRpmFileToTargetDir(snapshot, installPath, &expectedIdentity)
	if err != nil {
		t.Fatalf("install from private OBShell RPM snapshot: %v", err)
	}
	obshellPath := filepath.Join(installPath, "home/admin/oceanbase/bin/obshell")
	if _, err := os.Stat(obshellPath); err != nil {
		t.Fatalf("stat OBShell extracted from private snapshot: %v", err)
	}
	contents, err := os.ReadFile(obshellPath)
	if err != nil {
		t.Fatalf("read OBShell extracted from private snapshot: %v", err)
	}
	expectedDigest := sha256.Sum256(contents)
	if digest != hex.EncodeToString(expectedDigest[:]) {
		t.Fatalf("signed OBShell digest mismatch: got %s, want %x", digest, expectedDigest)
	}
}

func TestInstallRejectsUnexpectedPackageName(t *testing.T) {
	rpmPath := filepath.Join(t.TempDir(), "expected-obshell.rpm")
	if err := os.WriteFile(rpmPath, loadSignedTestRPM(t), 0600); err != nil {
		t.Fatalf("write non-OBShell RPM fixture: %v", err)
	}
	expectedIdentity := readRpmIdentity(t, rpmPath)
	expectedIdentity.Name = "obshell"
	if _, err := installVerifiedObshellRpmToTargetDir(rpmPath, t.TempDir(), expectedIdentity); err == nil {
		t.Fatal("expected OBShell install to reject an RPM with a different package name")
	}
}

func TestInstallOceanBaseRpmDoesNotRequireSignature(t *testing.T) {
	damaged := damageRpmGPGSignature(t, loadSignedTestRPM(t))
	rpmPath := filepath.Join(t.TempDir(), "oceanbase-libs-with-damaged-signature.rpm")
	if err := os.WriteFile(rpmPath, damaged, 0600); err != nil {
		t.Fatalf("write OceanBase RPM fixture: %v", err)
	}
	if err := InstallRpmPkgToTargetDir(rpmPath, t.TempDir()); err != nil {
		t.Fatalf("OceanBase RPM unexpectedly required a valid signature: %v", err)
	}
}

func readRpmIdentity(t *testing.T, path string) RpmIdentity {
	t.Helper()
	input, err := os.Open(path)
	if err != nil {
		t.Fatalf("open RPM identity fixture: %v", err)
	}
	defer func() {
		if err := input.Close(); err != nil {
			t.Errorf("close RPM identity fixture: %v", err)
		}
	}()

	rpmPkg, err := ReadRpm(input)
	if err != nil {
		t.Fatalf("read RPM identity fixture: %v", err)
	}
	return RpmIdentity{
		Name:         rpmPkg.Name(),
		Version:      rpmPkg.Version(),
		Release:      rpmPkg.Release(),
		Architecture: rpmPkg.Architecture(),
	}
}

type multipartFile struct {
	bytes.Reader
}

func (*multipartFile) Close() error { return nil }

func loadSignedTestRPM(t *testing.T) []byte {
	t.Helper()
	content, err := base64.StdEncoding.DecodeString(string(testSignedRPMBase64))
	if err != nil {
		t.Fatalf("decode signed RPM fixture: %v", err)
	}
	return content
}

func damageRpmGPGSignature(t *testing.T, content []byte) []byte {
	t.Helper()
	const (
		rpmLeadSize          = 96
		headerPrefixSize     = 16
		headerIndexEntrySize = 16
		rpmSigTagPGP         = 1002
		rpmSigTagPGP5        = 1006
		rpmSigTagGPG         = 1005
	)

	damaged := bytes.Clone(content)
	if len(damaged) < rpmLeadSize+headerPrefixSize {
		t.Fatal("RPM fixture is too short to contain a signature header")
	}
	headerStart := rpmLeadSize
	indexCount := int(binary.BigEndian.Uint32(damaged[headerStart+8 : headerStart+12]))
	dataLength := int(binary.BigEndian.Uint32(damaged[headerStart+12 : headerStart+16]))
	indexStart := headerStart + headerPrefixSize
	dataStart := indexStart + indexCount*headerIndexEntrySize
	if indexCount <= 0 || dataLength <= 0 || dataStart+dataLength > len(damaged) {
		t.Fatal("RPM fixture has an invalid signature header layout")
	}

	for index := 0; index < indexCount; index++ {
		entryStart := indexStart + index*headerIndexEntrySize
		tag := binary.BigEndian.Uint32(damaged[entryStart : entryStart+4])
		if tag != rpmSigTagPGP && tag != rpmSigTagPGP5 && tag != rpmSigTagGPG {
			continue
		}
		offset := int(binary.BigEndian.Uint32(damaged[entryStart+8 : entryStart+12]))
		count := int(binary.BigEndian.Uint32(damaged[entryStart+12 : entryStart+16]))
		if offset < 0 || count <= 0 || offset+count > dataLength {
			t.Fatal("RPM fixture has an invalid GPG signature entry")
		}
		damaged[dataStart+offset+count/2] ^= 0xff
		return damaged
	}

	t.Fatal("RPM fixture does not contain a GPG signature entry")
	return nil
}

func removeRpmGPGSignatureTags(t *testing.T, content []byte) []byte {
	t.Helper()
	const (
		rpmLeadSize          = 96
		headerPrefixSize     = 16
		headerIndexEntrySize = 16
		unusedSignatureTag   = 1999
	)
	gpgTags := map[uint32]struct{}{
		1002: {}, // RPMSIGTAG_PGP
		1005: {}, // RPMSIGTAG_GPG
		1006: {}, // RPMSIGTAG_PGP5
	}

	unsigned := bytes.Clone(content)
	if len(unsigned) < rpmLeadSize+headerPrefixSize {
		t.Fatal("RPM fixture is too short to contain a signature header")
	}
	headerStart := rpmLeadSize
	indexCount := int(binary.BigEndian.Uint32(unsigned[headerStart+8 : headerStart+12]))
	indexStart := headerStart + headerPrefixSize
	if indexCount <= 0 || indexStart+indexCount*headerIndexEntrySize > len(unsigned) {
		t.Fatal("RPM fixture has an invalid signature index")
	}

	replaced := 0
	for index := 0; index < indexCount; index++ {
		entryStart := indexStart + index*headerIndexEntrySize
		tag := binary.BigEndian.Uint32(unsigned[entryStart : entryStart+4])
		if _, ok := gpgTags[tag]; !ok {
			continue
		}
		binary.BigEndian.PutUint32(unsigned[entryStart:entryStart+4], unusedSignatureTag+uint32(replaced))
		replaced++
	}
	if replaced == 0 {
		t.Fatal("RPM fixture does not contain a GPG signature tag")
	}
	return unsigned
}
