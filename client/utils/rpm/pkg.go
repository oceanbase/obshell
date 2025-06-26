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
	"os"
	"path/filepath"
	"strings"

	"github.com/cavaliergopher/rpm"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/client/lib/stdio"
)

func GetAllRpmsInDirByName(pkgDir, pkgName string) (pkgs map[string]*rpm.Package, err error) {
	pkgs = make(map[string]*rpm.Package)
	if _, err = os.Stat(pkgDir); err != nil {
		return nil, err
	}
	dirEntries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, err
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() || !(strings.HasSuffix(dirEntry.Name(), ".rpm") &&
			strings.HasPrefix(dirEntry.Name(), pkgName) &&
			strings.Contains(dirEntry.Name(), constant.DIST)) {
			continue
		}

		pkg, err := checkFileName(filepath.Join(pkgDir, dirEntry.Name()), pkgName)
		if err != nil {
			continue
		}
		stdio.Verbosef("rpm package %s found", dirEntry.Name())
		pkgs[dirEntry.Name()] = pkg
	}

	if len(pkgs) == 0 {
		return nil, errors.Occur(errors.ErrCliUpgradePackageNotFoundInPath, pkgName, pkgDir)
	}
	return pkgs, nil
}

func GetAllRpmsInDirByNames(pkgDir string, pkgNames []string) (pkgs map[string]*rpm.Package, err error) {
	pkgs = make(map[string]*rpm.Package)
	for _, pkgName := range pkgNames {
		stdio.Verbosef("searching rpm package %s", pkgName)
		p, err := GetAllRpmsInDirByName(pkgDir, pkgName)
		if err != nil {
			continue
		}
		for k, v := range p {
			pkgs[k] = v
		}
	}
	return pkgs, nil
}

func checkFileName(path, name string) (*rpm.Package, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	pkg, err := rpm.Read(file)
	if err != nil {
		return nil, err
	}
	if pkg.Name() != name {
		return nil, errors.Occur(errors.ErrPackageNameMismatch, pkg.Name(), name)
	}
	pkg.Version()
	items := strings.Split(pkg.Release(), ".")
	if len(items) != 2 {
		return nil, errors.Occur(errors.ErrPackageReleaseInvalid, pkg.Release())
	}
	return pkg, nil
}
