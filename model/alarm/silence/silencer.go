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

package silence

import (
	"strings"
	"time"

	alarmconstant "github.com/oceanbase/obshell/agent/executor/alarm/constant"
	"github.com/oceanbase/obshell/model/alarm"
	"github.com/oceanbase/obshell/model/oceanbase"

	ammodels "github.com/prometheus/alertmanager/api/v2/models"
	log "github.com/sirupsen/logrus"
)

type SilencerApiResponse struct {
	Id string `json:"id" binding:"required"`
}

type Status struct {
	State State `json:"state" binding:"required"`
}

type Silencer struct {
	Comment   string          `json:"comment" binding:"required"`
	CreatedBy string          `json:"created_by" binding:"required"`
	StartsAt  int64           `json:"starts_at" binding:"required"`
	EndsAt    int64           `json:"ends_at" binding:"required"`
	Matchers  []alarm.Matcher `json:"matchers" binding:"required"`
}

type SilencerResponse struct {
	Id        string                 `json:"id" binding:"required"`
	Instances []oceanbase.OBInstance `json:"instances" binding:"required"`
	Status    *Status                `json:"status" binding:"required"`
	UpdatedAt int64                  `json:"updated_at" binding:"required"`
	Rules     []string               `json:"rules" binding:"required"`
	Silencer  `json:",inline"`
}

type SilencerIdentity struct {
	Id string `json:"id" binding:"required"`
}

type SilencerParam struct {
	Id        string                 `json:"id,omitempty"`
	Instances []oceanbase.OBInstance `json:"instances" binding:"required"`
	Rules     []string               `json:"rules" binding:"required"`
	Silencer
}

func extractInstances(matcherMap map[string]alarm.Matcher) []oceanbase.OBInstance {
	instances := make([]oceanbase.OBInstance, 0)
	var matchedInstanceType oceanbase.OBInstanceType
	clusterMatcher, matchCluster := matcherMap[alarmconstant.LabelOBCluster]
	zoneMatcher, matchZone := matcherMap[alarmconstant.LabelOBZone]
	serverMatcher, matchServer := matcherMap[alarmconstant.LabelOBServer]
	tenantMatcher, matchTenant := matcherMap[alarmconstant.LabelOBTenant]
	if matchCluster {
		matchedInstanceType = oceanbase.TypeOBCluster
	}
	if matchZone {
		matchedInstanceType = oceanbase.TypeOBZone
	}
	if matchServer {
		matchedInstanceType = oceanbase.TypeOBServer
	}
	if matchTenant {
		matchedInstanceType = oceanbase.TypeOBTenant
	}
	switch matchedInstanceType {
	case oceanbase.TypeOBCluster:
		log.Debugf("Cluster matcher is: %v", clusterMatcher)
		clusterNames := clusterMatcher.ExtractMatchedValues()
		for _, clusterName := range clusterNames {
			instances = append(instances, oceanbase.OBInstance{
				Type:      oceanbase.TypeOBCluster,
				OBCluster: clusterName,
			})
		}
	case oceanbase.TypeOBZone:
		if !matchCluster {
			log.Error("Cluster matcher not exists")
			break
		} else if clusterMatcher.IsRegex {
			log.Error("Multiple cluster matches for zone matcher")
			break
		}
		log.Debugf("Cluster matcher is: %v", clusterMatcher)
		log.Debugf("Zone matcher is: %v", zoneMatcher)
		zoneNames := zoneMatcher.ExtractMatchedValues()
		for _, zone := range zoneNames {
			instances = append(instances, oceanbase.OBInstance{
				Type:      oceanbase.TypeOBZone,
				OBCluster: clusterMatcher.Value,
				OBZone:    zone,
			})
		}
	case oceanbase.TypeOBServer:
		if !matchCluster {
			log.Error("Cluster matcher not exists")
			break
		} else if clusterMatcher.IsRegex {
			log.Error("Multiple cluster matches for observer matcher")
			break
		}
		log.Debugf("Cluster matcher is: %v", clusterMatcher)
		log.Debugf("Server matcher is: %v", serverMatcher)
		serverIps := serverMatcher.ExtractMatchedValues()
		for _, serverIp := range serverIps {
			instances = append(instances, oceanbase.OBInstance{
				Type:      oceanbase.TypeOBServer,
				OBCluster: clusterMatcher.Value,
				OBServer:  serverIp,
			})
		}
	case oceanbase.TypeOBTenant:
		if !matchCluster {
			log.Error("Cluster matcher not exists")
			break
		} else if clusterMatcher.IsRegex {
			log.Error("Multiple cluster matches for obtenant matcher")
			break
		}
		log.Debugf("Cluster matcher is: %v", clusterMatcher)
		log.Debugf("Tenant matcher is: %v", tenantMatcher)
		tenantNames := tenantMatcher.ExtractMatchedValues()
		for _, tenant := range tenantNames {
			instances = append(instances, oceanbase.OBInstance{
				Type:      oceanbase.TypeOBTenant,
				OBCluster: clusterMatcher.Value,
				OBTenant:  tenant,
			})
		}
	}
	return instances
}

func NewSilencerResponse(gettableSilencer *ammodels.GettableSilence) *SilencerResponse {
	matchers := make([]alarm.Matcher, 0)
	matcherMap := make(map[string]alarm.Matcher)
	rules := make([]string, 0)
	for _, silenceMatcher := range gettableSilencer.Matchers {
		matcher := alarm.Matcher{
			IsRegex: *silenceMatcher.IsRegex,
			Name:    *silenceMatcher.Name,
			Value:   *silenceMatcher.Value,
		}
		matchers = append(matchers, matcher)
		matcherMap[matcher.Name] = matcher
		if matcher.Name == alarmconstant.LabelRuleName {
			rules = strings.Split(matcher.Value, alarmconstant.RegexOR)
		}
	}

	instances := extractInstances(matcherMap)
	silencer := &Silencer{
		Comment:   *gettableSilencer.Comment,
		CreatedBy: *gettableSilencer.CreatedBy,
		StartsAt:  time.Time(*gettableSilencer.StartsAt).Unix(),
		EndsAt:    time.Time(*gettableSilencer.EndsAt).Unix(),
		Matchers:  matchers,
	}
	silencerResponse := &SilencerResponse{
		Silencer:  *silencer,
		Id:        *gettableSilencer.ID,
		UpdatedAt: time.Time(*gettableSilencer.UpdatedAt).Unix(),
		Status: &Status{
			State: State(*gettableSilencer.Status.State),
		},
		Instances: instances,
		Rules:     rules,
	}
	return silencerResponse
}
