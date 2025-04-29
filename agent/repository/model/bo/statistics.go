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

package bo

type OSInfo struct {
	OS      string `json:"os"`
	Version string `json:"version"`
}

type CpuInfo struct {
	Frequency     string `json:"frequency"`
	ModelName     string `json:"modelName"`
	LogicalCores  int    `json:"logicalCores"`
	PhysicalCores int    `json:"physicalCores"`
}

type HostBasicInfo struct {
	HostHash string `json:"hostHash"`
	HostType string `json:"hostType"`
}

type DiskInfo struct {
	Used       string `json:"used"`
	Total      string `json:"total"`
	MountHash  string `json:"mountHash"`
	DeviceName string `json:"deviceName"`
}

type MemoryInfo struct {
	Free      string `json:"free"`
	Total     string `json:"total"`
	Available string `json:"available"`
}

type UlimitInfo struct {
	NofileHard           string `json:"nofileHard"`
	NofileSoft           string `json:"nofileSoft"`
	MaxUserProcessesHard string `json:"maxUserProcessesHard"`
	MaxUserProcessesSoft string `json:"maxUserProcessesSoft"`
}

type HostInfo struct {
	OS     *OSInfo        `json:"os"`
	Cpu    *CpuInfo       `json:"cpuInfo"`
	Basic  *HostBasicInfo `json:"basic"`
	Disks  []DiskInfo     `json:"disks"`
	Memory *MemoryInfo    `json:"memoryInfo"`
	Ulimit *UlimitInfo    `json:"ulimit"`
}

type ObserverInfo struct {
	Type         string `json:"type"`
	Version      string `json:"version"`
	Revision     string `json:"revision"`
	CpuCount     int    `json:"cpuCount"`
	MemoryLimit  string `json:"memoryLimit"`
	DataFileSize string `json:"dataFileSize"`
	LogDiskSize  string `json:"logDiskSize"`
	HostHash     string `json:"hostHash"`
}

type ObclusterStatisticInfo struct {
	ClusterId        string         `json:"clusterId"`
	Reporter         string         `json:"reporter"`
	ObshellVersion   string         `json:"obshellVersion"`
	ObshellRevision  string         `json:"obshellRevision"`
	ReportTime       int64          `json:"reportTime"`
	TelemetryVersion int            `json:"telemetryVersion"`
	Hosts            []HostInfo     `json:"hosts"`
	Instances        []ObserverInfo `json:"instances"`
}
