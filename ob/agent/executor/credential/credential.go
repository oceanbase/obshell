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
	"encoding/base64"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	executorcommon "github.com/oceanbase/obshell/ob/agent/executor/common"
	"github.com/oceanbase/obshell/ob/agent/lib/crypto"
	"github.com/oceanbase/obshell/ob/agent/lib/sshutil"
	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	obmodel "github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/secure"
	credentialservice "github.com/oceanbase/obshell/ob/agent/service/credential"
	"github.com/oceanbase/obshell/ob/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
	"github.com/oceanbase/obshell/ob/utils"
)

var (
	credentialService = credentialservice.CredentialService{}
)

const (
	TARGET_TYPE_HOST   = "HOST"
	AUTH_TYPE_PASSWORD = "PASSWORD"
	DEFAULT_SSH_PORT   = 22
)

// validateTarget validates a Target structure and normalizes it
// If port is 0 or not provided, default port 22 is used
// Returns: normalized Target, error
func validateTarget(target param.Target) (oceanbase.Target, error) {
	// Validate IP address
	if !utils.IsValidIp(target.IP) {
		return oceanbase.Target{}, errors.Occurf(errors.ErrCommonInvalidIp, target.IP)
	}

	// Set default port if not provided
	port := target.Port
	if port == 0 {
		port = DEFAULT_SSH_PORT
	}

	// Validate port range
	if port < 1 || port > 65535 {
		return oceanbase.Target{}, errors.Occurf(errors.ErrCredentialInvalidSshPort, strconv.Itoa(port))
	}

	return oceanbase.Target{
		IP:   target.IP,
		Port: port,
	}, nil
}

// validateTargets validates targets list and returns normalized targets
// Returns: []Target, error
func validateTargets(targets []param.Target) ([]oceanbase.Target, error) {
	if len(targets) == 0 {
		return nil, errors.Occur(errors.ErrCommonIllegalArgument, "targets list cannot be empty")
	}

	normalizedTargets := make([]oceanbase.Target, 0, len(targets))
	for _, target := range targets {
		normalizedTarget, err := validateTarget(target)
		if err != nil {
			return nil, err
		}
		normalizedTargets = append(normalizedTargets, normalizedTarget)
	}

	return normalizedTargets, nil
}

// validateSshCredentialProperty validates SSH credential property
// Returns validated targets list and error if validation fails
func validateSshCredentialProperty(sshProp *param.SshCredentialProperty) ([]oceanbase.Target, error) {
	// Validate and normalize targets
	targets, err := validateTargets(sshProp.Targets)
	if err != nil {
		return nil, err
	}

	// Validate username (basic check: not empty)
	if sshProp.Username == "" {
		return nil, errors.Occur(errors.ErrCommonIllegalArgument, "username cannot be empty")
	}

	// Validate authentication type
	if sshProp.Type != AUTH_TYPE_PASSWORD {
		return nil, errors.Occur(errors.ErrCredentialAuthTypeNotSupported)
	}

	return targets, nil
}

// validateCredentialUniqueness validates credential uniqueness constraints
// For create: excludeId should be 0
// For update: excludeId should be the current credential ID to exclude from checks
func validateCredentialUniqueness(targets []oceanbase.Target, name string, excludeId int64) error {
	// Check if each target host exists in all_agent table
	for _, target := range targets {
		hostExists, err := credentialService.CheckHostExistsInAllAgent(target.IP)
		if err != nil {
			return errors.Wrap(err, "check host in all_agent failed")
		}
		if !hostExists {
			return errors.Occurf(errors.ErrCredentialHostNotInAllAgent, target.IP)
		}
	}

	// Check if credential name already exists (excluding current credential if updating)
	existingCredentialByName, err := credentialService.GetByName(name)
	if err != nil {
		return errors.Wrap(err, "check credential name failed")
	}
	if existingCredentialByName != nil && (excludeId == 0 || existingCredentialByName.ID != excludeId) {
		return errors.Occurf(errors.ErrCredentialNameAlreadyExists, name)
	}

	// Check if each target host already has a credential (excluding current credential if updating)
	for _, target := range targets {
		existingCredential, err := credentialService.GetByHost(target.IP)
		if err != nil {
			return errors.Wrap(err, "check host credential failed")
		}
		if existingCredential != nil && (excludeId == 0 || existingCredential.ID != excludeId) {
			return errors.Occurf(errors.ErrCredentialHostAlreadyExists, target.IP)
		}
	}

	return nil
}

