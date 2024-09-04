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
	"runtime"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/agent/secure"
)

func (agentService *AgentService) GetTargetServerRpcList(serverRange string) (res []string, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("select ''''||ip||':'||rpc_port||'''' as server from all_agent where ip||':'||port in (%s)", serverRange)
	err = sqliteDb.Raw(sql).Find(&res).Error
	return
}

func (agentService *AgentService) GetTargetServerMysqlList(serverRange string) (res []string, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("select ''''||ip||':'||mysql_port||'''' as server from all_agent where ip||':'||port in (%s)", serverRange)
	err = sqliteDb.Raw(sql).Find(&res).Error
	return
}

func (agentService *AgentService) GetRsOfMaster(identity string) (rs string, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.AllAgent{}).Raw("select ip||':'||rpc_port||':'||mysql_port as rs from all_agent where identity = ? ", identity).Find(&rs).Error
	return
}

func (agentService *AgentService) GetRsListExceptMaster(zone string) (rsList []string, err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = sqliteDb.Model(&sqlite.AllAgent{}).Raw("select ip||':'||rpc_port||':'||mysql_port as rs from all_agent where zone != ? group by zone", zone).Find(&rsList).Error
	return
}

func (s *AgentService) updateInfo(db *gorm.DB, info *sqlite.OcsInfo) error {
	return db.Model(info).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(info).Error
}

func (s *AgentService) updateIdentity(db *gorm.DB, identity meta.AgentIdentity) error {
	info := sqlite.OcsInfo{
		Name:  constant.OCS_INFO_IDENTITY,
		Value: string(identity),
	}
	if err := s.updateInfo(db, &info); err != nil {
		return err
	}
	ocsAgent.Identity = identity
	return nil
}

func (s *AgentService) updateAgentInfo(db *gorm.DB, agentInfo meta.AgentInfoInterface) (err error) {

	infos := []*sqlite.OcsInfo{
		{Name: constant.OCS_INFO_IP, Value: agentInfo.GetIp()},
		{Name: constant.OCS_INFO_PORT, Value: fmt.Sprintf("%d", agentInfo.GetPort())},
	}

	for _, info := range infos {
		if err = s.updateInfo(db, info); err != nil {
			return
		}
	}
	ocsAgent.Ip = agentInfo.GetIp()
	ocsAgent.Port = agentInfo.GetPort()
	return nil
}

func (s *AgentService) UpdateAgentIP(ip string) error {
	if ocsAgent == nil {
		return errors.New("agent is not initialized")
	}
	if ocsAgent.GetIp() != ip {
		if !ocsAgent.IsSingleAgent() && !ocsAgent.IsUnidentified() {
			return errors.New("agent is not single, can not update agent ip")
		}
		ocsAgent.Ip = ip
	}
	return nil
}

func (s *AgentService) UpdateAgentInfo(agentInfo meta.AgentInfoInterface) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// Agent is not under maintenance.
		var status int
		if err = tx.Model(&sqlite.OcsInfo{}).Select("value").Where("name=?", constant.OCS_INFO_STATUS).Scan(&status).Error; err != nil {
			return err
		}
		if status == task.GLOBAL_MAINTENANCE {
			return errors.New("agent is under maintenance, can not update agent info")
		}
		return s.updateAgentInfo(tx, agentInfo)
	})
}

func (s *AgentService) setZone(tx *gorm.DB, zone string) error {
	info := sqlite.OcsInfo{
		Name:  constant.OCS_INFO_ZONE,
		Value: zone,
	}
	if err := s.updateInfo(tx, &info); err != nil {
		return err
	}
	ocsAgent.Zone = zone
	return nil
}

func (s *AgentService) BeMasterAgent(zone string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				ocsAgent.Identity = meta.SINGLE
				ocsAgent.Zone = ""
			}
		}()
		// Update info.
		ocsInfoList := []*sqlite.OcsInfo{
			{Name: constant.OCS_INFO_VERSION, Value: constant.VERSION},
			{Name: constant.OCS_INFO_OS, Value: global.Os},
			{Name: constant.OCS_INFO_ARCHITECTURE, Value: global.Architecture},
		}
		for _, info := range ocsInfoList {
			if err = s.updateInfo(tx, info); err != nil {
				return
			}
		}
		if err = s.setZone(tx, zone); err != nil {
			return
		}
		err = s.updateIdentity(tx, meta.MASTER)

		agentInstance := meta.NewAgentInstanceByAgent(ocsAgent)
		if err = s.addAgent(tx, agentInstance, global.HomePath, global.Os, global.Architecture, secure.Public()); err != nil {
			return
		}
		return
	})
}

