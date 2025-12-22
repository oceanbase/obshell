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

package credential

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	obmodel "github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	oceanbase "github.com/oceanbase/obshell/ob/model/oceanbase"
)

type CredentialService struct{}

func (s *CredentialService) Create(credential *obmodel.ProfileCredential) error {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return errors.Wrap(err, "get ocs instance failed")
	}
	return db.Create(credential).Error
}

func (s *CredentialService) GetByID(id int64) (*obmodel.ProfileCredential, error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, errors.Wrap(err, "get ocs instance failed")
	}
	var credential obmodel.ProfileCredential
	err = db.Where("id = ?", id).First(&credential).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "get credential failed")
	}
	return &credential, nil
}

func (s *CredentialService) Update(credential *obmodel.ProfileCredential) error {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return errors.Wrap(err, "get ocs instance failed")
	}
	return db.Model(credential).Updates(map[string]interface{}{
		"name":        credential.Name,
		"secret":      credential.Secret,
		"description": credential.Description,
	}).Error
}

func (s *CredentialService) Delete(id int64) error {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return errors.Wrap(err, "get ocs instance failed")
	}
	// Hard delete: directly delete from database
	return db.Delete(&obmodel.ProfileCredential{}, id).Error
}

func (s *CredentialService) BatchDelete(ids []int64) error {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return errors.Wrap(err, "get ocs instance failed")
	}
	// Hard delete: directly delete from database
	return db.Where("id IN ?", ids).Delete(&obmodel.ProfileCredential{}).Error
}

func (s *CredentialService) List(query *ListQuery) ([]obmodel.ProfileCredential, int64, error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, 0, errors.Wrap(err, "get ocs instance failed")
	}

	queryBuilder := db.Model(&obmodel.ProfileCredential{})

	// Apply filters
	if query.CredentialId > 0 {
		queryBuilder = queryBuilder.Where("id = ?", query.CredentialId)
	}
	if query.TargetType != "" {
		queryBuilder = queryBuilder.Where("access_target = ?", query.TargetType)
	}
	if query.KeyWord != "" {
		// Search in name, username, ip:port
		keyWord := "%" + query.KeyWord + "%"
		queryBuilder = queryBuilder.Where(
			"name LIKE ? OR secret LIKE ?",
			keyWord, keyWord,
		)
	}

	// Get total count
	var total int64
	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "count credentials failed")
	}

	// Apply sorting
	if query.Sort != "" {
		sortOrder := "ASC"
		if query.SortOrder == "desc" || query.SortOrder == "DESC" {
			sortOrder = "DESC"
		}
		queryBuilder = queryBuilder.Order(query.Sort + " " + sortOrder)
	} else {
		queryBuilder = queryBuilder.Order("id DESC")
	}

	// Apply pagination
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		queryBuilder = queryBuilder.Offset(offset).Limit(query.PageSize)
	}

	var credentials []obmodel.ProfileCredential
	if err := queryBuilder.Find(&credentials).Error; err != nil {
		return nil, 0, errors.Wrap(err, "list credentials failed")
	}

	return credentials, total, nil
}

func (s *CredentialService) GetByIDs(ids []int64) ([]obmodel.ProfileCredential, error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, errors.Wrap(err, "get ocs instance failed")
	}
	var credentials []obmodel.ProfileCredential
	err = db.Where("id IN ?", ids).Find(&credentials).Error
	if err != nil {
		return nil, errors.Wrap(err, "get credentials by ids failed")
	}
	return credentials, nil
}

// GetByHost checks if a host (ip) already has a credential
// Returns the credential if exists, nil if not exists
// Uses optimized LIKE query to match IP in targets array
func (s *CredentialService) GetByHost(ip string) (*obmodel.ProfileCredential, error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, errors.Wrap(err, "get ocs instance failed")
	}

	// Use LIKE query with precise patterns to match IP in targets array
	// Secret format: {"username":"xxx","targets":[{"ip":"xxx","port":xxx},...],...}
	// Match patterns:
	// 1. "ip":"xxx" - IP in target object
	var credential obmodel.ProfileCredential
	err = db.Model(&obmodel.ProfileCredential{}).Where("access_target = ? AND secret LIKE ?",
		"HOST",
		fmt.Sprintf(`%%"ip":"%s"%%`, ip), // Match "ip":"xxx" in target object
	).First(&credential).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "query credential by host failed")
	}

	// Verify the match by parsing JSON to ensure accuracy
	var secretData struct {
		Targets []oceanbase.Target `json:"targets"`
	}
	if err := json.Unmarshal([]byte(credential.Secret), &secretData); err != nil {
		// If JSON parsing fails, return nil (treat as not found)
		return nil, nil
	}

	// Double-check: verify IP is actually in targets array
	for _, target := range secretData.Targets {
		if target.IP == ip {
			return &credential, nil
		}
	}

	// IP not found in targets (false positive from LIKE query)
	return nil, nil
}

// GetByName checks if a credential with the same name already exists
// Returns the credential if exists, nil if not exists
func (s *CredentialService) GetByName(name string) (*obmodel.ProfileCredential, error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, errors.Wrap(err, "get ocs instance failed")
	}

	var credential obmodel.ProfileCredential
	err = db.Model(&obmodel.ProfileCredential{}).Where("name = ?", name).First(&credential).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "query credential by name failed")
	}

	return &credential, nil
}

// CheckHostExistsInAllAgent checks if the host (ip) exists in all_agent table
// Returns true if exists, false if not exists
func (s *CredentialService) CheckHostExistsInAllAgent(ip string) (bool, error) {
	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return false, errors.Wrap(err, "get ocs instance failed")
	}

	var count int64
	err = db.Model(&obmodel.AllAgent{}).Where("ip = ?", ip).Count(&count).Error
	if err != nil {
		return false, errors.Wrap(err, "check host in all_agent failed")
	}

	return count > 0, nil
}

// ListAllCredentialsTx returns all credentials using given transaction
func (s *CredentialService) ListAllCredentialsTx(tx *gorm.DB) ([]obmodel.ProfileCredential, error) {
	if tx == nil {
		return nil, errors.Occur(errors.ErrCommonIllegalArgument, "tx is required")
	}
	var credentials []obmodel.ProfileCredential
	if err := tx.Find(&credentials).Error; err != nil {
		return nil, errors.Wrap(err, "query credentials failed")
	}
	return credentials, nil
}

// UpdateCredentialSecretTx updates secret of given credential id within tx
func (s *CredentialService) UpdateCredentialSecretTx(tx *gorm.DB, id int64, secret string) error {
	if tx == nil {
		return errors.Occur(errors.ErrCommonIllegalArgument, "tx is required")
	}
	return tx.Model(&obmodel.ProfileCredential{}).
		Where("id = ?", id).
		Update("secret", secret).Error
}

// SaveCredentialAESKeyTx saves encoded AES key into ocs_config within tx
func (s *CredentialService) SaveCredentialAESKeyTx(tx *gorm.DB, encodedKey string) error {
	if tx == nil {
		return errors.Occur(errors.ErrCommonIllegalArgument, "tx is required")
	}
	ocsConfig := obmodel.OcsConfig{
		Name:  constant.CREDENTIAL_AES_KEY_CONFIG,
		Value: encodedKey,
		Info:  "AES key for credential passphrase encryption",
	}
	return tx.Save(&ocsConfig).Error
}

type ListQuery struct {
	CredentialId int64
	TargetType   string
	KeyWord      string
	Page         int
	PageSize     int
	Sort         string
	SortOrder    string
}
