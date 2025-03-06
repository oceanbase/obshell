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

package task

type Maintainer interface {
	IsMaintenance() bool
	GetMaintenanceType() int
	GetMaintenanceKey() string
}

type maintenance struct {
	maintenanceType int
	maintenanceKey  string
}

func (m *maintenance) IsMaintenance() bool {
	return m.maintenanceType != NOT_UNDER_MAINTENANCE
}

func (m *maintenance) GetMaintenanceType() int {
	return m.maintenanceType
}

func (m *maintenance) GetMaintenanceKey() string {
	return m.maintenanceKey
}

const (
	NOT_BOOTSTRAP = iota
	NOT_UNDER_MAINTENANCE
	GLOBAL_MAINTENANCE
	TENANT_MAINTENANCE
	OBPROXY_MAINTENACE
)

func UnMaintenance() Maintainer {
	return &maintenance{
		maintenanceType: NOT_UNDER_MAINTENANCE,
	}
}

func GlobalMaintenance() Maintainer {
	return &maintenance{
		maintenanceType: GLOBAL_MAINTENANCE,
	}
}

func TenantMaintenance(tenantName string) Maintainer {
	return &maintenance{
		maintenanceType: TENANT_MAINTENANCE,
		maintenanceKey:  tenantName,
	}
}

func ObproxyMaintenance() Maintainer {
	return &maintenance{
		maintenanceType: OBPROXY_MAINTENACE,
	}
}

func NewMaintenance(maintenanceType int, maintenanceKey string) Maintainer {
	return &maintenance{
		maintenanceType: maintenanceType,
		maintenanceKey:  maintenanceKey,
	}
}

func mergeMaintainers(maintainers ...Maintainer) Maintainer {
	if len(maintainers) == 0 {
		return nil
	}

	maintainer := UnMaintenance()
	for _, m := range maintainers {
		switch maintainer.GetMaintenanceType() {
		case NOT_UNDER_MAINTENANCE:
			if m.GetMaintenanceType() != NOT_UNDER_MAINTENANCE {
				maintainer = m
			}
		case GLOBAL_MAINTENANCE:
			continue
		default:
			if m.GetMaintenanceType() != maintainer.GetMaintenanceType() {
				maintainer = GlobalMaintenance()
			} else if m.GetMaintenanceKey() != maintainer.GetMaintenanceKey() {
				maintainer = GlobalMaintenance()
			}
		}
	}

	return maintainer
}