// CreateCredential creates a new credential with SSH validation and encryption
func CreateCredential(p *param.CreateCredentialParam) (*bo.Credential, error) {
	// Validate target type
	if p.TargetType != TARGET_TYPE_HOST {
		return nil, errors.Occur(errors.ErrCredentialTargetTypeNotSupported)
	}

	// Validate SSH credential property
	targets, err := validateSshCredentialProperty(&p.SshCredentialProperty)
	if err != nil {
		return nil, err
	}

	// Validate credential uniqueness (excludeId = 0 for create)
	err = validateCredentialUniqueness(targets, p.Name, 0)
	if err != nil {
		return nil, err
	}

	// Validate SSH connection for each target
	for _, target := range targets {
		err = sshutil.ValidateSSHConnection(
			target.IP,
			target.Port,
			p.SshCredentialProperty.Username,
			*p.SshCredentialProperty.Passphrase,
		)
		if err != nil {
			return nil, errors.WrapRetain(errors.ErrCredentialSSHValidationFailed, err)
		}
	}

	// Encrypt passphrase
	encryptedPassphrase, err := secure.EncryptCredentialPassphrase(*p.SshCredentialProperty.Passphrase)
	if err != nil {
		return nil, errors.WrapRetain(errors.ErrCredentialEncryptFailed, err)
	}

	// Create credential model
	credential := &obmodel.ProfileCredential{
		AccessTarget: p.TargetType,
		Name:         p.Name,
		Secret: SerializeSecret(
			p.SshCredentialProperty.Username,
			targets,
			p.SshCredentialProperty.Type,
			encryptedPassphrase,
		),
		Description: p.Description,
		Deleted:     false,
	}

	// Save to database
	err = credentialService.Create(credential)
	if err != nil {
		return nil, errors.Wrap(err, "create credential in database failed")
	}

	// get credential from database
	credential, err = credentialService.GetByID(credential.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get credential from database failed")
	}
	if credential == nil {
		return nil, errors.Occur(errors.ErrCredentialNotFound)
	}

	// Convert to BO
	return convertToBO(credential), nil
}

// GetCredential retrieves a credential by ID and converts to BO
func GetCredential(id int64) (*bo.Credential, error) {
	credential, err := credentialService.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "get credential from database failed")
	}
	if credential == nil {
		return nil, errors.Occur(errors.ErrCredentialNotFound)
	}

	return convertToBO(credential), nil
}

