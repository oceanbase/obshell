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

package process

import (
	"fmt"
	"os"

	"github.com/oceanbase/obshell/agent/errors"
	log "github.com/sirupsen/logrus"
)

const (
	Red    = "\033[31m"
	Reset  = "\033[0m"
	Yellow = "\033[33m"
)

func ExitWithFailure(code int, msg string) {
	log.Errorf("exit with code %d: %s", code, msg)
	fmt.Fprintf(os.Stderr, "%s[FAILED]%s %s\n", Red, Reset, msg)
	os.Exit(code)
}

func ExitWithError(code int, err error) {
	if err == nil {
		return
	}
	var msg string
	if ocsAgentErr, ok := err.(errors.OcsAgentErrorInterface); ok {
		msg = fmt.Sprintf("%s[ERROR]%s %sCode%s: %s, %sMessage%s: %s", Red, Reset, Yellow, Reset, ocsAgentErr.ErrorCode().Code, Yellow, Reset, ocsAgentErr.Error())
	} else {
		msg = fmt.Sprintf("%s[FAILED]%s %s", Red, Reset, err.Error())
	}
	ExitWithMsg(code, msg)
}

func ExitWithMsg(code int, msg string) {
	log.Infof("exit with code %d: %s", code, msg)
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(code)
}

func Exit(code int) {
	log.Infof("exit with code %d", code)
	os.Exit(code)
}
