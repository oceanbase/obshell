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
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/pool"
	"github.com/oceanbase/obshell/agent/executor/script"
	"github.com/oceanbase/obshell/agent/executor/zone"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/meta"
	tenantservice "github.com/oceanbase/obshell/agent/service/tenant"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
)

func checkCharsetAndCollation(charset string, collation string) (err error) {
	if charset == "" && collation == "" {
		return nil
	}

	res, err := obclusterService.GetCharsetAndCollation(charset, collation)
	if err != nil {
		return err
	}
	if res == nil {
		if charset != "" && collation != "" {
			return errors.Errorf("Charset '%s' and Collation '%s' is not match.", charset, collation)
		} else if charset != "" {
			return errors.Errorf("Charset '%s' is not exist.", charset)
		} else {
			return errors.Errorf("Collation '%s' is not exist.", collation)
		}
	}

	return nil
}

// transferNumber transfer float64(Scientific Notation) to int64 or float64
func transferNumber(mp map[string]interface{}) {
	for key, value := range mp {
		if number, ok := value.(float64); ok {
			if number == float64(int(number)) {
				mp[key] = int(number)
			} else {
				mp[key] = float64(number)
			}
		}
	}
}

func checkVariables(vars map[string]interface{}) error {
	for k, v := range vars {
		if k == "" || v == nil {
			return errors.New("Variable name and value should not be empty.")
		}
	}

	transferNumber(vars)
	return tenantService.CheckVariables(vars)
}

func checkParameters(parameters map[string]interface{}) error {
	for k, v := range parameters {
		if k == "" || v == nil {
			return errors.New("Variable name and value should not be empty.")
		}
	}

	for k := range parameters {
		// Check whether the parameter is exist.
		if param, err := tenantService.GetTenantParameter(constant.TENANT_SYS_ID, k); err != nil {
			return errors.Wrap(err, "Get tenant parameter failed.")
		} else if param == nil {
			return errors.Errorf("Parameter '%s' is not exist.", k)
		}
	}

	return nil
}

func checkAndLoadScenario(param *param.CreateTenantParam, scenario string) error {
	if scenario == "" {
		return nil
	}

	scenarios := getAllSupportedScenarios()
	if len(scenarios) == 0 {
		return errors.New("current observer does not support scenario")
	}
	if !utils.ContainsString(scenarios, strings.ToLower(scenario)) {
		errors.Errorf("scenario only support to be one of %s", strings.Join(scenarios, ", "))
	}

	variables, err := parseTemplate(VARIABLES_TEMPLATE, path.ObshellDefaultVariablePath(), scenario)
	if err != nil {
		return errors.Wrap(err, "Parse variable template failed")
	}
	for key, value := range variables {
		if _, exist := param.Variables[key]; !exist {
			param.Variables[key] = value
		}
	}

	parameters, err := parseTemplate(PARAMETERS_TEMPLATE, path.ObshellDefaultParameterPath(), scenario)
	if err != nil {
		return errors.Wrap(err, "Parse parameter template failed")
	}
	for key, value := range parameters {
		if _, exist := param.Parameters[key]; !exist {
			param.Parameters[key] = value
		}
	}
	return nil
}

func renderCreateTenantParam(param *param.CreateTenantParam) error {
	if param.PrimaryZone == "" {
		param.PrimaryZone = constant.PRIMARY_ZONE_RANDOM
	}
	if strings.ToUpper(param.PrimaryZone) == constant.PRIMARY_ZONE_RANDOM {
		param.PrimaryZone = constant.PRIMARY_ZONE_RANDOM
	}
	if param.Mode == "" {
		param.Mode = constant.MYSQL_MODE
	} else {
		param.Mode = strings.ToUpper(param.Mode)
	}

	if param.Whitelist == nil {
		var whitelist string
		param.Whitelist = &whitelist
		if param.Variables != nil {
			if value, exist := param.Variables[constant.VARIABLE_OB_TCP_INVITED_NODES]; exist {
				if tcp, ok := value.(string); ok {
					param.Whitelist = &tcp
				} else {
					return errors.New("Incorrect argument type to variable 'ob_tcp_invited_nodes'")
				}
			}
		}
	}
	delete(param.Variables, constant.VARIABLE_OB_TCP_INVITED_NODES)

	if value, exist := param.Variables[constant.VARIABLE_TIME_ZONE]; exist {
		if timeZone, ok := value.(string); ok {
			param.TimeZone = timeZone
		} else {
			return errors.New("Incorrect argument type to variable 'time_zone'")
		}
		delete(param.Variables, constant.VARIABLE_TIME_ZONE)
	}

	zone.RenderZoneParams(param.ZoneList)
	return nil
}

