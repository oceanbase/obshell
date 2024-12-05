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
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/oceanbase/obshell/agent/errors"
)

const (
	RELEASE_NULL = "null"
)

type VersionDep struct {
	Version           string        `yaml:"version"`
	Release           string        `yaml:"release"`
	Next              []*VersionDep `yaml:"next,omitempty"`
	CanBeUpgradedTo   []string      `yaml:"can_be_upgraded_to,flow,omitempty"`
	DirectComeFrom    []*VersionDep `yaml:"directComeFrom,omitempty"`
	Deprecated        bool          `yaml:"deprecated,omitempty"`
	RequireFromBinary interface{}   `yaml:"require_from_binary,flow,omitempty"`
	Precursor         *VersionDep   `yaml:"precursor,omitempty"`
	DirectUpgrade     bool          `yaml:"directUpgrade,omitempty"`
	DeprecatedInfo    []string
}

type RequireFromBinarySpec struct {
	Value        bool     `yaml:"value,omitempty"`
	WhenComeFrom []string `yaml:"when_come_from,omitempty"`
}

type UpgradeRoute struct {
	Version           string
	RequireFromBinary bool
}

type RouteNode struct {
	Version           string
	Release           string
	BuildVersion      string
	RequireFromBinary bool
	DeprecatedInfo    []string
}

type Repository struct {
	Version string
	Release string
}

func GetOBUpgradeRoute(currentRepo, targetRepo Repository, file string) ([]RouteNode, error) {
	data := []byte(file)
	var versionDep []VersionDep
	err := yaml.Unmarshal(data, &versionDep)
	if err != nil {
		log.WithError(err).Error("unmarshal failed")
		return nil, err
	}
	graph, err := Build(versionDep)
	if err != nil {
		log.WithError(err).Error("build graph failed")
		return nil, err
	}
	res, err := FindShortestUpgradePath(graph, currentRepo, targetRepo)
	if err != nil {
		log.WithError(err).Error("find shortest upgrade path failed")
		return nil, err
	}
	return res, nil
}

func Build(versionDeps []VersionDep) (map[string]*VersionDep, error) {
	nodeMap := make(map[string]*VersionDep)
	for idx, info := range versionDeps {
		if _, ok := nodeMap[info.Version]; ok {
			return nil, fmt.Errorf("version %s has duplicate", info.Version)
		}
		node := BuildVersionNode(&versionDeps[idx])
		nodeMap[info.Version] = node
	}
	for _, v := range nodeMap {
		nodeMap = BuildNeighbours(nodeMap, v, v.CanBeUpgradedTo, false)
		nodeMap = BuildNeighbours(nodeMap, v, v.CanBeUpgradedTo, true)
	}
	return nodeMap, nil
}

func BuildVersionNode(versionDep *VersionDep) *VersionDep {
	strList := strings.Split(versionDep.Version, "-")
	versionDep.Version = strList[0]
	if len(strList) > 1 {
		versionDep.Release = strList[1]
	} else {
		versionDep.Release = RELEASE_NULL
	}
	return versionDep
}

func BuildNeighbours(nodeMap map[string]*VersionDep, current *VersionDep, neighborVersions []string, direct bool) map[string]*VersionDep {
	for _, k := range neighborVersions {
		var node *VersionDep
		var ok bool
		if node, ok = nodeMap[k]; !ok {
			node = BuildVersionNode(&VersionDep{
				Version: k,
			})
		}
		if direct {
			node.DirectComeFrom = append(node.DirectComeFrom, node)
		}
		if node.Release == RELEASE_NULL {
			current.Next = append(current.Next, node)
		} else {
			current.Next = append([]*VersionDep{node}, current.Next...)
		}
	}
	return nodeMap
}

func GetNode(nodeMap map[string]*VersionDep, targetRepo Repository) *VersionDep {
	buildVersion := fmt.Sprintf("%s-%s", targetRepo.Version, targetRepo.Release)
	if node, ok := nodeMap[buildVersion]; ok {
		return node
	}
	var find *VersionDep
	for _, v := range nodeMap {
		if v.Version == targetRepo.Version {
			if find == nil || v.Release > find.Release {
				find = v
			}
		}
	}
	return find
}

