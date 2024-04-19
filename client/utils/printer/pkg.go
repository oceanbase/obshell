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

package printer

import (
	"sort"
	"strings"

	"github.com/cavaliergopher/rpm"

	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/agent/lib/pkg"
)

// RpmSlice is a slice of *rpm.Package that can be sorted.
type RpmSlice []*PkgInfo

// Len is the number of elements in the collection.
func (r RpmSlice) Len() int {
	return len(r)
}

// Less reports whether the element with index i should sort before the element with index j.
func (r RpmSlice) Less(i, j int) bool {
	if r[i].PkgInfo.Version() == r[j].PkgInfo.Version() {
		if pkg.CompareVersion(r[i].PkgInfo.Release(), r[j].PkgInfo.Release()) == 0 {
			return len(r[i].PkgInfo.Name()) < len(r[j].PkgInfo.Name())
		}
		return pkg.CompareVersion(r[i].PkgInfo.Release(), r[j].PkgInfo.Release()) < 0
	}
	return r[i].PkgInfo.Version() < r[j].PkgInfo.Version()
}

// Swap swaps the elements with indexes i and j.
func (r RpmSlice) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type PkgInfo struct {
	FileName string
	PkgInfo  *rpm.Package
}

func PrintPkgsTable(pkgs map[string]*rpm.Package) {
	header := []string{"File Name", "Name", "Version", "Build Number", "Distribution"}
	rows := make([][]string, 0, len(pkgs))

	// Convert map to slice
	rpmSlice := make(RpmSlice, 0, len(pkgs))
	for name, pkg := range pkgs {
		rpmSlice = append(rpmSlice, &PkgInfo{
			FileName: name,
			PkgInfo:  pkg,
		})

	}
	sort.Sort(rpmSlice)

	// Print the sorted slice
	for _, pkg := range rpmSlice {
		items := strings.Split(pkg.PkgInfo.Release(), ".")
		rows = append(rows, []string{
			pkg.FileName,
			pkg.PkgInfo.Name(),
			pkg.PkgInfo.Version(),
			items[0],
			items[1],
		})
	}
	stdio.PrintTable(header, rows)
}