func checkCreateTenantParam(param *param.CreateTenantParam) (err error) {
	if len(param.ZoneList) == 0 {
		return errors.New("zone_list is empty")
	}

	if param.Mode != constant.MYSQL_MODE {
		return errors.New("only support mysql mode")
	}

	if err = zone.CheckZoneParams(param.ZoneList); err != nil {
		return
	}

	if err = zone.CheckAtLeastOnePaxosReplica(param.ZoneList); err != nil {
		return
	}

	zoneList := make([]string, 0)
	for _, zone := range param.ZoneList {
		zoneList = append(zoneList, zone.Name)
	}
	if err = zone.CheckPrimaryZone(param.PrimaryZone, zoneList); err != nil {
		return
	}

	if err = checkCharsetAndCollation(param.Charset, param.Collation); err != nil {
		return
	}

	locality := make(map[string]string, 0)
	for _, zone := range param.ZoneList {
		locality[zone.Name] = zone.ReplicaType
	}
	if err = zone.CheckPrimaryZoneAndLocality(param.PrimaryZone, locality); err != nil {
		return
	}

	if err = checkVariables(param.Variables); err != nil {
		return
	}

	if err = checkParameters(param.Parameters); err != nil {
		return
	}

	if err = checkAndLoadScenario(param, param.Scenario); err != nil {
		return
	}

	return nil
}

func checkTenantName(name string) error {
	if name == "" {
		return errors.New("Tenant name should not be empty.")
	}
	if strings.Contains(name, "$") {
		return errors.New("since 4.2.1, manually creating a tenant name containing '$' is not supported")
	}
	if name == constant.TENANT_ALL || name == constant.TENANT_ALL_META || name == constant.TENANT_ALL_USER {
		return errors.Errorf("since 4.2.1, using '%s' (case insensitive) as a tenant name is not supported", name)
	}
	if !regexp.MustCompile(TENANT_NAME_PATTERN).MatchString(name) {
		return errors.New("Tenant names may only contain letters, numbers, and special characters(- _ # ~ +).")
	}
	return nil
}

func checkZoneResourceForUnit(zone string, unitName string, unitNum int) error {
	source, err := tenantService.GetObServerCapacityByZone(zone)
	if err != nil {
		return errors.New("Get servers's info failed.")
	}
	unit, err := unitService.GetUnitConfigByName(unitName)
	if err != nil {
		return errors.New("Get unit config failed.")
	}

	var validServer int
	var checkErr error
	if len(source) < unitNum {
		return errors.Errorf("The number of servers in zone '%s' is %d, less than the number of units %d.", zone, len(source), unitNum)
	}
	for _, server := range source {
		gatheredUnitInfo, err := gatherAllUnitsOnServer(server.SvrIp, server.SvrPort)
		if err != nil {
			return err
		}

		serverStr := meta.NewAgentInfo(server.SvrIp, server.SvrPort).String()
		log.Infof("server %s used resource: %v", serverStr, gatheredUnitInfo)
		if server.CpuCapacity-gatheredUnitInfo.MinCpu < unit.MinCpu ||
			server.CpuCapacityMax-gatheredUnitInfo.MaxCpu < unit.MaxCpu {
			checkErr = errors.Errorf("server %s CPU resource not enough", serverStr)
			continue
		}
		if server.MemCapacity-gatheredUnitInfo.MemorySize < unit.MemorySize {
			checkErr = errors.Errorf("server %s MEMORY_SIZE resource not enough", serverStr)
			continue
		}
		if server.LogDiskCapacity-gatheredUnitInfo.LogDiskSize < unit.LogDiskSize {
			checkErr = errors.Errorf("server %s LOG_DISK_SIZE resource not enough", serverStr)
			continue
		}
		validServer += 1
	}
	if validServer >= unitNum {
		return nil
	}
	return checkErr
}

