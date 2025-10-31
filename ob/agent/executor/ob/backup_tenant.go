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

package ob

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

func PostTenantBackupConfig(tenantName string, p *param.TenantBackupConfigParam) (*task.DagDetailDTO, error) {
	if p.DataBaseUri == nil || *p.DataBaseUri == "" {
		return nil, errors.Occur(errors.ErrObBackupDataBaseUriEmpty)
	}
	if p.ArchiveBaseUri == nil || *p.ArchiveBaseUri == "" {
		return nil, errors.Occur(errors.ErrObBackupArchiveBaseUriEmpty)
	}
	return tenantBackupConfig(tenantName, p)
}

func PatchTenantBackupConfig(tenant *oceanbase.DbaObTenant, p *param.TenantBackupConfigParam) (*task.DagDetailDTO, error) {
	if p.ArchiveBaseUri == nil || *p.ArchiveBaseUri == "" {
		if p.Binding != nil && *p.Binding != "" {
			return nil, errors.OccurWithMessage("If binding is set, archive_base_uri must be set", errors.ErrObBackupArchiveBaseUriEmpty)
		}
		if p.PieceSwitchInterval != nil && *p.PieceSwitchInterval != "" {
			return nil, errors.OccurWithMessage("If piece_switch_interval is set, archive_base_uri must be set", errors.ErrObBackupArchiveBaseUriEmpty)
		}
	}
	return tenantBackupConfig(tenant.TenantName, p)
}

func tenantBackupConfig(tenantName string, tenantP *param.TenantBackupConfigParam) (*task.DagDetailDTO, error) {
	log.Infof("Start to set backup config for %s", tenantName)
	p := param.NewBackupConfigParamForTenant(tenantP)
	backupConf, err := p.Check()
	if err != nil {
		return nil, err
	}

	template := buildSetBackupConfigTemplate(&tenantName)
	ctx, err := buildSetBackupConfigTaskContext(backupConf, &tenantName)
	if err != nil {
		return nil, err
	}

	dag, err := taskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func GetTenantBackupConfig(tenantName string) (*bo.BackupDestInfo, error) {
	tenant, err := tenantService.GetTenantByName(tenantName)
	if err != nil {
		return nil, err
	}

	var configs bo.BackupDestInfo
	configs.TenantID = tenant.TenantID
	archiveBaseUri, err := tenantService.GetArchiveDestByID(tenant.TenantID)
	if err != nil {
		return nil, err
	}

	if archiveBaseUri != "" {
		archiveStorage, err := system.GetStorageInterfaceByURI(archiveBaseUri)
		if err != nil {
			return nil, err
		}
		configs.ArchiveBaseUri = archiveStorage.GenerateURIWithoutSecret()
	}

	dataBaseUri, err := tenantService.GetDataBackupDestByID(tenant.TenantID)
	if err != nil {
		return nil, err
	}
	if dataBaseUri != "" {
		dataStorage, err := system.GetStorageInterfaceByURI(dataBaseUri)
		if err != nil {
			return nil, err
		}
		configs.DataBaseUri = dataStorage.GenerateURIWithoutSecret()
	}

	if strings.Contains(configs.ArchiveBaseUri, "access_id") || strings.Contains(configs.ArchiveBaseUri, "access_key") ||
		strings.Contains(configs.DataBaseUri, "access_id") || strings.Contains(configs.DataBaseUri, "access_key") {
		return nil, nil
	}

	return &configs, nil
}

func TenantStartBackup(tenant *oceanbase.DbaObTenant, p *param.BackupParam) (*task.DagDetailDTO, error) {
	if err := checkAllDest(tenant); err != nil {
		return nil, err
	}
	ctx, err := buildStartTenantBackupCtx(p, tenant.TenantName)
	if err != nil {
		return nil, err
	}

	template := buildStartTenantBackupTemplate(*p.Mode, tenant.TenantName)

	dag, err := taskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}