// UpdateCredential updates a credential with SSH validation and encryption
func UpdateCredential(id int64, p *param.UpdateCredentialParam) (*bo.Credential, error) {
	// Get existing credential
	existing, err := credentialService.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "get credential from database failed")
	}
	if existing == nil {
		return nil, errors.Occur(errors.ErrCredentialNotFound)
	}

	// Validate SSH credential property
	targets, err := validateSshCredentialProperty(&p.SshCredentialProperty)
	if err != nil {
		return nil, err
	}

	// Validate credential uniqueness (excludeId = id for update to exclude current credential)
	err = validateCredentialUniqueness(targets, p.Name, id)
	if err != nil {
		return nil, err
	}

	// Get passphrase for validation
	var passphrase string
	if *p.SshCredentialProperty.Passphrase != "" {
		passphrase = *p.SshCredentialProperty.Passphrase
	} else {
		// get passphrase from existing credential
		encryptedPassphrase, err := DeserializeSecret(existing.Secret)
		if err != nil {
			return nil, errors.Wrap(err, "deserialize credential secret failed")
		}
		passphrase, err = secure.DecryptCredentialPassphrase(encryptedPassphrase.Passphrase)
		if err != nil {
			return nil, errors.WrapRetain(errors.ErrCredentialDecryptFailed, err)
		}
	}

	// Validate SSH connection for each target
	for _, target := range targets {
		err = sshutil.ValidateSSHConnection(
			target.IP,
			target.Port,
			p.SshCredentialProperty.Username,
			passphrase,
		)
		if err != nil {
			return nil, errors.WrapRetain(errors.ErrCredentialSSHValidationFailed, err)
		}
	}

	// Encrypt passphrase
	var encryptedPassphrase string
	if *p.SshCredentialProperty.Passphrase != "" {
		encryptedPassphrase, err = secure.EncryptCredentialPassphrase(*p.SshCredentialProperty.Passphrase)
		if err != nil {
			return nil, errors.WrapRetain(errors.ErrCredentialEncryptFailed, err)
		}
	} else {
		// Extract encrypted passphrase and type from existing secret
		secretData, err := DeserializeSecret(existing.Secret)
		if err != nil {
			return nil, errors.Wrap(err, "deserialize credential secret failed")
		}
		encryptedPassphrase = secretData.Passphrase
	}

	// Serialize secret
	secret := SerializeSecret(
		p.SshCredentialProperty.Username,
		targets,
		p.SshCredentialProperty.Type,
		encryptedPassphrase,
	)

	// Update credential
	existing.Name = p.Name
	existing.Secret = secret
	existing.Description = p.Description

	err = credentialService.Update(existing)
	if err != nil {
		return nil, errors.Wrap(err, "update credential failed")
	}

	return convertToBO(existing), nil
}

// DeleteCredential deletes a credential (hard delete)
func DeleteCredential(id int64) error {
	// Check if credential exists
	credential, err := credentialService.GetByID(id)
	if err != nil {
		return errors.Wrap(err, "get credential from database failed")
	}
	if credential == nil {
		return errors.Occur(errors.ErrCredentialNotFound)
	}

	return credentialService.Delete(id)
}

// BatchDeleteCredential deletes multiple credentials (hard delete)
func BatchDeleteCredential(ids []int64) error {
	if len(ids) == 0 {
		return errors.Occur(errors.ErrCommonIllegalArgument, "credential id list is empty")
	}
	return credentialService.BatchDelete(ids)
}

// ListCredentials lists credentials with filtering, pagination, and sorting, and wraps paginated response.
func ListCredentials(query *param.ListCredentialQueryParam) (*bo.PaginatedCredentialResponse, error) {
	// Convert param to service query
	serviceQuery := &credentialservice.ListQuery{
		CredentialId: int64(query.CredentialId),
		TargetType:   query.TargetType,
		KeyWord:      query.KeyWord,
		Page:         query.Page,
		PageSize:     query.PageSize,
		Sort:         query.Sort,
		SortOrder:    query.SortOrder,
	}

	credentials, total, err := credentialService.List(serviceQuery)
	if err != nil {
		return nil, errors.Wrap(err, "list credentials failed")
	}

	// Convert to BO
	boCredentials := make([]bo.Credential, len(credentials))
	for i := range credentials {
		boCredentials[i] = *convertToBO(&credentials[i])
	}

	return &bo.PaginatedCredentialResponse{
		Contents: boCredentials,
		Page: bo.CustomPage{
			Number:        uint64(query.Page),
			Size:          uint64(query.PageSize),
			TotalPages:    executorcommon.CalculateTotalPages(uint64(total), uint64(query.PageSize)),
			TotalElements: uint64(total),
		},
	}, nil
}

