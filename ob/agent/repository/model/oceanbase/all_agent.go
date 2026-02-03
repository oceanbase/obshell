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

package oceanbase

import "github.com/oceanbase/obshell/ob/agent/repository/model/bo"

type AllAgent struct {
	Ip           string `gorm:"primaryKey;type:varchar(64);not null" json:"ip"`
	Port         int    `gorm:"primaryKey;type:bigint(20);not null" json:"port"`
	Identity     string `gorm:"type:varchar(64);not null" json:"identity"`
	Version      string `gorm:"type:varchar(64);not null" json:"version"`
	Os           string `gorm:"type:varchar(64);not null" json:"os"`
	Architecture string `gorm:"type:varchar(64);not null" json:"architecture"`
	Zone         string `gorm:"type:varchar(64);not null" json:"zone"`
	MysqlPort    int    `gorm:"type:bigint(20)" json:"mysql_port"`
	RpcPort      int    `gorm:"type:bigint(20)" json:"rpc_port"`
	HomePath     string `gorm:"type:text" json:"home_path"`
	PublicKey    string `gorm:"type:text" json:"public_key"`
}

func (a *AllAgent) ToBo() *bo.AllAgent {
	return &bo.AllAgent{
		Ip:           a.Ip,
		Port:         a.Port,
		Identity:     a.Identity,
		Version:      a.Version,
		Os:           a.Os,
		Architecture: a.Architecture,
		Zone:         a.Zone,
		MysqlPort:    a.MysqlPort,
		RpcPort:      a.RpcPort,
		HomePath:     a.HomePath,
		PublicKey:    a.PublicKey,
	}
}
