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
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/agent/secure"
)

func (s *AgentService) GetAllAgentInstances() (agents []meta.AgentInstance, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.AllAgent{}).Find(&agents).Error
	return
}

func (s *AgentService) RecoveryMasterAgentFromOB() (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	var masterAgent meta.AgentInfoWithZone
	if err = oceanbaseDb.Model(&oceanbase.AllAgent{}).Where("identity = ?", meta.MASTER).First(&masterAgent).Error; err != nil {
		return nil
	}
	ocsAgent.MasterAgent = meta.NewAgentWithZoneByAgentInfo(&masterAgent, masterAgent.Zone)
	return nil
}

func (s *AgentService) GetAllAgentsInfo() (agents []meta.AgentInfo, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.AllAgent{}).Find(&agents).Error
	return
}

func (s *AgentService) GetAllAgentsInfoFromOB() (agents []meta.AgentInfo, err error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = db.Model(&oceanbase.AllAgent{}).Find(&agents).Error
	return
}

func (s *AgentService) GetTakeOverMasterAgent() (agent *meta.AgentInfo, err error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = db.Model(&oceanbase.AllAgent{}).Where("identity = ?", meta.TAKE_OVER_MASTER).Scan(&agent).Error
	return
}

func (s *AgentService) GetAllAgentsDOFromOB() (agents []oceanbase.AllAgent, err error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = db.Model(&oceanbase.AllAgent{}).Find(&agents).Error
	return
}

func (s *AgentService) GetAllAgentsDO() (agents []sqlite.AllAgent, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	err = db.Model(&sqlite.AllAgent{}).Find(&agents).Error
	return
}

func (s *AgentService) GetAgentDO(agentInfo meta.AgentInfoInterface) (agent sqlite.AllAgent, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.AllAgent{}).Where("ip=? and port=?", agentInfo.GetIp(), agentInfo.GetPort()).Scan(&agent).Error
	return
}

// GetAgentInstance will get agent by ip and port, if not exist, return err
func (s *AgentService) GetAgentInstance(agentInfo meta.AgentInfoInterface) (*meta.AgentInstance, error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}

	var agent *meta.AgentInstance
	err = db.Model(&sqlite.AllAgent{}).Where("ip=? and port=?", agentInfo.GetIp(), agentInfo.GetPort()).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return agent, nil
}

// FindAgentInstance will get agent by ip and port, if not exist, return nil
func (s *AgentService) FindAgentInstance(agentInfo meta.AgentInfoInterface) (agent *meta.AgentInstance, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}

	err = db.Model(&sqlite.AllAgent{}).Where("ip=? and port=?", agentInfo.GetIp(), agentInfo.GetPort()).Scan(&agent).Error
	return
}

func (s *AgentService) GetAgentInstanceByZone(zone string) (agents []meta.AgentInstance, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.AllAgent{}).Where("zone = ?", zone).Find(&agents).Error
	return
}

func (s *AgentService) GetAgentInfoByZoneFromOB(zone string) (agents []meta.AgentInfo, err error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = db.Model(&oceanbase.AllAgent{}).Where("zone = ?", zone).Find(&agents).Error
	return
}

func (s *AgentService) FindAgentInstanceByZone(zone string) (agents []meta.AgentInstance, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.AllAgent{}).Where("zone = ?", zone).Scan(&agents).Error
	return
}

func (s *AgentService) IsAgentExist(agentInfo meta.AgentInfoInterface) (bool, error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Model(&sqlite.AllAgent{}).Where("ip=? and port=?", agentInfo.GetIp(), agentInfo.GetPort()).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddAgent will add agent to all_agent
func (s *AgentService) AddAgent(agentInstance meta.Agent, homePath string, os string, arch string, publicKey string, token string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		err = s.addAgentToken(tx, agentInstance, token)
		if err == nil {
			err = s.addAgent(tx, agentInstance, homePath, os, arch, publicKey)
		}
		return err
	})
}

func (s *AgentService) AddAgentInOB(agent oceanbase.AllAgent) error {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return db.Model(&oceanbase.AllAgent{}).Create(&agent).Error
}

func (s *AgentService) addAgentToken(tx *gorm.DB, agentInfo meta.AgentInfoInterface, token string) error {
	ocsToken := sqlite.OcsToken{
		Ip:    agentInfo.GetIp(),
		Port:  agentInfo.GetPort(),
		Token: token,
	}
	return tx.Model(&sqlite.OcsToken{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ip"}, {Name: "port"}},
		DoUpdates: clause.AssignmentColumns([]string{"token"}),
	}).Create(&ocsToken).Error
}

