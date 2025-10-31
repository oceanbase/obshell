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

package task

import (
	"fmt"
	"strings"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
)

func askConfirmForTaskOperation(id string, operator string) error {
	msg := fmt.Sprintf("Please confirm if you need to %s the task with ID %s ", strings.ToLower(operator), id)
	confirmed, err := stdio.Confirm(msg)
	if err != nil {
		return errors.Wrapf(err, "ask for task %s confirmation failed", strings.ToLower(operator))
	}
	if !confirmed {
		return errors.Occur(errors.ErrCliOperationCancelled)
	}
	return nil
}
