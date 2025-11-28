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
	"sync"
	"time"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	log "github.com/sirupsen/logrus"
)

var (
	sessionMgr *SessionManager
)

func initSessionManager() error {
	// get the session timeout from the config
	var sessionTimeout int
	err := getOCSConfig(constant.CONFIG_SESSION_TIMEOUT, &sessionTimeout)
	if err != nil {
		sessionTimeout = 1800
	}
	var sessionGCInterval int
	err = getOCSConfig(constant.CONFIG_SESSION_GC_INTERVAL, &sessionGCInterval)
	if err != nil {
		sessionGCInterval = 30
	}
	var maxSessionCount int
	err = getOCSConfig(constant.CONFIG_MAX_SESSION_COUNT, &maxSessionCount)
	if err != nil {
		maxSessionCount = 50
	}
	sessionMgr = NewSessionManager(time.Duration(sessionTimeout)*time.Second, time.Duration(sessionGCInterval)*time.Second, maxSessionCount)
	return nil
}

// Session represents a session.
type Session struct {
	ID         string                 // Session ID
	Data       map[string]interface{} // Session data
	CreatedAt  time.Time              // Created time
	LastAccess time.Time              // Last access time
	ExpiresAt  time.Time              // Expiration time
}

// SessionManager manages sessions.
type SessionManager struct {
	sessions        map[string]*Session // Session storage
	mutex           sync.RWMutex        // Read-write lock
	timeout         time.Duration       // Session timeout
	cleanupTicker   *time.Ticker        // Cleanup ticker
	stopChan        chan bool           // Stop cleanup signal
	maxSessionCount int                 // Max session count
}

// NewSessionManager creates a new session manager.
func NewSessionManager(timeout time.Duration, gcInterval time.Duration, maxSessionCount int) *SessionManager {
	sm := &SessionManager{
		sessions:        make(map[string]*Session),
		timeout:         timeout,
		cleanupTicker:   time.NewTicker(gcInterval),
		stopChan:        make(chan bool),
		maxSessionCount: maxSessionCount,
	}

	// Start background cleanup goroutine.
	go sm.startCleanup()

	return sm
}

// GenerateSessionID generates a random session ID.
func (sm *SessionManager) generateSessionID() (string, error) {
	// Generate 32 bytes random data.
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hexadecimal string.
	sessionID := hex.EncodeToString(bytes)
	return sessionID, nil
}

// CreateSession creates a new session.
func (sm *SessionManager) createSession() (*Session, error) {
	sessionID, err := sm.generateSessionID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	session := &Session{
		ID:         sessionID,
		Data:       make(map[string]interface{}),
		CreatedAt:  now,
		LastAccess: now,
		ExpiresAt:  now.Add(sm.timeout),
	}

	sm.mutex.Lock()
	sm.sessions[sessionID] = session
	sm.mutex.Unlock()
	// check the session count and control the session count
	if sm.Count() > sm.maxSessionCount {
		sm.InvalidateAllSessions()
	}

	return session, nil
}

// GetSession gets a session (with lazy-delete).
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, errors.Occur(errors.ErrSecurityAuthenticationSessionInvalid)
	}

	// Lazy-delete: check if expired.
	if time.Now().After(session.ExpiresAt) {
		sm.DeleteSession(sessionID)
		return nil, errors.Occur(errors.ErrSecurityAuthenticationSessionExpired)
	}

	// Update last access time and expiration time.
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	session.LastAccess = time.Now()
	session.ExpiresAt = time.Now().Add(sm.timeout)

	return session, nil
}

// DeleteSession deletes a specified session.
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	delete(sm.sessions, sessionID)
	sm.mutex.Unlock()
}

// GetAllSessions gets all active sessions.
func (sm *SessionManager) InvalidateAllSessions() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for sessionID := range sm.sessions {
		delete(sm.sessions, sessionID)
	}

	return nil
}

// Count gets the number of sessions.
func (sm *SessionManager) Count() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return len(sm.sessions)
}

// CleanExpiredSessions cleans up expired sessions (manually triggered).
func (sm *SessionManager) CleanExpiredSessions() int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	count := 0

	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, id)
			count++
		}
	}

	return count
}

// startCleanup cleans up expired sessions periodically.
func (sm *SessionManager) startCleanup() {
	for {
		select {
		case <-sm.cleanupTicker.C:
			count := sm.CleanExpiredSessions()
			if count > 0 {
				log.Infof("Cleaned %d expired sessions", count)
			}
		case <-sm.stopChan:
			sm.cleanupTicker.Stop()
			return
		}
	}
}

// Stop stops the session manager.
func (sm *SessionManager) Stop() {
	sm.stopChan <- true
}

// IsValid checks if a session is valid.
// This method will refresh the expiration time of the session.
func (sm *SessionManager) IsValid(sessionID string) bool {
	session, err := sm.GetSession(sessionID)
	return err == nil && session != nil
}

func ValidateSession(sessionID string) error {
	_, err := sessionMgr.GetSession(sessionID)
	return err
}

func CreateSession() (string, error) {
	session, err := sessionMgr.createSession()
	if err != nil {
		return "", err
	}
	return session.ID, nil
}

func InvalidateSession(sessionID string) {
	sessionMgr.DeleteSession(sessionID)
}

func InvalidateAllSessions() error {
	return sessionMgr.InvalidateAllSessions()
}
