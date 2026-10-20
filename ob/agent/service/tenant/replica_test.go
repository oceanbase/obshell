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

package tenant

import (
	"testing"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func tenantJobTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.Exec("ATTACH DATABASE ':memory:' AS oceanbase").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE oceanbase.DBA_OB_TENANT_JOBS (JOB_ID INTEGER, JOB_TYPE TEXT, TENANT_ID INTEGER, SQL_TEXT TEXT)").Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestTargetTenantJobFindsNewestAcrossQuoteFormats(t *testing.T) {
	db := tenantJobTestDB(t)
	for _, row := range []struct {
		id   int
		text string
	}{
		{1, "ALTER TENANT `test` LOCALITY = \"F@zone1\""},
		{2, "ALTER TENANT `test` LOCALITY = 'F@zone1'"},
	} {
		if err := db.Exec("INSERT INTO oceanbase.DBA_OB_TENANT_JOBS VALUES (?, ?, ?, ?)", row.id, constant.ALTER_TENANT_LOCALITY, 1002, row.text).Error; err != nil {
			t.Fatal(err)
		}
	}
	var id int
	if err := targetTenantJobQuery(db, constant.ALTER_TENANT_LOCALITY, 1002, "%\"F@zone1\"", "%'F@zone1'").Scan(&id).Error; err != nil {
		t.Fatal(err)
	}
	if id != 2 {
		t.Fatalf("matched job %d, want newest single-quoted job 2", id)
	}
}

func TestTargetTenantJobKeepsTenantTypeAndLiteralTargetBoundaries(t *testing.T) {
	db := tenantJobTestDB(t)
	rows := []struct {
		id, tenant int
		kind, text string
	}{
		{1, 1002, constant.ALTER_TENANT_LOCALITY, "ALTER TENANT `test` LOCALITY = 'F@zone_1'"},
		{99, 1003, constant.ALTER_TENANT_LOCALITY, "ALTER TENANT `other` LOCALITY = 'F@zone_1'"},
		{98, 1002, constant.ALTER_TENANT_PRIMARY_ZONE, "ALTER TENANT `test` PRIMARY_ZONE = 'F@zone_1'"},
		{97, 1002, constant.ALTER_TENANT_LOCALITY, "ALTER TENANT `test` LOCALITY = 'F@zoneX1'"},
		{96, 1002, constant.ALTER_TENANT_LOCALITY, "ALTER TENANT `test` LOCALITY = 'F@zone_1_suffix'"},
	}
	for _, row := range rows {
		if err := db.Exec("INSERT INTO oceanbase.DBA_OB_TENANT_JOBS VALUES (?, ?, ?, ?)", row.id, row.kind, row.tenant, row.text).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, patterns := range [][]string{
		{"%'F@zone!_1'"},
		{"%\"F@zone!_1\"", "%'F@zone!_1'"},
	} {
		var id int
		if err := targetTenantJobQuery(db, constant.ALTER_TENANT_LOCALITY, 1002, patterns...).Scan(&id).Error; err != nil {
			t.Fatal(err)
		}
		if id != 1 {
			t.Fatalf("matched job %d, want exact target job 1 for patterns %#v", id, patterns)
		}
	}
}

func TestTargetTenantJobPrimaryZoneQuoteCompatibility(t *testing.T) {
	db := tenantJobTestDB(t)
	for _, row := range []struct {
		id   int
		text string
	}{
		{1, "ALTER TENANT `test` PRIMARY_ZONE = `zone1`"},
		{2, "ALTER TENANT `test` PRIMARY_ZONE = 'zone1'"},
	} {
		if err := db.Exec("INSERT INTO oceanbase.DBA_OB_TENANT_JOBS VALUES (?, ?, ?, ?)", row.id, constant.ALTER_TENANT_PRIMARY_ZONE, 1002, row.text).Error; err != nil {
			t.Fatal(err)
		}
	}
	var id int
	if err := targetTenantJobQuery(db, constant.ALTER_TENANT_PRIMARY_ZONE, 1002, "%`zone1`", "%'zone1'").Scan(&id).Error; err != nil {
		t.Fatal(err)
	}
	if id != 2 {
		t.Fatalf("matched primary zone job %d, want 2", id)
	}
}

func TestTargetTenantJobNoMatchAndDatabaseErrors(t *testing.T) {
	db := tenantJobTestDB(t)
	var id int
	if err := targetTenantJobQuery(db, constant.ALTER_TENANT_LOCALITY, 1002, "%'missing'").Scan(&id).Error; err != nil {
		t.Fatal(err)
	}
	if id != 0 {
		t.Fatalf("missing job = %d, want 0", id)
	}
	if err := targetTenantJobQuery(db, constant.ALTER_TENANT_LOCALITY, 1002).Scan(&id).Error; err == nil {
		t.Fatal("empty patterns must return an error, not select an arbitrary job")
	}
	if err := db.Exec("DROP TABLE oceanbase.DBA_OB_TENANT_JOBS").Error; err != nil {
		t.Fatal(err)
	}
	if err := targetTenantJobQuery(db, constant.ALTER_TENANT_LOCALITY, 1002, "%'missing'").Scan(&id).Error; err == nil {
		t.Fatal("database query error was lost")
	}
}