func (s *AgentService) BeFollowerAgent(masterAgent meta.AgentInstance, zone string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				ocsAgent.Identity = meta.SINGLE
				ocsAgent.Zone = ""
				ocsAgent.MasterAgent = nil
			}
		}()
		// Update info.
		ocsInfoList := []*sqlite.OcsInfo{
			{Name: constant.OCS_INFO_VERSION,
				Value: constant.VERSION},
			{Name: constant.OCS_INFO_OS,
				Value: global.Os},
			{Name: constant.OCS_INFO_ARCHITECTURE,
				Value: global.Architecture},
		}
		for _, info := range ocsInfoList {
			if err = s.updateInfo(tx, info); err != nil {
				return
			}
		}
		if err = s.updateIdentity(tx, meta.FOLLOWER); err != nil {
			return
		}
		if err = s.setZone(tx, zone); err != nil {
			return
		}

		// add agent to table
		agentInstance := meta.NewAgentInstanceByAgent(ocsAgent)
		if err = s.addAgent(tx, agentInstance, global.HomePath, global.Os, global.Architecture, secure.Public()); err != nil {
			return
		}
		if err = s.addAgent(tx, meta.NewAgentInstance(masterAgent.Ip, masterAgent.Port, masterAgent.Zone, meta.MASTER, masterAgent.Version), "", "", "", ""); err != nil {
			return
		}

		ocsAgent.MasterAgent = meta.NewAgentWithZoneByAgentInfo(&masterAgent, masterAgent.Zone)
		return
	})
}

func (s *AgentService) BeSingleAgent() error {
	if ocsAgent.IsUnidentified() {
		ocsAgent.Identity = meta.SINGLE
		return nil
	}

	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		zone := ocsAgent.Zone
		identity := ocsAgent.Identity
		defer func() {
			if err != nil {
				ocsAgent.Zone = zone
				ocsAgent.Identity = identity
			}
		}()

		if err := s.deleteAllAgents(tx); err != nil {
			return err
		}
		if err := s.setZone(tx, ""); err != nil {
			return err
		}
		if err := s.updateIdentity(tx, meta.SINGLE); err != nil {
			return err
		}
		if err := secure.UpdateObPasswordInTransaction(tx, ""); err != nil {
			return err
		}
		ocsAgent.MasterAgent = nil
		return nil
	})
}

func (s *AgentService) BeScalingOutAgent(zone string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) (err error) {
		identity := ocsAgent.Identity
		defer func() {
			if err != nil {
				ocsAgent.Identity = identity
				ocsAgent.Zone = ""
			}
		}()
		// Update info.
		ocsInfoList := []*sqlite.OcsInfo{
			{Name: constant.OCS_INFO_VERSION, Value: constant.VERSION},
			{Name: constant.OCS_INFO_OS, Value: runtime.GOOS},
			{Name: constant.OCS_INFO_ARCHITECTURE, Value: runtime.GOARCH},
		}
		for _, info := range ocsInfoList {
			if err = s.updateInfo(tx, info); err != nil {
				return
			}
		}
		if err = s.setZone(tx, zone); err != nil {
			return
		}
		err = s.updateIdentity(tx, meta.SCALING_OUT)

		agentInstance := meta.NewAgentInstanceByAgent(ocsAgent)
		if err = s.addAgent(tx, agentInstance, global.HomePath, runtime.GOOS, runtime.GOARCH, secure.Public()); err != nil {
			return
		}
		return
	})
}

func (s *AgentService) SyncAgentData() (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	var agents []oceanbase.AllAgent
	err = oceanbaseDb.Model(&oceanbase.AllAgent{}).Find(&agents).Error
	if err != nil {
		return err
	}

	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return sqliteDb.Transaction(func(tx *gorm.DB) (err error) {
		if err = s.deleteAllAgents(tx); err != nil {
			return
		}

		for _, agent := range agents {
			agentInstance := meta.NewAgentInstance(agent.Ip, agent.Port, agent.Zone, meta.AgentIdentity(agent.Identity), agent.Version)
			if err = s.addAgent(tx, agentInstance, agent.HomePath, agent.Os, agent.Architecture, agent.PublicKey); err != nil {
				return
			}
			if err = s.UpdateAgentOBPortWithTx(tx, agentInstance, agent.MysqlPort, agent.RpcPort); err != nil {
				return
			}
			if ocsAgent.Equal(agentInstance) {
				if err = s.updateIdentity(tx, agentInstance.GetIdentity()); err != nil {
					return
				}
			}
			if agentInstance.IsMasterAgent() {
				ocsAgent.MasterAgent = meta.NewAgentWithZoneByAgentInfo(agentInstance, agent.Zone)
			}
		}
		return nil
	})

}

func (s *AgentService) GetIP() (ip string, err error) {
	err = getOCSInfo(constant.OCS_INFO_IP, &ip)
	return
}

func getOCSInfo(key string, value interface{}) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.OcsInfo{}).Select("value").Where("name=?", key).First(value).Error
	return
}
