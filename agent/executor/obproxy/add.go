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

package obproxy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/meta"
	obproxydb "github.com/oceanbase/obshell/agent/repository/db/obproxy"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type addObproxyOptions struct {
	appName                  string
	homePath                 string
	version                  string
	clusterName              string // only for "RS_LIST" mode
	encryptedSysPwd          string
	encryptedProxyroPassword string
	parameters               map[string]string
	sqlPort                  int
	exportPort               int
}

func buildAddObproxyOptions(param *param.AddObproxyParam) (*addObproxyOptions, error) {
	version, err := getObproxyVersion(param.HomePath)
	if err != nil {
		return nil, err
	}

	options := addObproxyOptions{
		appName:    param.Name,
		homePath:   param.HomePath,
		version:    version,
		parameters: make(map[string]string),
	}

	for k, v := range param.Parameters {
		options.parameters[k] = v
	}

	options.encryptedSysPwd, err = secure.Encrypt(param.ObproxySysPassword)
	if err != nil {
		return nil, err
	}

	options.encryptedProxyroPassword, err = secure.Encrypt(param.ProxyroPassword)
	if err != nil {
		return nil, err
	}
	return &options, nil
}

func AddObproxy(param param.AddObproxyParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if meta.IsObproxyAgent() {
		return nil, errors.Occur(errors.ErrBadRequest, "agent has already managed obproxy")
	}

	if err := checkObproxyHomePath(param.HomePath); err != nil {
		return nil, errors.Occurf(errors.ErrBadRequest, "invalid obproxy home path: %s", err)
	}

	options, err := buildAddObproxyOptions(&param)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}

	if err := checkAndFillObproxyPort(&param, options); err != nil {
		return nil, err
	}

	// Check obproxy version
	if err := checkAndFillObproxyVersion(&param, options); err != nil {
		return nil, err
	}

	if err := checkAndFillWorkMode(&param, options); err != nil {
		return nil, err
	}

	if rsList, ok := options.parameters[constant.OBPROXY_CONFIG_RS_LIST]; ok && rsList != "" {
		if clusterName, err := checkProxyroPasswordAndGetClusterName(rsList, param.ProxyroPassword); err != nil {
			return nil, errors.Occur(errors.ErrBadRequest, err)
		} else {
			options.clusterName = clusterName
			log.Infof("cluster name: %s", clusterName)
		}
	}

	ctx := buildAddObproxyContext(options)
	template := buildAddObproxyTemplate(options)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

func checkAndFillWorkMode(param *param.AddObproxyParam, options *addObproxyOptions) *errors.OcsAgentError {
	// Check work mode.
	if param.RsList != nil && param.ConfigUrl != nil {
		return errors.Occur(errors.ErrBadRequest, "rs_list and config_url can not be specified at the same time")
	}

	if param.RsList != nil {
		options.parameters[constant.OBPROXY_CONFIG_RS_LIST] = *param.RsList
		options.parameters[constant.OBPROXY_CONFIG_CONFIG_SERVER_URL] = ""
	} else if param.ConfigUrl != nil {
		options.parameters[constant.OBPROXY_CONFIG_CONFIG_SERVER_URL] = *param.ConfigUrl
	} else {
		if !meta.OCS_AGENT.IsClusterAgent() {
			return errors.Occur(errors.ErrBadRequest, "rs_list or config_url must be specified when agent is not cluster agent")
		} else {
			// Use the rs_list of current ob cluster.
			rsListStr, err := obclusterService.GetRsListStr()
			if err != nil {
				// The observer may be inactive.
				return errors.Occur(errors.ErrUnexpected, err)
			}

			options.parameters[constant.OBPROXY_CONFIG_RS_LIST] = convertToRootServerList(rsListStr)
			options.parameters[constant.OBPROXY_CONFIG_CONFIG_SERVER_URL] = ""
		}
	}
	return nil
}

