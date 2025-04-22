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

import "sync"

type PasswordMap struct {
	m sync.Map
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