func (s *AgentService) addAgent(db *gorm.DB, agentInstance meta.Agent, homePath string, os string, arch string, publicKey string) error {
	agent := &sqlite.AllAgent{
		Ip:           agentInstance.GetIp(),
		Port:         agentInstance.GetPort(),
		Identity:     string(agentInstance.GetIdentity()),
		Os:           os,
		Architecture: arch,
		Version:      agentInstance.GetVersion(),
		Zone:         agentInstance.GetZone(),
		HomePath:     homePath,
		PublicKey:    publicKey,
	}
	return db.Create(agent).Error
}

func (s *AgentService) UpdateAgent(agentInstance meta.Agent, homePath string, os string, arch string, publicKey string, token string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		err = s.addAgentToken(tx, agentInstance, token)
		if err == nil {
			err = s.updateAgent(tx, agentInstance, homePath, os, arch, publicKey)
		}
		return err
	})
}

func (s *AgentService) UpdateAgentOBPort(agent meta.AgentInfoInterface, mysqlPort, rpcPort int) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}

	allAgentDO := sqlite.AllAgent{
		Ip:        agent.GetIp(),
		Port:      agent.GetPort(),
		RpcPort:   rpcPort,
		MysqlPort: mysqlPort,
	}
	return db.Updates(&allAgentDO).Error
}

func (s *AgentService) UpdateAgentOBPortWithTx(tx *gorm.DB, agent meta.AgentInfoInterface, mysqlPort, rpcPort int) error {
	allAgentDO := sqlite.AllAgent{
		Ip:        agent.GetIp(),
		Port:      agent.GetPort(),
		RpcPort:   rpcPort,
		MysqlPort: mysqlPort,
	}
	return tx.Updates(&allAgentDO).Error
}

func (s *AgentService) updateAgent(db *gorm.DB, agentInstance meta.Agent, homePath string, os string, arch string, publicKey string) error {
	agent := &sqlite.AllAgent{
		Ip:           agentInstance.GetIp(),
		Port:         agentInstance.GetPort(),
		Identity:     string(agentInstance.GetIdentity()),
		Os:           os,
		Architecture: arch,
		Version:      agentInstance.GetVersion(),
		Zone:         agentInstance.GetZone(),
		HomePath:     homePath,
		PublicKey:    publicKey,
	}
	return db.Updates(agent).Error
}

func (s *AgentService) GetAllAgents() ([]meta.AgentInstance, error) {
	return s.getAllAgentsByIdentity("")
}

func (s *AgentService) GetAllAgentsFromOB() ([]meta.AgentInstance, error) {
	return s.getAllAgentsByIdentityFromOB("")
}

func (s *AgentService) GetFollowerAgentsFromOB() ([]meta.AgentInstance, error) {
	return s.getAllAgentsByIdentityFromOB(meta.FOLLOWER)
}

func (s *AgentService) GetTakeOverFollowerAgentsFromOB() ([]meta.AgentInstance, error) {
	return s.getAllAgentsByIdentityFromOB(meta.TAKE_OVER_FOLLOWER)
}

func (s *AgentService) GetMasterAgentInfo() *meta.AgentInfo {
	switch ocsAgent.GetIdentity() {
	case meta.MASTER:
		return &ocsAgent.AgentInfo
	case meta.FOLLOWER:
		return &ocsAgent.MasterAgent.AgentInfo
	default:
		return nil
	}
}

func (s *AgentService) GetFollowerAgents() ([]meta.AgentInstance, error) {
	return s.getAllAgentsByIdentity(meta.FOLLOWER)
}

func (s *AgentService) getAllAgentsByIdentity(identity meta.AgentIdentity) (agents []meta.AgentInstance, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	resp := sqliteDb.Model(&sqlite.AllAgent{})
	if identity != "" {
		resp = resp.Where("identity", identity)
	}
	err = resp.Find(&agents).Error
	return
}

func (s *AgentService) getAllAgentsByIdentityFromOB(identity meta.AgentIdentity) (agents []meta.AgentInstance, err error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	resp := db.Model(&oceanbase.AllAgent{})
	if identity != "" {
		resp = resp.Where("identity", identity)
	}
	err = resp.Find(&agents).Error
	return
}