func checkObproxyHomePath(homePath string) error {
	if err := utils.CheckDirExists(homePath); err != nil {
		return err
	}

	err := syscall.Access(homePath, syscall.O_RDWR)
	if err != nil {
		return errors.Errorf("no read/write permission for directory '%s'", homePath)
	}

	// Check obproxy is installed.
	if err := utils.CheckDirExists(filepath.Join(homePath, constant.OBPROXY_DIR_BIN)); err != nil {
		return err
	}
	if err := utils.CheckDirExists(filepath.Join(homePath, constant.OBPROXY_DIR_LIB)); err != nil {
		return err
	}

	// Check if obproxy has run in the home path.
	entrys, err := os.ReadDir(filepath.Join(homePath, constant.OBPROXY_DIR_ETC))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if len(entrys) != 0 {
		return errors.New("obproxy etc directory is not empty")
	}

	return nil
}

// checkObproxyVersion checks the version of obproxy located at the given homePath.
// If the version is lower than the minimum supported version (4.0.0), it returns an error.
func checkAndFillObproxyVersion(param *param.AddObproxyParam, options *addObproxyOptions) *errors.OcsAgentError {
	version, err := getObproxyVersion(param.HomePath)
	if err != nil {
		return errors.Occur(errors.ErrBadRequest, "get obproxy version failed: %v", err)
	}
	if version < constant.OBPROXY_MIN_VERSION_SUPPORT {
		return errors.Occurf(errors.ErrBadRequest, "obproxy version %s is lower than the minimum supported version %s", version, constant.OBPROXY_MIN_VERSION_SUPPORT)
	}
	options.version = version
	return nil
}

func checkProxyroPasswordAndGetClusterName(rsListStr string, password string) (clusterName string, err error) {
	rsList := strings.Split(rsListStr, ";")
	dsConfig := config.NewObDataSourceConfig().
		SetTryTimes(1).
		SetDBName(constant.DB_OCEANBASE).
		SetTimeout(10).
		SetPassword(password).
		SetUsername(constant.SYS_USER_PROXYRO)
	var tempDb *gorm.DB
	defer func() {
		if tempDb != nil {
			oceanbaseDB, _ := tempDb.DB()
			oceanbaseDB.Close()
		}
	}()
	for _, rs := range rsList {
		observerInfo := meta.NewAgentInfoByString(rs)
		if observerInfo == nil {
			err = errors.Errorf("invalid observer info: %s", rs)
			continue
		}
		dsConfig.SetIp(observerInfo.GetIp()).SetPort(observerInfo.GetPort())
		tempDb, err = oceanbase.LoadTempOceanbaseInstance(dsConfig)
		if err != nil {
			continue
		}
		clusterName, err = obproxyService.GetObclusterName(tempDb)
		if err != nil {
			continue
		}

		return clusterName, nil
	}
	return "", err
}

func checkAndFillObproxyPort(param *param.AddObproxyParam, options *addObproxyOptions) *errors.OcsAgentError {
	if param.SqlPort != nil {
		options.parameters[constant.OBPROXY_CONFIG_LISTEN_PORT] = strconv.Itoa(*param.SqlPort)
	} else if options.parameters[constant.OBPROXY_CONFIG_LISTEN_PORT] == "" {
		options.parameters[constant.OBPROXY_CONFIG_LISTEN_PORT] = strconv.Itoa(constant.OBPROXY_DEFAULT_SQL_PORT)
	}
	if param.RpcPort != nil {
		options.parameters[constant.OBPROXY_CONFIG_RPC_LISTEN_PORT] = strconv.Itoa(*param.RpcPort)
	} else if options.parameters[constant.OBPROXY_CONFIG_RPC_LISTEN_PORT] == "" {
		options.parameters[constant.OBPROXY_CONFIG_RPC_LISTEN_PORT] = strconv.Itoa(constant.OBPROXY_DEFAULT_RPC_PORT)
	}
	if param.ExporterPort != nil {
		options.parameters[constant.OBPROXY_CONFIG_PROMETHUES_LISTEN_PORT] = strconv.Itoa(*param.ExporterPort)
	} else if options.parameters[constant.OBPROXY_CONFIG_PROMETHUES_LISTEN_PORT] == "" {
		options.parameters[constant.OBPROXY_CONFIG_PROMETHUES_LISTEN_PORT] = strconv.Itoa(constant.OBPROXY_DEFAULT_EXPORTER_PORT)
	}

	// Check port is valid.
	var ports = []string{constant.OBPROXY_CONFIG_LISTEN_PORT, constant.OBPROXY_CONFIG_PROMETHUES_LISTEN_PORT, constant.OBPROXY_CONFIG_RPC_LISTEN_PORT}
	for _, port := range ports {
		if _, err := strconv.Atoi(options.parameters[port]); err != nil {
			return errors.Occur(errors.ErrBadRequest, "invalid port: %s", options.parameters[port])
		}
	}
	options.sqlPort, _ = strconv.Atoi(options.parameters[constant.OBPROXY_CONFIG_LISTEN_PORT])
	options.exportPort, _ = strconv.Atoi(options.parameters[constant.OBPROXY_CONFIG_PROMETHUES_LISTEN_PORT])
	return nil
}