// ValidateCredential validates a credential without storing it
func ValidateCredential(p *param.ValidateCredentialParam) (*bo.ValidationResult, error) {
	// Validate target type
	if p.TargetType != TARGET_TYPE_HOST {
		return nil, errors.Occur(errors.ErrCredentialTargetTypeNotSupported)
	}

	// Validate SSH credential property and parse targets
	// parseTarget is done here as part of parameter validation
	targets, err := validateSshCredentialProperty(&p.SshCredentialProperty)
	if err != nil {
		return nil, err
	}

	result := &bo.ValidationResult{
		TargetType:     p.TargetType,
		SucceededCount: 0,
		FailedCount:    0,
		Details:        make([]bo.ValidationDetail, 0),
	}

	// Validate SSH connection for each target
	// Note: targets have already been parsed and validated in parameter validation phase,
	// so we can directly use IP and port without re-parsing
	for _, target := range targets {
		err = sshutil.ValidateSSHConnection(
			target.IP,
			target.Port,
			p.SshCredentialProperty.Username,
			*p.SshCredentialProperty.Passphrase,
		)
		detail := bo.ValidationDetail{
			Target: bo.Target{
				IP:   target.IP,
				Port: target.Port,
			},
		}
		if err != nil {
			detail.ConnectionResult = bo.ConnectionResultConnectFailed
			detail.Message = "SSH validation failed: " + err.Error()
			result.FailedCount++
		} else {
			detail.ConnectionResult = bo.ConnectionResultSuccess
			detail.Message = ""
			result.SucceededCount++
		}

		result.Details = append(result.Details, detail)
	}

	return result, nil
}

// BatchValidateCredential validates multiple credentials by ID list
func BatchValidateCredential(credentialIds []int) ([]bo.ValidationResult, error) {
	// Convert []int to []int64
	ids := make([]int64, len(credentialIds))
	for i, id := range credentialIds {
		ids[i] = int64(id)
	}

	// Get credentials
	credentials, err := credentialService.GetByIDs(ids)
	if err != nil {
		return nil, errors.Wrap(err, "get credentials failed")
	}

	// Check if all credential IDs exist
	credentialMap := make(map[int64]*obmodel.ProfileCredential)
	for i := range credentials {
		credentialMap[credentials[i].ID] = &credentials[i]
	}

	// Find missing credential IDs
	var missingIds []int64
	for _, id := range ids {
		if _, exists := credentialMap[id]; !exists {
			missingIds = append(missingIds, id)
		}
	}

	// If any credential ID is missing, return error
	if len(missingIds) > 0 {
		return nil, errors.Occur(errors.ErrCredentialNotFound)
	}

	results := make([]bo.ValidationResult, 0, len(credentialIds))

	// Validate each credential
	for _, credentialId := range ids {
		credential := credentialMap[credentialId]
		result := bo.ValidationResult{
			CredentialId:   credentialId,
			TargetType:     credential.AccessTarget,
			SucceededCount: 0,
			FailedCount:    0,
			Details:        make([]bo.ValidationDetail, 0),
		}

		// Deserialize secret
		secretData, err := DeserializeSecret(credential.Secret)
		if err != nil {
			// Deserialization failed, create a detail with empty target
			result.Details = append(result.Details, bo.ValidationDetail{
				Target:           bo.Target{},
				ConnectionResult: bo.ConnectionResultConnectFailed,
				Message:          "failed to deserialize secret: " + err.Error(),
			})
			result.FailedCount = 1
			results = append(results, result)
			continue
		}

		// Decrypt passphrase
		passphrase, err := secure.DecryptCredentialPassphrase(secretData.Passphrase)
		if err != nil {
			// Decryption failed, create a detail with empty target
			result.Details = append(result.Details, bo.ValidationDetail{
				Target:           bo.Target{},
				ConnectionResult: bo.ConnectionResultConnectFailed,
				Message:          "failed to decrypt passphrase: " + err.Error(),
			})
			result.FailedCount = len(secretData.Targets)
			result.SucceededCount = 0
			results = append(results, result)
			continue
		}

		// Validate SSH connection for each target
		for _, target := range secretData.Targets {
			// Validate target fields
			if target.IP == "" || target.Port < 1 || target.Port > 65535 {
				// Invalid target format in stored credential, return error
				return nil, errors.Occur(errors.ErrCredentialSecretFormatInvalid, "invalid target format in credential: ip or port is invalid")
			}

			detail := bo.ValidationDetail{
				Target: bo.Target{
					IP:   target.IP,
					Port: target.Port,
				},
			}

			err = sshutil.ValidateSSHConnection(target.IP, target.Port, secretData.Username, passphrase)
			if err != nil {
				detail.ConnectionResult = bo.ConnectionResultConnectFailed
				detail.Message = "SSH validation failed: " + err.Error()
				result.FailedCount++
			} else {
				detail.ConnectionResult = bo.ConnectionResultSuccess
				detail.Message = ""
				result.SucceededCount++
			}

			result.Details = append(result.Details, detail)
		}

		results = append(results, result)
	}

	return results, nil
}

