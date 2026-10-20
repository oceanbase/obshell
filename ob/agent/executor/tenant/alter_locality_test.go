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
	"reflect"
	"testing"
)

func TestTenantJobSQLTextPatternsEscapeLiteralMetacharacters(t *testing.T) {
	got := tenantJobSQLTextPatterns("F@zone_1!,R@zone%2", '"')
	want := []string{
		"%\"F@zone!_1!!,R@zone!%2\"",
		"%'F@zone!_1!!,R@zone!%2'",
		"%'F@zone!_1!!,R@zone!%2'",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SQL patterns = %#v, want %#v", got, want)
	}
}

func TestTenantJobSQLTextPatternsSupportSQLModeAndLegacyQuotes(t *testing.T) {
	got := tenantJobSQLTextPatterns("F@z\\a'b\"c", '"')
	want := []string{
		"%\"F@z\\\\a'b\\\"c\"",
		"%'F@z\\\\a''b\"c'",
		"%'F@z\\a''b\"c'",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SQL mode patterns = %#v, want %#v", got, want)
	}
	primary := tenantJobSQLTextPatterns("zone_1", '`')
	if primary[0] != "%`zone!_1`" || primary[1] != "%'zone!_1'" {
		t.Fatalf("primary zone patterns = %#v", primary)
	}
}