func buildAddObproxyContext(options *addObproxyOptions) *task.TaskContext {
	ctx := task.NewTaskContext().
		SetParam(PARAM_OBPROXY_HOME_PATH, options.homePath).
		SetParam(PARAM_OBPROXY_SQL_PORT, options.sqlPort).
		SetParam(PARAM_OBPROXY_EXPORTER_PORT, options.exportPort).
		SetParam(PARAM_OBPROXY_APP_NAME, options.appName).
		SetParam(PARAM_OBPROXY_VERSION, options.version).
		SetParam(PARAM_OBPROXY_CLUSTER_NAME, options.clusterName).
		SetParam(PARAM_OBPROXY_SYS_PASSWORD, options.encryptedSysPwd)
	return ctx
}

func buildAddObproxyTemplate(options *addObproxyOptions) *task.Template {
	templateBuilder := task.NewTemplateBuilder(DAG_ADD_OBPROXY).
		SetType(task.DAG_OBPROXY).
		AddNode(newPrepareForObproxyAgentNode(false)).
		AddNode(newStartObproxyNode(options.parameters)).
		AddNode(NewSetObproxyUserPasswordForObNode(options.encryptedProxyroPassword)).
		AddTask(newPersistObproxyInfoTask(), false).
		SetMaintenance(task.ObproxyMaintenance())
	return templateBuilder.Build()
}

func convertToRootServerList(rsListStr string) string {
	var result []string
	entries := strings.Split(rsListStr, ";")
	for _, entry := range entries {
		parts := strings.Split(entry, ":")
		if len(parts) == 3 {
			result = append(result, fmt.Sprintf("%s:%s", parts[0], parts[2]))
		}
	}
	return strings.Join(result, ";")
}

type StartObproxyTask struct {
	task.Task
	homePath           string
	sqlPort            int
	parameters         map[string]string
	appName            string
	optionsStr         string
	clusterName        string
	obproxySysPassword string
	encryptedSysPwd    string

	startWithOption bool
}

func newStartObproxyNode(parameters map[string]string) *task.Node {
	newTask := newStartObproxyTask()
	ctx := task.NewTaskContext().SetParam(PARAM_OBPROXY_START_PARAMS, parameters).SetParam(PARAM_OBPROXY_START_WITH_OPTIONS, true)
	return task.NewNodeWithContext(newTask, false, ctx)
}

func newStartObproxyWithoutOptionsNode() *task.Node {
	newTask := newStartObproxyTask()
	ctx := task.NewTaskContext().SetParam(PARAM_OBPROXY_START_WITH_OPTIONS, false)
	return task.NewNodeWithContext(newTask, false, ctx)
}

func newStartObproxyTask() *StartObproxyTask {
	newTask := &StartObproxyTask{
		Task: *task.NewSubTask(TASK_START_OBPROXY),
	}
	newTask.SetCanRetry().SetCanContinue()
	return newTask
}

