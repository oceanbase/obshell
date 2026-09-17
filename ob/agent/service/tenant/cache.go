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

package tenant

import (
	"context"
	"hash/fnv"
	"sync"
)

type PasswordMap struct {
	m sync.Map
	// Fixed stripes bound memory usage while serializing validation and storage
	// for the same tenant. A plain sync.Map only serializes individual stores.
	updateLocks [64]passwordUpdateLock
}

type passwordUpdateLock struct {
	once  sync.Once
	token chan struct{}
}

func (pm *PasswordMap) SetValidated(ctx context.Context, key, value string, validate func() error) error {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(key)); err != nil {
		return err
	}
	lock := &pm.updateLocks[hash.Sum32()%uint32(len(pm.updateLocks))]
	// A one-token channel permits cancellation while waiting for the same
	// tenant. sync.Once initializes it safely; no worker goroutine is needed.
	lock.once.Do(func() { lock.token = make(chan struct{}, 1) })
	select {
	case lock.token <- struct{}{}:
		defer func() { <-lock.token }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validate(); err != nil {
		return err
	}
	pm.Set(key, value)
	return nil
}

func (pm *PasswordMap) Set(key, value string) {
	pm.m.Store(key, value)
}

func (pm *PasswordMap) Get(key string) (string, bool) {
	value, ok := pm.m.Load(key)
	if !ok {
		return "", false
	}
	return value.(string), true
}

var globalPasswordMap *PasswordMap
var once sync.Once

func GetPasswordMap() *PasswordMap {
	once.Do(func() {
		globalPasswordMap = &PasswordMap{}
	})
	return globalPasswordMap
}
