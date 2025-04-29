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

package system

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/json"
	"github.com/oceanbase/obshell/agent/lib/path"
)

type SCN struct {
	Val int64 `json:"val"`
}

type TenantKey struct {
	TenantId int `json:"tenant_id"`
}

type RestoreWindows struct {
	Windows []RestoreWindow `json:"restore_windows"`
}

type RestoreWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ArchiveInfo contains the information of archive log, which only contains the key, start scn and checkpoint scn but not the display time of scn.
type ArchiveInfo struct {
	Key           TenantKey `json:"key"`
	StartSCN      SCN       `json:"start_scn"`
	CheckPointSCN SCN       `json:"checkpoint_scn"`
}

type BackupSet struct {
	TenantKey
	BackupSetID         int  `json:"backup_set_id"`
	PlusArchivelog      bool `json:"plus_archivelog"`
	PrevFullBackupSetID int  `json:"prev_full_backup_set_id"`
	PrevIncBackupSetID  int  `json:"prev_inc_backup_set_id"`
	StartReplaySCN      SCN  `json:"start_replay_scn"`
	MinRestoreSCN       SCN  `json:"min_restore_scn"`
}

func ExecCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	var output strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(stdoutStderr))
	for scanner.Scan() {
		line := scanner.Text()
		output.WriteString(line)
		output.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return output.String(), nil
}

func formateBackupInfo(context string) []string {
	var infos []string
	re := regexp.MustCompile(`\|\s*\d+\|(.*?)\n`)
	matches := re.FindAllStringSubmatch(context, -1)
	for _, match := range matches {
		info := match[1]
		fixedStr := regexp.MustCompile(`([\{,])\s*(\w+):`).ReplaceAllString(info, `$1"$2":`)
		fixedStr = regexp.MustCompile(`:\s*true\b`).ReplaceAllString(fixedStr, `: true`)
		fixedStr = regexp.MustCompile(`:\s*false\b`).ReplaceAllString(fixedStr, `: false`)
		fixedStr = regexp.MustCompile(`:\s*null\b`).ReplaceAllString(fixedStr, `: null`)
		infos = append(infos, fixedStr)
	}
	return infos
}

func getLogPointAndDataSet(clogCtx, dataCtx string) ([][]int64, map[int]*BackupSet, error) {
	clogSet := formateBackupInfo(clogCtx)
	var logPointSet [][]int64

	for _, logPoint := range clogSet {
		var logPointData ArchiveInfo
		if err := json.Unmarshal([]byte(logPoint), &logPointData); err != nil {
			return nil, nil, errors.Wrap(err, "Failed to parse logPoint data")
		}
		logPointSet = append(logPointSet, []int64{logPointData.StartSCN.Val, logPointData.CheckPointSCN.Val})
	}

	dataSet := make(map[int]*BackupSet)
	for _, data := range formateBackupInfo(dataCtx) {
		var backupSet BackupSet
		if err := json.Unmarshal([]byte(data), &backupSet); err != nil {
			return nil, nil, errors.Wrap(err, "Failed to parse backupSet data")
		}
		dataSet[backupSet.BackupSetID] = &backupSet
	}
	return logPointSet, dataSet, nil
}

func getRestoreWindows(dataURI, logURI string) ([][2]int64, error) {
	log.Info("Get archive log context")
	archiveLogCtx, err := getOBAdminCtxByURI(logURI)
	if err != nil {
		return nil, errors.Wrapf(err, "execute archive log command failed")
	}

	log.Info("Get data backup context")
	dataCtx, err := getOBAdminCtxByURI(dataURI)
	if err != nil {
		return nil, errors.Wrapf(err, "execute data backup command failed")
	}

	logPointSet, dataSet, err := getLogPointAndDataSet(archiveLogCtx, dataCtx)
	if err != nil {
		return nil, err
	}

	// sort logPointSet by start scn
	sort.Slice(logPointSet, func(i, j int) bool {
		return logPointSet[i][0] < logPointSet[j][0]
	})

	var restoreWindows [][2]int64
	for _, data := range dataSet {
		if data.PlusArchivelog {
			// plugs 备份暂未实现
			continue
		}

		if data.PrevFullBackupSetID > 0 && dataSet[data.PrevFullBackupSetID] == nil {
			continue
		}
		if data.PrevIncBackupSetID > 0 && dataSet[data.PrevIncBackupSetID] == nil {
			continue
		}

		for i := 0; i < len(logPointSet); {
			if !(logPointSet[i][0] <= data.StartReplaySCN.Val && data.StartReplaySCN.Val <= logPointSet[i][1]) {
				i++
				continue
			}
			for (i < len(logPointSet)-1) && (logPointSet[i+1][0] == logPointSet[i][1]) {
				i++
			}
			restoreWindows = append(restoreWindows, [2]int64{data.MinRestoreSCN.Val, logPointSet[i][1]})
			break
		}
	}

	log.Infof("restoreWindows: %+v", restoreWindows)
	return mergeWindows(restoreWindows), nil
}

func mergeWindows(intervals [][2]int64) [][2]int64 {
	ans := make([][2]int64, 0)
	slices.SortFunc(intervals, func(a, b [2]int64) int {
		return int(a[0] - b[0])
	})

	l, r := intervals[0][0], intervals[0][1]
	for _, inte := range intervals {
		if inte[0] > r {
			ans = append(ans, [2]int64{l, r})
			l, r = inte[0], inte[1]
		} else if inte[1] > r {
			r = inte[1]
		}
	}

	return append(ans, [2]int64{l, r})
}

func checkRestoreTime(dataURI, logURI string, scn int64) (bool, error) {
	restoreWindows, err := getRestoreWindows(dataURI, logURI)
	if err != nil {
		return false, err
	}

	for _, window := range restoreWindows {
		if window[0] <= scn && scn <= window[1] {
			return true, nil
		}
	}
	return false, nil
}

func CheckRestoreTime(dataURI, logURI string, scn int64) (err error) {
	canRestore, err := checkRestoreTime(dataURI, logURI, scn)
	if err != nil {
		return errors.Wrapf(err, "check restore time")
	}
	if !canRestore {
		return errors.New("restore time is not valid")
	}
	return nil
}

func getOBAdminCtxByURI(uri string) (string, error) {
	storage, err := GetStorageInterfaceByURI(uri)
	if err != nil {
		return "", errors.Wrapf(err, "get storage interface failed")
	}
	cmd := newOBAdminCommand(storage)
	return ExecCommand(cmd)
}

func newOBAdminCommand(storage StorageInterface) string {
	cmd := fmt.Sprintf("export LD_LIBRARY_PATH='%s/lib'; %s dump_backup -q -d '%s'", global.HomePath, path.OBAdmin(), storage.GenerateURIWhitoutParams())
	if storage.GenerateQueryParams() != "" {
		cmd += fmt.Sprintf(" -s '%s'", storage.GenerateQueryParams())
	}
	return cmd
}

func GetRestoreWindows(dataURI, logURI string) (*RestoreWindows, error) {
	windows, err := getRestoreWindows(dataURI, logURI)
	if err != nil {
		return nil, err
	}

	res := new(RestoreWindows)
	for _, window := range windows {
		res.Windows = append(res.Windows, RestoreWindow{
			StartTime: time.Unix(0, window[0]),
			EndTime:   time.Unix(0, window[1]),
		})
	}

	return res, err
}