func (t *StartObproxyTask) Execute() error {
	var err error
	if err = t.GetContext().GetParamWithValue(PARAM_OBPROXY_HOME_PATH, &t.homePath); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_OBPROXY_START_WITH_OPTIONS, &t.startWithOption); err != nil {
		return err
	}

	var startCmd string
	if !t.startWithOption {
		startCmd = t.buildAtartObproxyWithoutOptionsCmd(t.homePath)
	} else {
		if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_START_PARAMS, &t.parameters); err != nil {
			return err
		}
		if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_APP_NAME, &t.appName); err != nil {
			return err
		}
		if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_CLUSTER_NAME, &t.clusterName); err != nil {
			return err
		}
		if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_SYS_PASSWORD, &t.encryptedSysPwd); err != nil {
			return err
		}
		if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_SQL_PORT, &t.sqlPort); err != nil {
			return err
		}
		t.obproxySysPassword, err = secure.Decrypt(t.encryptedSysPwd)
		if err != nil {
			return errors.Errorf("decrypt obproxy sys password failed: %v", err)
		}

		if err := t.buildStartOptionStr(); err != nil {
			return err
		}

		startCmd = fmt.Sprintf("cd %s; ./bin/obproxy -o %s", t.homePath, t.optionsStr)
		if t.appName != "" {
			startCmd = fmt.Sprintf("%s -n %s", startCmd, t.appName)
		}
		if t.clusterName != "" {
			startCmd = fmt.Sprintf("%s -c %s", startCmd, t.clusterName)
		}
	}
	t.ExecuteLogf("start obproxy cmd: %s", startCmd)
	if output, err := exec.Command("/bin/bash", "-c", startCmd).CombinedOutput(); err != nil {
		return errors.Errorf("failed to start obproxy: %v, output: %s", err, string(output))
	}

	if err := t.healthCheck(); err != nil {
		return errors.Wrap(err, "obproxy start failed")
	}

	if pid, err := process.FindPIDByPort(uint32(t.sqlPort)); err != nil {
		return errors.Errorf("get obproxy pid failed: %v", err)
	} else if err := process.WritePidForce(filepath.Join(t.homePath, constant.OBPROXY_DIR_RUN, "obproxy.pid"), int(pid)); err != nil {
		return errors.Errorf("write obproxy pid failed: %v", err)
	}
	return nil
}

func (t *StartObproxyTask) buildStartOptionStr() error {
	parameters := t.parameters
	// Add single quotes to rs_list.
	if rsList, ok := parameters[constant.OBPROXY_CONFIG_RS_LIST]; ok && !strings.HasPrefix(rsList, "'") && !strings.HasSuffix(rsList, "'") {
		parameters[constant.OBPROXY_CONFIG_RS_LIST] = fmt.Sprintf("'%s'", rsList)
	}

	if t.obproxySysPassword != "" {
		parameters[constant.OBPROXY_CONFIG_OBPROXY_SYS_PASSWORD] = utils.Sha1(t.obproxySysPassword)
	} else {
		// If obproxy sys password is empty, do not need to sha1 it.
		parameters[constant.OBPROXY_CONFIG_OBPROXY_SYS_PASSWORD] = ""
	}

	optionStrs := make([]string, 0, len(parameters))
	for k, v := range parameters {
		optionStrs = append(optionStrs, fmt.Sprintf("%s=%s", k, v))
	}
	t.optionsStr = strings.Join(optionStrs, ",")
	return nil
}

func (t *StartObproxyTask) buildAtartObproxyWithoutOptionsCmd(homePath string) string {
	return fmt.Sprintf("cd %s; ./bin/obproxy", homePath)
}

func (t *StartObproxyTask) healthCheck() error {
	// Try to connect to obproxy to confirm that it has started.
	if t.sqlPort == 0 {
		t.sqlPort = meta.OBPROXY_SQL_PORT
	}
	if t.obproxySysPassword == "" {
		t.obproxySysPassword = meta.OBPROXY_SYS_PWD
	}
	t.ExecuteLog("start obproxy health check")
	dsConfig := config.NewObproxyDataSourceConfig().SetPort(t.sqlPort).SetPassword(t.obproxySysPassword)
	for retryCount := 1; retryCount <= obproxydb.WAIT_OBPROXY_CONNECTED_MAX_TIMES; retryCount++ {
		time.Sleep(obproxydb.WAIT_OBPROXY_CONNECTED_MAX_INTERVAL)
		if err := obproxydb.LoadObproxyInstanceForHealthCheck(dsConfig); err != nil {
			t.ExecuteWarnLogf("obproxy health check failed: %v", err)
			if strings.Contains(err.Error(), "connection refused") {
				return err
			}
			continue
		}
		if err := obproxyService.UpdateSqlPort(t.sqlPort); err != nil {
			return errors.Errorf("update obproxy sql port failed: %v", err)
		}
		if err := obproxyService.UpdateObproxySysPassword(t.obproxySysPassword); err != nil {
			return errors.Errorf("update obproxy sys password failed: %v", err)
		}
		return nil
	}

	return errors.New("obproxy health check timeout")
}

