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

package constant

const (
	COMMAND_HOST_TYPE                 string = "systemd-detect-virt"
	COMMAND_OS_NAME                   string = "cat /etc/os-release | grep \"^ID=\" | cut -f2 -d="
	COMMAND_OS_RELEASE                string = "cat /etc/os-release | grep \"^VERSION_ID=\" | cut -f2 -d="
	COMMAND_CPU_PHYSICAL_CORES        string = "cat /proc/cpuinfo | grep \"physical id\" | sort | uniq | wc -l"
	COMMAND_CPU_LOGIC_CORES           string = "cat /proc/cpuinfo | grep \"processor\" | wc -l"
	COMMAND_CPU_MODEL                 string = "cat /proc/cpuinfo | grep name | cut -f2 -d: | uniq"
	COMMAND_CPU_FREQUENCY             string = "cat /proc/cpuinfo | grep MHz | cut -f2 -d: | uniq"
	COMMAND_MEMORY_TOTAL              string = "cat /proc/meminfo | grep MemTotal | cut -f2 -d: | uniq"
	COMMAND_MEMORY_FREE               string = "cat /proc/meminfo | grep MemFree | cut -f2 -d: | uniq"
	COMMAND_MEMORY_AVAILABLE          string = "cat /proc/meminfo | grep MemAvailable | cut -f2 -d: | uniq"
	COMMAND_ULIMIT_NOFILE             string = "ulimit -n"
	COMMAND_ULIMIT_MAX_USER_PROCESSES string = "ulimit -u"
	COMMAND_DF                        string = "df -h | grep -v Filesystem"
)