type gatheredUnitInfo struct {
	MinCpu      float64
	MaxCpu      float64
	MemorySize  int
	LogDiskSize int
}

func gatherAllUnitsOnServer(svrIp string, svrPort int) (*gatheredUnitInfo, error) {
	units, err := obclusterService.GetObUnitsOnServer(svrIp, svrPort)
	if err != nil {
		return nil, errors.Errorf("Get all units on server %s failed.", meta.NewAgentInfo(svrIp, svrPort).String())
	}
	used := &gatheredUnitInfo{}
	for _, unit := range units {
		used.MaxCpu += unit.MaxCpu
		used.MinCpu += unit.MinCpu
		used.MemorySize += unit.MemorySize
		used.LogDiskSize += unit.LogDiskSize
	}
	return used, nil
}

func CheckResourceEnough(zoneList []param.ZoneParam) error {
	for _, zone := range zoneList {
		if err := checkZoneResourceForUnit(zone.Name, zone.UnitConfigName, zone.UnitNum); err != nil {
			return err
		}
	}
	return nil
}

func CreateTenant(param *param.CreateTenantParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if err := checkTenantName(*param.Name); err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err.Error())
	}

	if exist, err := tenantService.IsTenantExist(*param.Name); err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "check tenant '%s' exist failed", *param.Name)
	} else if exist {
		return nil, errors.Occurf(errors.ErrBadRequest, "Tenant '%s' already exists.", *param.Name)
	}

	renderCreateTenantParam(param)

	if err := checkCreateTenantParam(param); err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err.Error())
	}

	if err := CheckResourceEnough(param.ZoneList); err != nil {
		return nil, errors.Occur(errors.ErrBadRequest, err.Error())
	}

	// Create 'Create tenant' dag instance.
	template, err := buildCreateTenantDagTemplate(param)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	context := buildCreateTenantDagContext(param)
	dag, err := clusterTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "create '%s' dag instance failed: %s", DAG_CREATE_TENANT, err.Error())
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildCreateTenantDagTemplate(param *param.CreateTenantParam) (*task.Template, error) {
	createTenantNode, err := newCreateTenantNode(param)
	if err != nil {
		return nil, err
	}
	templateBuilder := task.NewTemplateBuilder(fmt.Sprintf(DAG_CREATE_TENANT, *param.Name)).
		SetMaintenance(task.TenantMaintenance(*param.Name)).
		AddNode(createTenantNode)
	if param.TimeZone != "" {
		templateBuilder.AddNode(newSetTenantTimeZoneNode(param.TimeZone))
	}
	if param.Parameters != nil && len(param.Parameters) != 0 {
		templateBuilder.AddNode(newSetTenantParameterNode(param.Parameters))
	}
	templateBuilder.AddNode(newModifyTenantWhitelistNode(*param.Whitelist))

	// Delete the read-only variables
	for k := range param.Variables {
		if utils.ContainsString(CREATE_TENANT_STATEMENT_VARIABLES, k) {
			delete(param.Variables, k)
		}
	}
	if param.Variables != nil && len(param.Variables) != 0 {
		node, err := newSetTenantVariableNode(param.Variables)
		if err != nil {
			return nil, err
		}
		templateBuilder.AddNode(node)
	}

	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}
	if param.ImportScript {
		templateBuilder.AddNode(script.NewParallelImportScriptForTenantNode(agents, false))
	}

	if param.RootPassword != "" {
		setRootPwdNode, err := newSetRootPwdNode(param.RootPassword)
		if err != nil {
			return nil, err
		}
		templateBuilder.AddNode(setRootPwdNode)
	}

	return templateBuilder.Build(), nil
}

func buildCreateTenantDagContext(param *param.CreateTenantParam) *task.TaskContext {
	context := task.NewTaskContext()
	context.SetParam(PARAM_TENANT_NAME, param.Name).
		SetParam(task.FAILURE_EXIT_MAINTENANCE, true)
	return context
}

