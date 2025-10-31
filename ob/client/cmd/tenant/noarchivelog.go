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
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	cmdlib "github.com/oceanbase/obshell/ob/client/lib/cmd"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
)

func newNoArchiveLogCmd() *cobra.Command {
	opts := &ArchiveLogFlags{
		status: constant.ARCHIVELOG_STATUS_STOP,
	}
	archiveLogCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_NO_ARCHIVE_LOG,
		Short:   "Close the archive log of the specified tenant.",
		PreRunE: cmdlib.ValidateArgTenantName,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)

			opts.tenantName = args[0]
			return tenantOperatorArchiveLog(opts)
		}),
		Example: `  obshell tenant noarchivelog t1`,
	})

	archiveLogCmd.Flags().SortFlags = false
	archiveLogCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}

	archiveLogCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	archiveLogCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip confirmation", false)

	return archiveLogCmd.Command
}
