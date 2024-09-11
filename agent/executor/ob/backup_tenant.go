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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/param"
)

func PostTenantBackupConfig(tenantName string, p *param.TenantBackupConfigParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if p.DataBaseUri == nil || *p.DataBaseUri == "" || p.ArchiveBaseUri == nil || *p.ArchiveBaseUri == "" {
		return nil, errors.Occur(errors.ErrIllegalArgument, errors.New("data_base_uri or archive_base_uri cannot be empty"))
	}
	return tenantBackupConfig(tenantName, p)
}

func PatchTenantBackupConfig(tenant *oceanbase.DbaObTenant, p *param.TenantBackupConfigParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if p.ArchiveBaseUri == nil || *p.ArchiveBaseUri == "" {
		if p.Binding != nil && *p.Binding != "" {
			return nil, errors.Occur(errors.ErrIllegalArgument, errors.New("If binding is set, archive_base_uri must be set"))
		}
		if p.PieceSwitchInterval != nil && *p.PieceSwitchInterval != "" {
			return nil, errors.Occur(errors.ErrIllegalArgument, errors.New("If piece_switch_interval is set, archive_base_uri must be set"))
		}
	}
	return tenantBackupConfig(tenant.TenantName, p)
}

func tenantBackupConfig(tenantName string, tenantP *param.TenantBackupConfigParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	log.Infof("Start to set backup config for %s", tenantName)
	p := param.NewBackupConfigParamForTenant(tenantP)
	backupConf, err := p.Check()
	if err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err, "backup config is invalid")
	}

	template := buildSetBackupConfigTemplate(&tenantName)
	ctx, err := buildSetBackupConfigTaskContext(backupConf, &tenantName)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	dag, err := taskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

func TenantStartBackup(tenant *oceanbase.DbaObTenant, p *param.BackupParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if err := checkAllDest(tenant); err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err)
	}
	ctx, err := buildStartTenantBackupCtx(p, tenant.TenantName)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	template := buildStartTenantBackupTemplate(*p.Mode, tenant.TenantName)

	dag, err := taskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}