func FindShortestUpgradePath(nodeMap map[string]*VersionDep, currentRepo, targetRepo Repository) ([]RouteNode, error) {
	startNode := GetNode(nodeMap, currentRepo)
	if startNode == nil {
		return nil, errors.New("Can not find the upgrade path from the current version to the target version")
	}
	queue := []*VersionDep{startNode}
	visited := mapset.NewSet(startNode)
	var finalNode *VersionDep
	for k := range nodeMap {
		nodeMap[k].Precursor = nil
	}

	for len(queue) > 0 {
		node := queue[len(queue)-1]
		queue = queue[0 : len(queue)-1]
		if node.Version == targetRepo.Version {
			if node.Release == targetRepo.Release {
				finalNode = node
				break
			}
			if node.Release == RELEASE_NULL {
				flag := false
				for _, v := range node.Next {
					if !visited.Contains(v) && v.Version == targetRepo.Version {
						flag = true
						v.Precursor = node
						queue = append(queue, v)
						visited.Add(v)
					}
				}
				if !flag {
					finalNode = node
				}
			}
		} else {
			for _, v := range node.Next {
				if !visited.Contains(v) {
					v.Precursor = node
					queue = append(queue, v)
					visited.Add(v)
				}
			}
		}
		if finalNode != nil {
			break
		}
	}
	p := finalNode
	var res []*VersionDep
	var pre *VersionDep
	for p != nil {
		res = append([]*VersionDep{p}, res...)
		pre = p.Precursor
		for pre != nil && pre.Precursor != nil && p.Version == pre.Version {
			pre = pre.Precursor
		}
		p = pre
	}
	n, i := len(res), 1
	for i < n {
		node := res[i]
		pre := res[i-1]
		for _, v := range node.DirectComeFrom {
			if v.Version == pre.Version && v.Release == pre.Release {
				node.DirectUpgrade = true
			}
		}
		i += 1
	}
	if len(res) == 1 {
		res = append([]*VersionDep{startNode}, res...)
	}
	if len(res) > 0 && res[len(res)-1].Deprecated {
		return nil, errors.New("target version is deprecated")
	}
	res = (AddDeprecatedInfo(nodeMap, res))
	return FormatRoute(res), nil
}

func AddDeprecatedInfo(nodeMap map[string]*VersionDep, res []*VersionDep) []*VersionDep {
	for _, finalNode := range res {
		for _, node := range nodeMap {
			if node.Version == finalNode.Version && node.Deprecated {
				finalNode.DeprecatedInfo = append(finalNode.DeprecatedInfo, node.Release)
			}
		}
	}
	return res
}

func FormatRoute(routes []*VersionDep) []RouteNode {
	var res []RouteNode
	preBarrier := routes[0]
	for _, v := range routes {
		// Format require_from_binary
		var requireFromBinary RequireFromBinarySpec
		if v.RequireFromBinary != nil {
			var ok bool
			if requireFromBinary.Value, ok = v.RequireFromBinary.(bool); !ok {
				if data, ok := v.RequireFromBinary.(map[string]interface{}); ok {
					if value, ok := data["value"]; ok {
						requireFromBinary.Value = value.(bool)
					}
					if whenComeFrom, ok := data["when_come_from"]; ok {
						requireFromBinary.WhenComeFrom = whenComeFrom.([]string)
					}
				}
			}
		}

		if len(requireFromBinary.WhenComeFrom) != 0 {
			var comeFrom bool
			for _, vv := range requireFromBinary.WhenComeFrom {
				if vv == preBarrier.Version || vv == fmt.Sprintf("%s-%s", preBarrier.Version, preBarrier.Release) {
					comeFrom = true
					break
				}
			}
			requireFromBinary.Value = requireFromBinary.Value && comeFrom
			if requireFromBinary.Value {
				preBarrier = v
			}
		}
		node := RouteNode{
			Version:           v.Version,
			Release:           v.Release,
			RequireFromBinary: requireFromBinary.Value,
			DeprecatedInfo:    v.DeprecatedInfo,
		}
		if v.Release == RELEASE_NULL {
			node.BuildVersion = v.Version
		} else {
			node.BuildVersion = fmt.Sprintf("%s-%s", v.Version, v.Release)
		}
		res = append(res, node)
	}

	// Supported since v4.2.1. Earlier upgrade paths prior to v4.1.0.0 are not required to be considered for compatibility.
	nodes := []RouteNode{res[len(res)-1]}
	for i := len(res) - 2; i >= 0; i-- {
		v := res[i]
		if v.RequireFromBinary {
			nodes = append([]RouteNode{v}, nodes...)
		}
	}
	nodes = append([]RouteNode{res[0]}, nodes...)
	return nodes
}
