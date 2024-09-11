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
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/path"
)

type SCN struct {
	Val int64 `json:"val"`
}

type TenantKey struct {
	TenantId int `json:"tenant_id"`
}

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

func checkRestoreTime(clogCtx, dataCtx string, scn int64) (bool, error) {
	clogSet := formateBackupInfo(clogCtx)
	var logPointSet [][]int64
	for _, logPoint := range clogSet {
		var logPointData ArchiveInfo
		if err := json.Unmarshal([]byte(logPoint), &logPointData); err != nil {
			return false, errors.Wrap(err, "Failed to parse logPoint data")
		}
		logPointSet = append(logPointSet, []int64{logPointData.StartSCN.Val, logPointData.CheckPointSCN.Val})
	}

	dataSet := make(map[int]*BackupSet)
	for _, data := range formateBackupInfo(dataCtx) {
		var backupSet BackupSet
		if err := json.Unmarshal([]byte(data), &backupSet); err != nil {
			return false, errors.Wrap(err, "Failed to parse backupSet data")
		}
		dataSet[backupSet.BackupSetID] = &backupSet
	}

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

		if scn < data.MinRestoreSCN.Val {
			continue
		}

		for _, logPoint := range logPointSet {
			if logPoint[0] <= data.StartReplaySCN.Val && data.StartReplaySCN.Val <= logPoint[1] && scn <= logPoint[1] {
				return true, nil
			}
		}
	}
	return false, nil
}

func CheckRestoreTime(dataURI, logURI, scn string) (err error) {
	log.Info("Get archive log context")
	archiveLogCtx, err := getOBAdminCtxByURI(logURI)
	if err != nil {
		return errors.Wrapf(err, "execute archive log command failed")
	}

	log.Info("Get data backup context")
	dataCtx, err := getOBAdminCtxByURI(dataURI)
	if err != nil {
		return errors.Wrapf(err, "execute data backup command failed")
	}

	scnInt, err := strconv.ParseInt(scn, 10, 64)
	if err != nil {
		return errors.Wrapf(err, "parse scn")
	}

	canRestore, err := checkRestoreTime(archiveLogCtx, dataCtx, scnInt)
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
	cmd := fmt.Sprintf("%s dump_backup -q -d '%s'", path.OBAdmin(), storage.GenerateURIWhitoutParams())
	if storage.GenerateQueryParams() != "" {
		cmd += fmt.Sprintf(" -s '%s'", storage.GenerateQueryParams())
	}
	return cmd
}