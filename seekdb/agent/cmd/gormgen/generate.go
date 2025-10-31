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

package main

import (
	"path/filepath"

	"gorm.io/gen"

	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/sqlite"
)

const REPO_ROOT = "../../repository/query"

func main() {
	GenerateSqlite()
	GenerateOceanbase()
}

func GenerateSqlite() {
	g := gen.NewGenerator(gen.Config{
		Mode:    gen.WithDefaultQuery,
		OutPath: filepath.Join(REPO_ROOT, "sqlite"),
	})
	g.ApplyBasic(sqlite.OcsInfo{},
		sqlite.ObSysParameter{},
		sqlite.ObConfig{}, sqlite.OcsConfig{})
	g.Execute()
}

func GenerateOceanbase() {
	g := gen.NewGenerator(gen.Config{
		Mode:    gen.WithDefaultQuery,
		OutPath: filepath.Join(REPO_ROOT, "oceanbase"),
	})
	g.ApplyBasic(oceanbase.AllAgent{},
		oceanbase.DagInstance{}, oceanbase.NodeInstance{}, oceanbase.SubtaskInstance{},
		oceanbase.UpgradePkgInfo{}, oceanbase.UpgradePkgChunk{},
	)
	g.Execute()
}
