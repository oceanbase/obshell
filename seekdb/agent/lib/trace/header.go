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

package trace

import (
	"crypto/rand"
	"fmt"
	"net/http"
)

const (
	TraceIdHeader = "X-OBSHELL-Trace-ID"
)

func GetTraceId(request *http.Request) string {
	traceId := request.Header.Get(TraceIdHeader)
	// If no traceId passed, generate one.
	if traceId == "" {
		traceId = RandomTraceId()
	}
	return traceId
}

func RandomTraceId() string {
	n := 8
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}
