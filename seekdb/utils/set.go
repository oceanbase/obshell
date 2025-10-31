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

package utils

func Difference[T comparable](a, b []T) ([]T, []T) {
	aMap := make(map[T]bool)
	bMap := make(map[T]bool)

	for _, v := range a {
		aMap[v] = true
	}

	for _, v := range b {
		bMap[v] = true
	}

	var onlyInA []T
	var onlyInB []T

	for _, v := range a {
		if !bMap[v] {
			onlyInA = append(onlyInA, v)
		}
	}

	for _, v := range b {
		if !aMap[v] {
			onlyInB = append(onlyInB, v)
		}
	}

	return onlyInA, onlyInB
}