type PersistObproxyInfoTask struct {
	task.Task
	homePath                 string
	sqlPort                  int
	version                  string
	encryptedSysPwd          string
	encryptedProxyroPassword string
}

func newPersistObproxyInfoTask() *PersistObproxyInfoTask {
	newTask := &PersistObproxyInfoTask{
		Task: *task.NewSubTask(TASK_PERSIST_OBPROXY_INFP),
	}
	newTask.SetCanRetry().SetCanContinue()
	return newTask
}

func (t *PersistObproxyInfoTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_HOME_PATH, &t.homePath); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_SQL_PORT, &t.sqlPort); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_VERSION, &t.version); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_SYS_PASSWORD, &t.encryptedSysPwd); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_PROXYRO_PASSWORD, &t.encryptedProxyroPassword); err != nil {
		return err
	}
	if err := agentService.AddObproxy(t.homePath, t.sqlPort, t.version, t.encryptedSysPwd, t.encryptedProxyroPassword); err != nil {
		return err
	}

	return nil
}

type PrepareForAddObproxyTask struct {
	task.Task
	expectObproxyAgent bool
	homePath           string
}

// PrepareForAddObproxyNode will check if the agent is an obproxy agent.
func newPrepareForObproxyAgentNode(expectObproxyAgent bool) *task.Node {
	newTask := &PrepareForAddObproxyTask{
		Task: *task.NewSubTask(TASK_CHECK_OBPROXY_STATUS),
	}
	newTask.SetCanRetry().SetCanContinue()

	ctx := task.NewTaskContext().SetParam(task.FAILURE_EXIT_MAINTENANCE, true).SetParam(PARAM_EXPECT_OBPROXY_AGENT, expectObproxyAgent)
	return task.NewNodeWithContext(newTask, false, ctx)
}

func (t *PrepareForAddObproxyTask) Execute() error {
	// Double check if the agent identify.
	if err := t.GetContext().GetParamWithValue(PARAM_EXPECT_OBPROXY_AGENT, &t.expectObproxyAgent); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_OBPROXY_HOME_PATH, &t.homePath); err != nil {
		return err
	}
	if t.expectObproxyAgent && !meta.IsObproxyAgent() {
		return errors.Errorf("This is not an obproxy agent")
	}
	if !t.expectObproxyAgent {
		if meta.IsObproxyAgent() {
			return errors.Errorf("agent has already managed obproxy")
		}
		// Create obproxy run path
		runPath := filepath.Join(t.homePath, constant.OBPROXY_DIR_RUN)
		if err := os.MkdirAll(runPath, 0755); err != nil {
			return errors.Errorf("create obproxy run path failed: %v", err)
		}
	}
	return nil
}

// Only support set global proxyro password currently
type SetObproxyUserPasswordForObTask struct {
	task.Task
	encryptedProxyroPassword string
}

func NewSetObproxyUserPasswordForObNode(encryptedProxyroPassword string) *task.Node {
	newTask := &SetObproxyUserPasswordForObTask{
		Task: *task.NewSubTask(TASK_SET_OBPROXY_USER_PASSWORD),
	}
	newTask.SetCanRetry().SetCanContinue()
	ctx := task.NewTaskContext().SetParam(PARAM_OBPROXY_PROXYRO_PASSWORD, encryptedProxyroPassword)
	return task.NewNodeWithContext(newTask, false, ctx)
}

func (t *SetObproxyUserPasswordForObTask) Execute() error {
	if t.GetContext().GetParamWithValue(PARAM_OBPROXY_PROXYRO_PASSWORD, &t.encryptedProxyroPassword) != nil {
		return errors.Errorf("get obproxy user password failed")
	}

	// Decrypt proxyro password.
	proxyroPassword, err := secure.Decrypt(t.encryptedProxyroPassword)
	if err != nil {
		return errors.Errorf("decrypt proxyro password failed: %v", err)
	}
	t.ExecuteLog("set obproxy user password")

	if err := obproxyService.SetProxyroPassword(proxyroPassword); err != nil {
		return errors.Errorf("set obproxy user password failed: %v", err)
	}
	return nil
}
