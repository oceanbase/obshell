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

package secure

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultSSOTokenExpirySec  = 60
	defaultSSOTokenGCInterval = 30 * 60
)

var (
	ssoTokenMgr *SSOTokenManager
)

func initSSOTokenManager() error {
	expirySec := defaultSSOTokenExpirySec
	var expiryStr string
	if err := getOCSConfig(constant.CONFIG_SSO_TOKEN_EXPIRY_SEC, &expiryStr); err == nil {
		if v, err := strconv.Atoi(strings.TrimSpace(expiryStr)); err == nil && v > 0 {
			expirySec = v
		}
	}
	gcSec := defaultSSOTokenGCInterval
	var gcStr string
	if err := getOCSConfig(constant.CONFIG_SSO_TOKEN_GC_INTERVAL_SEC, &gcStr); err == nil {
		if v, err := strconv.Atoi(strings.TrimSpace(gcStr)); err == nil && v > 0 {
			gcSec = v
		}
	}
	ssoTokenMgr = NewSSOTokenManager(time.Duration(expirySec)*time.Second, time.Duration(gcSec)*time.Second)
	return nil
}

// SSOToken represents a one-time SSO jump token.
type SSOToken struct {
	Token     string
	CreatedAt time.Time
	Used      bool
}

// SSOTokenManager manages one-time SSO tokens.
type SSOTokenManager struct {
	tokens         map[string]*SSOToken
	mutex          sync.RWMutex
	expiryDuration time.Duration
	cleanupTicker  *time.Ticker
	stopChan       chan bool
}

// NewSSOTokenManager creates a new SSO token manager.
func NewSSOTokenManager(expiry time.Duration, gcInterval time.Duration) *SSOTokenManager {
	m := &SSOTokenManager{
		tokens:         make(map[string]*SSOToken),
		expiryDuration: expiry,
		cleanupTicker:  time.NewTicker(gcInterval),
		stopChan:       make(chan bool),
	}
	go m.startCleanup()
	return m
}

func (m *SSOTokenManager) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSSOToken creates a new one-time SSO token.
func (m *SSOTokenManager) createSSOToken() (string, error) {
	token, err := m.generateToken()
	if err != nil {
		return "", err
	}
	now := time.Now()
	m.mutex.Lock()
	m.tokens[token] = &SSOToken{
		Token:     token,
		CreatedAt: now,
		Used:      false,
	}
	m.mutex.Unlock()
	log.Infof("create SSO token, expires in %v", m.expiryDuration)
	return token, nil
}

// validateSSOToken checks token exists, not used and not expired (read-only, does not consume).
func (m *SSOTokenManager) validateSSOToken(token string) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entry, exists := m.tokens[token]
	if !exists {
		return errors.Occur(errors.ErrSecuritySSOTokenNotFound)
	}
	if entry.Used {
		return errors.Occur(errors.ErrSecuritySSOTokenAlreadyUsed)
	}
	if time.Since(entry.CreatedAt) > m.expiryDuration {
		return errors.Occur(errors.ErrSecuritySSOTokenExpired)
	}
	return nil
}

// ExchangeSSOToken validates the token, marks it used, and returns nil if valid.
func (m *SSOTokenManager) exchangeSSOToken(token string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entry, exists := m.tokens[token]
	if !exists {
		return errors.Occur(errors.ErrSecuritySSOTokenNotFound)
	}
	if entry.Used {
		delete(m.tokens, token)
		return errors.Occur(errors.ErrSecuritySSOTokenAlreadyUsed)
	}
	if time.Since(entry.CreatedAt) > m.expiryDuration {
		delete(m.tokens, token)
		return errors.Occur(errors.ErrSecuritySSOTokenExpired)
	}
	entry.Used = true
	return nil
}

func (m *SSOTokenManager) cleanExpired() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	now := time.Now()
	count := 0
	for k, v := range m.tokens {
		if v.Used || now.Sub(v.CreatedAt) > m.expiryDuration {
			delete(m.tokens, k)
			count++
		}
	}
	return count
}

func (m *SSOTokenManager) startCleanup() {
	for {
		select {
		case <-m.cleanupTicker.C:
			if n := m.cleanExpired(); n > 0 {
				log.Infof("cleaned %d expired or used SSO tokens", n)
			}
		case <-m.stopChan:
			m.cleanupTicker.Stop()
			return
		}
	}
}

// CreateSSOToken creates a new one-time SSO token (package-level).
func CreateSSOToken() (string, error) {
	if ssoTokenMgr == nil {
		return "", errors.Occur(errors.ErrSecuritySSOTokenManagerNotReady)
	}
	return ssoTokenMgr.createSSOToken()
}

// ValidateSSOToken checks that the token exists, is not used and not expired (read-only). Used by middleware.
func ValidateSSOToken(token string) error {
	if ssoTokenMgr == nil {
		return errors.Occur(errors.ErrSecuritySSOTokenManagerNotReady)
	}
	return ssoTokenMgr.validateSSOToken(token)
}

// ExchangeSSOToken validates and consumes the token (package-level).
func ExchangeSSOToken(token string) error {
	if ssoTokenMgr == nil {
		return errors.Occur(errors.ErrSecuritySSOTokenManagerNotReady)
	}
	return ssoTokenMgr.exchangeSSOToken(token)
}
