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

package agent

import (
	"testing"

	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQueryObSysParametersUsesCompatibleView(t *testing.T) {
	db, err := gorm.Open(gormsqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite database: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	statements := []string{
		`ATTACH DATABASE ':memory:' AS oceanbase`,
		`CREATE TABLE oceanbase."V$OB_PARAMETERS" (
			NAME TEXT, DATA_TYPE TEXT, VALUE TEXT, INFO TEXT, SECTION TEXT,
			EDIT_LEVEL TEXT, DEFAULT_VALUE TEXT, ISDEFAULT TEXT
		)`,
		`INSERT INTO oceanbase."V$OB_PARAMETERS"
			(NAME, DATA_TYPE, VALUE, INFO, SECTION, EDIT_LEVEL, DEFAULT_VALUE, ISDEFAULT)
		VALUES
			('enabled_parameter', 'BOOL', 'True', 'enabled info', 'OBSERVER', 'DYNAMIC_EFFECTIVE', 'False', 'YES'),
			('disabled_parameter', 'BOOL', 'False', 'disabled info', 'OBSERVER', 'DYNAMIC_EFFECTIVE', 'False', 'NO')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("prepare compatible parameter view: %v", err)
		}
	}

	parameters, err := queryObSysParameters(db)
	if err != nil {
		t.Fatalf("query compatible parameter view: %v", err)
	}
	if len(parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(parameters))
	}

	byName := make(map[string]bool, len(parameters))
	for _, parameter := range parameters {
		byName[parameter.Name] = parameter.IsDefault
	}
	if !byName["enabled_parameter"] {
		t.Error("expected YES to be converted to true")
	}
	if byName["disabled_parameter"] {
		t.Error("expected NO to be converted to false")
	}
}