func buildCreateResourcePoolTaskParam(tenantName string, zoneParam []param.ZoneParam, timestamp int64) []param.CreateResourcePoolTaskParam {
	createResourcePoolParams := make([]param.CreateResourcePoolTaskParam, 0)
	for _, zone := range zoneParam {
		createResourcePoolParams = append(createResourcePoolParams, param.CreateResourcePoolTaskParam{
			PoolName:       strings.Join([]string{tenantName, zone.Name, strconv.FormatInt(timestamp, 10)}, "_"),
			ZoneName:       zone.Name,
			UnitConfigName: zone.PoolParam.UnitConfigName,
			UnitNum:        zone.PoolParam.UnitNum,
		})
	}
	return createResourcePoolParams
}

// Create Tenant Task
type CreateTenantTask struct {
	task.Task
	param.CreateTenantParam
	timestamp               int64 // use for pool name
	createResourcePoolParam []param.CreateResourcePoolTaskParam
	id                      int // tenant id
}

func newCreateTenantNode(param *param.CreateTenantParam) (*task.Node, error) {
	ctx := task.NewTaskContext().
		SetParam(PARAM_CREATE_TENANT, param).
		SetParam(PARAM_TIMESTAMP, time.Now().Unix())
	return task.NewNodeWithContext(newCreateTenantTask(), false, ctx), nil
}

func newCreateTenantTask() *CreateTenantTask {
	newTask := &CreateTenantTask{
		Task: *task.NewSubTask(TASK_NAME_CREATE_TENANT),
	}
	newTask.SetCanRollback().SetCanRetry().SetCanCancel().SetCanContinue().SetCanPass()
	return newTask
}

func buildCreateTenantSql(param *param.CreateTenantParam, poolList []string) (string, []interface{}) {
	resourcePoolList := "\"" + strings.Join(poolList, "\",\"") + "\""
	sql := fmt.Sprintf(tenantservice.SQL_CREATE_TENANT_BASIC, *param.Name, resourcePoolList)

	input := make([]interface{}, 0)

	var localityList []string
	for _, zone := range param.ZoneList {
		if zone.ReplicaType == "" {
			localityList = append(localityList, strings.Join([]string{constant.REPLICA_TYPE_FULL, zone.Name}, "@"))
		} else {
			localityList = append(localityList, strings.Join([]string{zone.ReplicaType, zone.Name}, "@"))
		}
	}
	sql += ", LOCALITY = \"%s\""
	input = append(input, transfer(strings.Join(localityList, ",")))

	sql += ", PRIMARY_ZONE = `%s`"
	input = append(input, param.PrimaryZone)

	if param.Charset != "" {
		sql += ", CHARSET = %s"
		input = append(input, transfer(param.Charset))
	}

	if param.Collation != "" && param.Mode == constant.MYSQL_MODE {
		sql += ", COLLATE = \"%s\"" // STRING_VALUE
		input = append(input, transfer(param.Collation))
	}
	if param.Comment != "" {
		sql += ", COMMENT = \"%s\""
		input = append(input, transfer(param.Comment))
	}

	if param.ReadOnly {
		sql += ", READ ONLY"
	}

	sql += " SET ob_tcp_invited_nodes = `%s`"
	input = append(input, "") // set empty string for ob_tcp_invited_nodes, to avoid tenant be used before the dag is SUCCEED

	if param.Mode != "" {
		sql += ", ob_compatibility_mode = `%s`"
		input = append(input, param.Mode)
	}

	transferNumber(param.Variables)
	for k, v := range param.Variables {
		if utils.ContainsString(CREATE_TENANT_STATEMENT_VARIABLES, k) {
			if _, ok := v.(string); ok {
				sql += ", " + k + "= `%s`"
			} else {
				sql += ", " + k + "= %v"
			}
			input = append(input, v)
		}
	}

	return sql, input
}