// convertToBO converts ProfileCredential model to BO
// If deserialization fails, logs a warning and returns empty values for SSH secret fields
func convertToBO(credential *obmodel.ProfileCredential) *bo.Credential {
	// Deserialize secret
	secretData, err := DeserializeSecret(credential.Secret)
	if err != nil {
		log.WithError(err).WithField("credential_id", credential.ID).Warn("failed to deserialize credential secret, returning empty values")
		// Return empty values for SSH secret fields
		secretData = &oceanbase.CredentialSecretData{
			Username: "",
			Targets:  []oceanbase.Target{},
		}
	}

	// Convert oceanbase.Target to bo.Target
	boTargets := make([]bo.Target, len(secretData.Targets))
	for i, target := range secretData.Targets {
		boTargets[i] = bo.Target{
			IP:   target.IP,
			Port: target.Port,
		}
	}

	boCredential := &bo.Credential{
		CredentialId: credential.ID,
		Name:         credential.Name,
		TargetType:   credential.AccessTarget,
		Description:  credential.Description,
		SshSecret: bo.SshSecret{
			Targets:  boTargets,
			Username: secretData.Username,
			Type:     secretData.SshType,
		},
		CreateTime: credential.CreateTime,
		UpdateTime: credential.UpdateTime,
	}

	return boCredential
}

// UpdateCredentialEncryptSecretKey updates AES key (Base64 encoded raw key) and re-encrypts all stored credentials
func UpdateCredentialEncryptSecretKey(p *param.UpdateCredentialEncryptSecretKeyParam) (err error) {
	if p == nil || p.AesKey == "" {
		return errors.Occur(errors.ErrCommonIllegalArgument, "key is required")
	}

	newKey, err := base64.StdEncoding.DecodeString(p.AesKey)
	if err != nil {
		return errors.Occur(errors.ErrCommonIllegalArgument, "invalid base64 key")
	}
	if l := len(newKey); l != 32 {
		return errors.Occur(errors.ErrCommonIllegalArgument, "invalid aes key length, must be 32 bytes")
	}

	oldKey, err := secure.GetCredentialAESKey()
	if err != nil {
		return err
	}

	db, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return errors.Wrap(err, "get ocs instance failed")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		credentials, err := credentialService.ListAllCredentialsTx(tx)
		if err != nil {
			return err
		}

		for i := range credentials {
			secretData, deserErr := DeserializeSecret(credentials[i].Secret)
			if deserErr != nil {
				return errors.Wrap(deserErr, "deserialize credential secret failed")
			}

			passphrase, decErr := secure.DecryptCredentialPassphraseWithKey(secretData.Passphrase, oldKey)
			if decErr != nil {
				return errors.Wrap(decErr, "decrypt credential passphrase failed")
			}

			newEncryptedPassphrase, encErr := secure.EncryptCredentialPassphraseWithKey(passphrase, newKey)
			if encErr != nil {
				return errors.Wrap(encErr, "encrypt credential passphrase failed")
			}

			newSecret := SerializeSecret(secretData.Username, secretData.Targets, secretData.SshType, newEncryptedPassphrase)
			if err := credentialService.UpdateCredentialSecretTx(tx, credentials[i].ID, newSecret); err != nil {
				return err
			}
		}

		encodedKey := crypto.CaesarBase64Encode(string(newKey), constant.CAESAR_SHIFT)
		if err := credentialService.SaveCredentialAESKeyTx(tx, encodedKey); err != nil {
			return err
		}
		return nil
	})
}