func (s *AgentService) deleteAllAgents(db *gorm.DB) error {
	return db.Delete(&sqlite.AllAgent{}, "1=1").Error
}

func (s *AgentService) DeleteAgent(agent meta.AgentInfoInterface) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Where("ip=? and port=?", agent.GetIp(), agent.GetPort()).Delete(&sqlite.AllAgent{}).Error
}

func (s *AgentService) DeleteAgentInOB(agent meta.AgentInfoInterface) error {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return db.Where("ip=? and port=?", agent.GetIp(), agent.GetPort()).Delete(&sqlite.AllAgent{}).Error
}

func (s *AgentService) CheckCanBeTakeOverMaster() (bool, error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return false, err
	}
	var servers []oceanbase.OBServer
	if err = oceanbaseDb.Raw("select * from oceanbase.dba_ob_servers").Find(&servers).Error; err != nil {
		return false, err
	}
	var agents []oceanbase.AllAgent
	if err = oceanbaseDb.Model(oceanbase.AllAgent{}).Find(&agents).Error; err != nil {
		return false, err
	}
	/* Compared to servers, agents just lack the agent $ip:$port */
	/* Check version consistency, ignore invalid agent */
	self_exist := false
	other_exist := true
	for _, server := range servers {
		if server.SvrIp == meta.OCS_AGENT.GetIp() && server.SvrPort == meta.RPC_PORT {
			self_exist = true
			continue
		}
		exist := false
		for _, agent := range agents {
			if server.SvrIp == agent.Ip && server.SvrPort == agent.RpcPort {
				if agent.Version != meta.OCS_AGENT.GetVersion() {
					return false, errors.New("agent version is not consistent")
				}
				exist = true
				break
			}
		}
		other_exist = other_exist && exist
	}

	if !self_exist {
		return false, fmt.Errorf("%s:%d not in cluster", meta.OCS_AGENT.GetIp(), meta.RPC_PORT)
	}

	return other_exist, nil
}

func (s *AgentService) GetAgentInstanceByIpAndRpcPortFromOB(ip string, rpcPort int) (agent *meta.AgentInstance, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	hasTable := oceanbaseDb.Migrator().HasTable(&oceanbase.AllAgent{})
	if !hasTable {
		return
	}
	err = oceanbaseDb.Model(&oceanbase.AllAgent{}).Where("ip=? and rpc_port=?", ip, rpcPort).Scan(&agent).Error
	return
}

func (s *AgentService) UpdateAgentPublicKey(publicKey string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Model(&oceanbase.AllAgent{}).Where("ip=? and port=?", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()).Update(constant.AGENT_PUBLIC_KEY, publicKey).Error
	return
}

func (s *AgentService) CreateTakeOverAgent(identity meta.AgentIdentity) (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	agent := oceanbase.AllAgent{
		Ip:           meta.OCS_AGENT.GetIp(),
		Port:         meta.OCS_AGENT.GetPort(),
		RpcPort:      meta.RPC_PORT,
		Os:           global.Os,
		Architecture: global.Architecture,
		HomePath:     global.HomePath,
		MysqlPort:    meta.MYSQL_PORT,
		Version:      meta.OCS_AGENT.GetVersion(),
		PublicKey:    secure.Public(),
		Zone:         meta.OCS_AGENT.GetZone(),
		Identity:     string(identity),
	}
	err = oceanbaseDb.Transaction(func(oceanbaseTx *gorm.DB) error {
		return sqliteDb.Transaction(func(sqliteTx *gorm.DB) error {
			if err = oceanbaseTx.Model(&oceanbase.AllAgent{}).Create(&agent).Error; err != nil {
				return err
			}
			if identity == meta.TAKE_OVER_FOLLOWER {
				return s.updateIdentity(sqliteTx, identity)
			}
			return nil
		})
	})
	return
}

func (s *AgentService) RemoveInvalidAgent() (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Exec("delete from all_agent where not exists (select 1 from oceanbase.dba_ob_servers where svr_ip = ip and svr_port = rpc_port)").Error
}

func (s *AgentService) UpdateAgentVersion() (err error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return db.Model(&oceanbase.AllAgent{}).Where("ip=? and port=?", meta.OCS_AGENT.GetIp(), meta.OCS_AGENT.GetPort()).Update("version", meta.OCS_AGENT.GetVersion()).Error
}
