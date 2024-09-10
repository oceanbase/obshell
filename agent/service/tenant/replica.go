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
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	bo "github.com/oceanbase/obshell/agent/repository/model/bo"
	model "github.com/oceanbase/obshell/agent/repository/model/oceanbase"
)

func (t *TenantService) GetTenantLocality(tenantId int) (locality string, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return "", err
	}
	err = db.Table(DBA_OB_TENANTS).Select("LOCALITY").Where("TENANT_ID = ?", tenantId).Scan(&locality).Error
	return
}

func (t *TenantService) AlterTenantLocality(tenantId int, locality string) error {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return err
	}
	tenantName, err := t.GetTenantName(tenantId)
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(SQL_ALTER_TENANT_LOCALITY, tenantName, transfer(locality))).Error
}

func (t *TenantService) GetTenantJobStatus(jobId int) (status string, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return "", err
	}
	err = db.Table(DBA_OB_TENANT_JOBS).Select("JOB_STATUS").Where("JOB_ID = ?", jobId).Scan(&status).Error
	return
}

func (t *TenantService) GetTargetTenantJob(jobType string, tenantId int, sqlText string) (id int, err error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return 0, err
	}
	err = db.Table(DBA_OB_TENANT_JOBS).
		Select("JOB_ID").Where("JOB_TYPE = ? AND SQL_TEXT like ? AND TENANT_ID = (?)", jobType, sqlText, tenantId).
		Order("JOB_ID DESC").
		Limit(1).
		Scan(&id).Error
	return
}

func (t *TenantService) GetInProgressTenantJobBo(jobType string, tenantId int) (*bo.DbaObTenantJobBo, error) {
	db, err := oceanbase.GetInstance()
	if err != nil {
		return nil, err
	}
	var job *model.DbaObTenantJob
	if err = db.Table(DBA_OB_TENANT_JOBS).
		Select("JOB_ID").
		Where("JOB_TYPE = ? AND JOB_STATUS = 'INPROGRESS' AND TENANT_ID = (?)", jobType, tenantId).
		Scan(&job).Error; err != nil {
		return nil, err
	}
	if job == nil {
		return nil, nil
	}
	return convertDbaObTenantJobToTenantJobBo(job)
}

// ParsePrimaryZone will parse the primary zone string to a slice of different priority zones.
// every priority zone will be a string with zones separated by comma and sorted.
func ParsePrimaryZone(primaryZone string) (res []string) {
	if primaryZone == constant.PRIMARY_ZONE_RANDOM {
		return []string{constant.PRIMARY_ZONE_RANDOM}
	}
	zonesSemicolonSeparated := strings.Split(primaryZone, ";")
	for _, zones := range zonesSemicolonSeparated {
		zonesCommaSeparated := strings.Split(zones, ",")
		sort.Slice(zonesCommaSeparated, func(i, j int) bool {
			return zonesCommaSeparated[i] < zonesCommaSeparated[j]
		})
		res = append(res, strings.Join(zonesCommaSeparated, ","))
	}
	return
}

func convertDbaObTenantJobToTenantJobBo(job *model.DbaObTenantJob) (*bo.DbaObTenantJobBo, error) {
	jobBo := &bo.DbaObTenantJobBo{
		JobId:     job.JobId,
		JobType:   job.JobType,
		JobStatus: job.JobStatus,
		TenantId:  job.TenantId,
		ExtraInfo: job.ExtraInfo,
	}
	pattern := `TO: '([^']+)'`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(job.ExtraInfo)
	if match == nil {
		return nil, errors.New("unexpect error: the target locality of alter locality job is unexpect.")
	}
	to := match[1]

	var err error
	if job.JobType == constant.ALTER_TENANT_LOCALITY {
		jobBo.CurrentTarget, err = parseLocalityToReplicaInfoMap(to)
		if err != nil {
			return nil, err
		}
	} else if job.JobType == constant.ALTER_TENANT_PRIMARY_ZONE {
		jobBo.CurrentTarget = ParsePrimaryZone(to)
	} else if job.JobType == constant.ALTER_RESOURCE_TENANT_UNIT_NUM {
		jobBo.CurrentTarget, err = strconv.Atoi(to)
		if err != nil {
			return nil, errors.New("unexpect error: the target unit num of alter unit num job is unexpect.")
		}
	}
	return jobBo, nil
}
