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

package observer

import "testing"

func TestResolveClusterName(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		ip          string
		port        int
		want        string
	}{
		{
			name:        "prefer configured cluster name",
			clusterName: "seekdb-demo",
			ip:          "127.0.0.1",
			port:        2881,
			want:        "seekdb-demo",
		},
		{
			name: "fall back to endpoint",
			ip:   "127.0.0.1",
			port: 2881,
			want: "127.0.0.1:2881",
		},
		{
			name: "missing ip leaves name empty",
			port: 2881,
		},
		{
			name: "invalid port leaves name empty",
			ip:   "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveClusterName(tt.clusterName, tt.ip, tt.port); got != tt.want {
				t.Fatalf("resolveClusterName(%q, %q, %d) = %q, want %q", tt.clusterName, tt.ip, tt.port, got, tt.want)
			}
		})
	}
}
