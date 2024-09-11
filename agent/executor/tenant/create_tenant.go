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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/pool"
	"github.com/oceanbase/obshell/agent/executor/script"
	tenantservice "github.com/oceanbase/obshell/agent/service/tenant"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
	log "github.com/sirupsen/logrus"
)

func checkReplicaType(localityType string) error {
	if localityType != constant.REPLICA_TYPE_FULL && localityType != constant.REPLICA_TYPE_READONLY && localityType != "" {
		return errors.New("ReplicaType should be 'FULL' or 'READONLY'")
	}
	return nil
}

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

func checkPrimaryZoneAndLocality(primaryZone string, locality map[string]string) error {
	// Get first priority zones.
	firstPriorityZones := make([]string, 0)
	if primaryZone == constant.PRIMARY_ZONE_RANDOM {
		for zone := range locality {
			firstPriorityZones = append(firstPriorityZones, zone)
		}
	} else {
		firstPriorityZones = strings.Split(strings.Split(primaryZone, ";")[0], ",")
	}

	// Build zone -> region map
	zonesWithRegion, err := obclusterService.GetAllZonesWithRegion()
	if err != nil {
		return err
	}
	zoneToRegionMap := make(map[string]string, 0)
	for _, z := range zonesWithRegion {
		zoneToRegionMap[z.Zone] = z.Region
	}

	// Check whether first priority zones are in the same region.
	var firstPriorityRegion string
	for _, zone := range firstPriorityZones {
		if firstPriorityRegion == "" {
			firstPriorityRegion = zoneToRegionMap[zone]
		} else if firstPriorityRegion != zoneToRegionMap[zone] {
			return errors.New("Tenant primary zone could not span regions.")
		}
	}

	// Check whether the locality has multi-region.
	firstPriorityRegion = zoneToRegionMap[firstPriorityZones[0]]
	hasMultiRegion := false
	for zone := range locality {
		if zoneToRegionMap[zone] != firstPriorityRegion {
			hasMultiRegion = true
			break
		}
	}
	// If there is only one region, no need to check the number of full replicas.
	if !hasMultiRegion {
		return nil
	}

	// The first priority region should have more than 1 full replica when locality has multi-region.
	fullReplicaNum := 0
	for zone, replicaType := range locality {
		if zoneToRegionMap[zone] == firstPriorityRegion {
			arr := strings.Split(replicaType, "{")
			if arr[0] == constant.REPLICA_TYPE_FULL || arr[0] == "F" || arr[0] == "" {
				fullReplicaNum++
			}
		}
	}
	if fullReplicaNum < 2 {
		return errors.Errorf("The region %v where the first priority of tenant zone is located needs to have at least 2 F replicas. In fact, there are only %d full replicas.", firstPriorityRegion, fullReplicaNum)
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

func checkScenario(scenario string) error {
	if scenario == "" {
		return nil
	}

	scenarios := getAllSupportedScenarios()
	if len(scenarios) == 0 {
		return errors.New("current observer does not support scenario")
	}
	if utils.ContainsString(scenarios, strings.ToLower(scenario)) {
		return nil
	}
	return errors.Errorf("scenario only support to be one of %s", strings.Join(scenarios, ", "))
}

func checkPrimaryZone(primaryZone string, zoneList []string) error {
	if primaryZone == constant.PRIMARY_ZONE_RANDOM {
		return nil
	}
	zonesSemicolonSeparated := strings.Split(primaryZone, ";")
	exsitZones := make([]string, 0)
	for _, zones := range zonesSemicolonSeparated {
		zonesCommaSeparated := strings.Split(zones, ",")
		for _, zone := range zonesCommaSeparated {
			if !utils.ContainsString(zoneList, zone) {
				return errors.Errorf("Zone '%s' is not in zone_list.", zone)
			} else if utils.ContainsString(exsitZones, zone) {
				return errors.Errorf("Zone '%s' is repeated in primary_zone.", zone)
			} else {
				exsitZones = append(exsitZones, zone)
			}
		}
	}
	return nil
}

func checkZoneParams(zoneList []param.ZoneParam) error {
	if len(zoneList) == 0 {
		return errors.New("zone_list is empty")
	}

	if err := staticCheckForZoneParams(zoneList); err != nil {
		return err
	}

	for _, zone := range zoneList {
		// Check whether the zone exists
		if exist, err := obclusterService.IsZoneExist(zone.Name); err != nil {
			return err
		} else if !exist {
			return errors.Errorf("Zone '%s' is not exist.", zone.Name)
		}

		// Check unit config if exsits.
		if exist, err := unitService.IsUnitConfigExist(zone.UnitConfigName); err != nil {
			return err
		} else if !exist {
			return errors.Errorf("Unit config '%s' is not exist.", zone.UnitConfigName)
		}

		servers, err := obclusterService.GetServerByZone(zone.Name)
		if err != nil {
			return err
		}
		if len(servers) < zone.UnitNum {
			return errors.Errorf("The number of servers in zone '%s' is %d, less than the number of units %d.", zone.Name, len(servers), zone.UnitNum)
		}
	}
	return nil
}

func checkAtLeastOnePaxosReplica(zoneList []param.ZoneParam) error {
	for _, zone := range zoneList {
		if zone.ReplicaType == constant.REPLICA_TYPE_FULL {
			return nil
		}
	}
	return errors.New("At least one zone should be FULL replica.")
}

func staticCheckForZoneParams(zoneList []param.ZoneParam) error {
	unitNum := 0
	existZones := make([]string, 0)
	for _, zone := range zoneList {
		if utils.ContainsString(existZones, zone.Name) {
			return errors.Errorf("Zone '%s' is repeated.", zone.Name)
		}
		existZones = append(existZones, zone.Name)

		if zone.UnitConfigName == "" {
			return errors.New("unit_config_name should not be empty.")
		}

		// Check replica type.
		if err := checkReplicaType(zone.ReplicaType); err != nil {
			return err
		}

		// Check unit num.
		if zone.UnitNum <= 0 {
			return errors.New("unit_num should be positive.")
		}

		if zone.UnitNum != unitNum && unitNum != 0 {
			return errors.New("unit_num should be same in all zones.")
		}
		unitNum = zone.UnitNum
	}
	return nil
}

func renderZoneParams(zoneList []param.ZoneParam) {
	for i := range zoneList {
		if zoneList[i].ReplicaType == "" {
			zoneList[i].ReplicaType = constant.REPLICA_TYPE_FULL
		} else {
			zoneList[i].ReplicaType = strings.ToUpper(zoneList[i].ReplicaType)
		}
	}
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

	renderZoneParams(param.ZoneList)
	return nil
}

func checkCreateTenantParam(param *param.CreateTenantParam) (err error) {
	if len(param.ZoneList) == 0 {
		return errors.New("zone_list is empty")
	}

	if param.Mode != constant.MYSQL_MODE {
		return errors.New("only support mysql mode")
	}

	if err = checkZoneParams(param.ZoneList); err != nil {
		return
	}

	if err = checkAtLeastOnePaxosReplica(param.ZoneList); err != nil {
		return
	}

	zoneList := make([]string, 0)
	for _, zone := range param.ZoneList {
		zoneList = append(zoneList, zone.Name)
	}
	if err = checkPrimaryZone(param.PrimaryZone, zoneList); err != nil {
		return
	}

	if err = checkScenario(param.Scenario); err != nil {
		return
	}

	if err = checkCharsetAndCollation(param.Charset, param.Collation); err != nil {
		return
	}

	locality := make(map[string]string, 0)
	for _, zone := range param.ZoneList {
		locality[zone.Name] = zone.ReplicaType
	}
	if err = checkPrimaryZoneAndLocality(param.PrimaryZone, locality); err != nil {
		return
	}

	if err = checkVariables(param.Variables); err != nil {
		return
	}

	if err = checkParameters(param.Parameters); err != nil {
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
	for _, server := range source {
		gatheredUnitInfo, err := gatherAllUnitsOnServer(server.SvrIp, server.SvrPort)
		if err != nil {
			return err
		}
		log.Infof("server %s:%d used resource: %v", server.SvrIp, server.SvrPort, gatheredUnitInfo)
		if server.CpuCapacity-gatheredUnitInfo.MinCpu < unit.MinCpu ||
			server.CpuCapacityMax-gatheredUnitInfo.MaxCpu < unit.MaxCpu {
			err = errors.Errorf("server %s:%d CPU resource not enough", server.SvrIp, server.SvrPort)
			continue
		}
		if server.MemCapacity-gatheredUnitInfo.MemorySize < unit.MemorySize {
			err = errors.Errorf("server %s:%d MEMORY_SIZE resource not enough", server.SvrIp, server.SvrPort)
			continue
		}
		if server.LogDiskCapacity-gatheredUnitInfo.LogDiskSize < unit.LogDiskSize {
			err = errors.Errorf("server %s:%d LOG_DISK_SIZE resource not enough", server.SvrIp, server.SvrPort)
			continue
		}
		validServer += 1
	}
	if validServer >= unitNum {
		return nil
	}
	return err
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
		return nil, errors.Errorf("Get all units on server %s:%d failed.", svrIp, svrPort)
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
	template, err := buildCreateTenatDagTemplate(param)
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

func buildCreateTenatDagTemplate(param *param.CreateTenantParam) (*task.Template, error) {
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
		templateBuilder.AddTask(newSetTenantParameterTask(), false)
	}
	if param.Scenario != "" {
		templateBuilder.AddNode(newOptimizeTenantNode(param.Scenario, param))
	}
	templateBuilder.AddNode(newModifyTenantWhitelistNode(*param.Whitelist))

	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}
	templateBuilder.AddNode(script.NewParallelImportScriptForTenantNode(agents, false))

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
		SetParam(PARAM_TENANT_PARAMETER, param.Parameters).
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

func buildCreateTenantSql(param param.CreateTenantParam, poolList []string) (string, []interface{}) {
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
		if _, ok := v.(string); ok {
			sql += ", " + k + "= `%s`"
		} else {
			sql += ", " + k + "= %v"
		}
		input = append(input, v)
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

	var createdResourcePool []param.CreateResourcePoolTaskParam
	t.createResourcePoolParam = buildCreateResourcePoolTaskParam(*t.CreateTenantParam.Name, t.CreateTenantParam.ZoneList, t.timestamp)

	pool.CreatePools(t.Task, t.createResourcePoolParam)

	var poolList []string
	for _, poolParam := range t.createResourcePoolParam {
		poolList = append(poolList, poolParam.PoolName)
	}
	basic, input := buildCreateTenantSql(t.CreateTenantParam, poolList)
	sql := fmt.Sprintf(basic, input...)
	t.ExecuteLogf("Create tenant sql: %s", sql)
	if err := tenantService.TryExecute(sql); err != nil {
		// drop all created resource pool
		pool.DropFreeResourcePools(t.Task, createdResourcePool)
		return err
	}
	// get tenant id
	tenantID, err := tenantService.GetTenantId(*t.CreateTenantParam.Name)
	if err != nil {
		return err
	}
	t.ExecuteLogf("Create tenant success, tenant id: %d", tenantID)
	t.GetContext().SetData(PARAM_TENANT_ID, tenantID)
	return nil
}

func (t *CreateTenantTask) Rollback() error {
	if err := t.GetContext().GetParamWithValue(PARAM_CREATE_TENANT, &t.CreateTenantParam); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_TIMESTAMP, &t.timestamp); err != nil {
		return errors.Wrapf(err, "Get timestamp failed")
	}

	t.GetContext().GetDataWithValue(PARAM_TENANT_ID, &t.id) // ignore the error

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
