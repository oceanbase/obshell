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

package alarm

import (
	"strings"

	alarmconstant "github.com/oceanbase/obshell/ob/agent/executor/alarm/constant"

	amlabels "github.com/prometheus/alertmanager/pkg/labels"
)

type Matcher struct {
	IsRegex bool   `json:"is_regex"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

func (m *Matcher) ToAmMatcher() (*amlabels.Matcher, error) {
	matchType := amlabels.MatchEqual
	if m.IsRegex {
		matchType = amlabels.MatchRegexp
	}
	return amlabels.NewMatcher(matchType, m.Name, m.Value)
}

func (m *Matcher) ExtractMatchedValues() []string {
	matchedValues := []string{m.Value}
	if m.IsRegex {
		matchedValues = strings.Split(m.Value, alarmconstant.RegexOR)
	}
	return matchedValues
}
