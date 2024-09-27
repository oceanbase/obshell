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
	"errors"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type tenantModifyFlags struct {
	primaryZone                 string
	whitelist                   string
	interactivelyChangePassowrd bool
	oldPwd                      string
	newPwd                      string
	verbose                     bool
}

func newModifyCmd() *cobra.Command {
	opts := &tenantModifyFlags{}
	modifyCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_MODIFY,
		Short: "Modify tenant's properties, include primary_zone, root password etc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			if len(args) <= 0 {
				stdio.Error("tenant name is required")
				cmd.SilenceUsage = false
				return errors.New("tenant name is required")
			}
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			if err := tenantModify(cmd, args[0], opts); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: `  obshell tenant modify t1 --primary_zone RANDOM"`,
	})

	modifyCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	modifyCmd.Flags().SortFlags = false
	modifyCmd.VarsPs(&opts.primaryZone, []string{FLAG_PRIMARY_ZONE}, "", "Set the primary zone of the tenant", false)
	modifyCmd.VarsPs(&opts.whitelist, []string{FLAG_WHITELIST}, "", "Set the whitelist of the tenant", false)
	modifyCmd.VarsPs(&opts.interactivelyChangePassowrd, []string{FLAG_PASSWORD}, false, "Change password in interactive mode.", false)
	modifyCmd.VarsPs(&opts.oldPwd, []string{FLAG_OLD_PASSWORD}, "", "The old root password of tenant, Default to empty.", false)
	modifyCmd.VarsPs(&opts.newPwd, []string{FLAG_NEW_PASSWORD}, "", "The new root password of tenant", false)
	modifyCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return modifyCmd.Command
}

func tenantModify(cmd *cobra.Command, tenantName string, opts *tenantModifyFlags) (err error) {
	if opts.interactivelyChangePassowrd && (cmd.Flags().Changed(FLAG_OLD_PASSWORD) || cmd.Flags().Changed(FLAG_NEW_PASSWORD)) {
		cmd.SilenceUsage = false
		return errors.New("could not specify both --password and --old-password/--new-password")
	}
	if cmd.Flags().Changed(FLAG_OLD_PASSWORD) && !cmd.Flags().Changed(FLAG_NEW_PASSWORD) {
		cmd.SilenceUsage = false
		return errors.New("need specify --new-password when --old-password is specified")
	}
	if (cmd.Flags().Changed(FLAG_NEW_PASSWORD) || opts.interactivelyChangePassowrd || cmd.Flags().Changed(FLAG_OLD_PASSWORD)) && tenantName == constant.TENANT_SYS {
		return errors.New("could not change password of sys tenant")
	}
	if cmd.Flags().Changed(FLAG_NEW_PASSWORD) {
		stdio.StartLoadingf("set password of tenant %s", tenantName)
		if err := api.CallApiWithMethod(http.PUT, constant.URI_TENANT_API_PREFIX+"/"+tenantName+constant.URI_ROOTPASSWORD, param.ModifyTenantRootPasswordParam{
			OldPwd: opts.oldPwd,
			NewPwd: &opts.newPwd,
		}, nil); err != nil {
			return err
		}
		stdio.LoadSuccessf("set password of tenant %s", tenantName)
	}
	if opts.interactivelyChangePassowrd {
		old_password, err := stdio.InputPassword("Enter the old password(enter means empty): ")
		if err != nil {
			return err
		}
		stdio.Print("") // just for a new line
		new_password, err := stdio.InputPassword("Enter the new password(enter means empty): ")
		if err != nil {
			return err
		}
		stdio.Print("") // just for a new line
		new_password_confirm, err := stdio.InputPassword("Enter the new password again(enter means empty): ")
		if err != nil {
			return err
		}
		if new_password != new_password_confirm {
			return errors.New("The new password is not the same as the confirmation password")
		}

		stdio.StartLoadingf("set password of tenant %s", tenantName)
		if err := api.CallApiWithMethod(http.PUT, constant.URI_TENANT_API_PREFIX+"/"+tenantName+constant.URI_ROOTPASSWORD, param.ModifyTenantRootPasswordParam{
			OldPwd: old_password,
			NewPwd: &new_password,
		}, nil); err != nil {
			return err
		}
		stdio.LoadSuccessf("set password of tenant %s", tenantName)
	}
	if cmd.Flags().Changed(FLAG_PRIMARY_ZONE) {
		uri := constant.URI_TENANT_API_PREFIX + "/" + tenantName + constant.URI_PRIMARYZONE
		stdio.StartLoadingf("Call API %s", uri)
		dag := task.DagDetailDTO{}
		if err := api.CallApiWithMethod(http.PUT, uri, param.ModifyTenantPrimaryZoneParam{
			PrimaryZone: &opts.primaryZone,
		}, &dag); err != nil {
			return err
		}
		stdio.LoadSuccessf("Call API %s", uri)
		if err = api.NewDagHandler(&dag).PrintDagStage(); err != nil {
			return err
		}
	}
	if cmd.Flags().Changed(FLAG_WHITELIST) {
		stdio.StartLoading("set whitelist")
		if err := api.CallApiWithMethod(http.PUT, constant.URI_TENANT_API_PREFIX+"/"+tenantName+constant.URI_WHITELIST, param.ModifyTenantWhitelistParam{
			Whitelist: &opts.whitelist,
		}, nil); err != nil {
			return err
		}
		stdio.LoadSuccess("set whitelist")
	}
	return nil
}