func (t *CreateTenantTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_CREATE_TENANT, &t.CreateTenantParam); err != nil {
		return errors.Wrapf(err, "Get create tenant param failed")
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TIMESTAMP, &t.timestamp); err != nil {
		return errors.Wrapf(err, "Get timestamp failed")
	}

	t.createResourcePoolParam = buildCreateResourcePoolTaskParam(*t.CreateTenantParam.Name, t.CreateTenantParam.ZoneList, t.timestamp)
	if err := pool.CreatePools(t.Task, t.createResourcePoolParam); err != nil {
		return err
	}

	var poolList []string
	for _, poolParam := range t.createResourcePoolParam {
		poolList = append(poolList, poolParam.PoolName)
	}
	basic, input := buildCreateTenantSql(&t.CreateTenantParam, poolList)
	sql := fmt.Sprintf(basic, input...)
	t.ExecuteLogf("Create tenant sql: %s", sql)
	if err := tenantService.TryExecute(sql); err != nil {
		// drop all created resource pool
		if err := pool.DropFreeResourcePools(t.Task, t.createResourcePoolParam); err != nil {
			t.ExecuteWarnLog(errors.Wrap(err, "Drop created resource pool failed."))
		}
		return err
	}
	// get tenant id
	tenantID, err := tenantService.GetTenantId(*t.CreateTenantParam.Name)
	if err != nil {
		return err
	}
	t.ExecuteLogf("Create tenant success, tenant id: %d", tenantID)
	t.GetContext().SetParam(PARAM_TENANT_ID, tenantID)
	return nil
}

func (t *CreateTenantTask) Rollback() error {
	if err := t.GetContext().GetParamWithValue(PARAM_CREATE_TENANT, &t.CreateTenantParam); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TIMESTAMP, &t.timestamp); err != nil {
		return errors.Wrapf(err, "Get timestamp failed")
	}

	// If error, tenant id will be 0.
	t.GetContext().GetParamWithValue(PARAM_TENANT_ID, &t.id)

	if t.CreateTenantParam.Name == nil {
		return errors.Errorf("Unexpected error, tenant name is nil")
	}

	t.ExecuteLogf("Drop tenant %s if exist", *t.CreateTenantParam.Name)
	// drop tenant if exist
	if t.id != 0 {
		tenantName, err := tenantService.GetTenantName(t.id)
		if err != nil {
			return errors.New("Get tenant name failed.")
		} else if tenantName != *t.CreateTenantParam.Name {
			return errors.Errorf("Tenant name %s is not equal to %s", tenantName, *t.CreateTenantParam.Name)
		}
		if err := tenantService.DropTenant(tenantName); err != nil {
			return errors.Wrap(err, "Drop tenant failed.")
		}
	}
	t.createResourcePoolParam = buildCreateResourcePoolTaskParam(*t.CreateTenantParam.Name, t.CreateTenantParam.ZoneList, t.timestamp)
	// drop resource if not used
	return pool.DropFreeResourcePools(t.Task, t.createResourcePoolParam)
}

type SetTenantTimeZoneTask struct {
	task.Task
	timeZone   string
	tenantName string
}

func newSetTenantTimeZoneTask() *SetTenantTimeZoneTask {
	newTask := &SetTenantTimeZoneTask{
		Task: *task.NewSubTask(TASK_NAME_SET_TENANT_TIME_ZONE),
	}
	newTask.SetCanRollback().SetCanRetry().SetCanCancel().SetCanPass().SetCanContinue()
	return newTask
}

func newSetTenantTimeZoneNode(timeZone string) *task.Node {
	ctx := task.NewTaskContext().SetParam(PARAM_TENANT_TIME_ZONE, timeZone)
	return task.NewNodeWithContext(newSetTenantTimeZoneTask(), false, ctx)
}

func (t *SetTenantTimeZoneTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_NAME, &t.tenantName); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_TENANT_TIME_ZONE, &t.timeZone); err != nil {
		return err
	}
	err := tenantService.SetTenantVariables(t.tenantName, map[string]interface{}{constant.VARIABLE_TIME_ZONE: t.timeZone})
	if err != nil {
		t.ExecuteWarnLogf("Set tenant %s time zone failed: %s", t.tenantName, err.Error())
	}
	return nil // always return nil to continue the dag
}
