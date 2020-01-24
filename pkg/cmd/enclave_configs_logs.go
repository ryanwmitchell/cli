/*
Copyright © 2019 Doppler <support@doppler.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/DopplerHQ/cli/pkg/configuration"
	"github.com/DopplerHQ/cli/pkg/http"
	"github.com/DopplerHQ/cli/pkg/printer"
	"github.com/DopplerHQ/cli/pkg/utils"
	"github.com/spf13/cobra"
)

var configsLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "List config audit logs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		jsonFlag := utils.OutputJSON
		localConfig := configuration.LocalConfig(cmd)
		// number := utils.GetIntFlag(cmd, "number", 16)

		utils.RequireValue("token", localConfig.Token.Value)
		utils.RequireValue("project", localConfig.EnclaveProject.Value)
		utils.RequireValue("config", localConfig.EnclaveConfig.Value)

		logs, err := http.GetConfigLogs(localConfig.APIHost.Value, utils.GetBool(localConfig.VerifyTLS.Value, true), localConfig.Token.Value, localConfig.EnclaveProject.Value, localConfig.EnclaveConfig.Value)
		if !err.IsNil() {
			utils.HandleError(err.Unwrap(), err.Message)
		}

		printer.ConfigLogs(logs, len(logs), jsonFlag)
	},
}

var configsLogsGetCmd = &cobra.Command{
	Use:   "get [log_id]",
	Short: "Get config audit log",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jsonFlag := utils.OutputJSON
		localConfig := configuration.LocalConfig(cmd)

		utils.RequireValue("token", localConfig.Token.Value)
		utils.RequireValue("project", localConfig.EnclaveProject.Value)
		utils.RequireValue("config", localConfig.EnclaveConfig.Value)

		log := cmd.Flag("log").Value.String()
		if len(args) > 0 {
			log = args[0]
		}

		configLog, err := http.GetConfigLog(localConfig.APIHost.Value, utils.GetBool(localConfig.VerifyTLS.Value, true), localConfig.Token.Value, localConfig.EnclaveProject.Value, localConfig.EnclaveConfig.Value, log)
		if !err.IsNil() {
			utils.HandleError(err.Unwrap(), err.Message)
		}

		printer.ConfigLog(configLog, jsonFlag, true)
	},
}

var configsLogsRollbackCmd = &cobra.Command{
	Use:   "rollback [log_id]",
	Short: "Rollback a config change",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jsonFlag := utils.OutputJSON
		silent := utils.GetBoolFlag(cmd, "silent")
		localConfig := configuration.LocalConfig(cmd)

		utils.RequireValue("token", localConfig.Token.Value)
		utils.RequireValue("project", localConfig.EnclaveProject.Value)
		utils.RequireValue("config", localConfig.EnclaveConfig.Value)

		log := cmd.Flag("log").Value.String()
		if len(args) > 0 {
			log = args[0]
		}

		configLog, err := http.RollbackConfigLog(localConfig.APIHost.Value, utils.GetBool(localConfig.VerifyTLS.Value, true), localConfig.Token.Value, localConfig.EnclaveProject.Value, localConfig.EnclaveConfig.Value, log)
		if !err.IsNil() {
			utils.HandleError(err.Unwrap(), err.Message)
		}

		if !silent {
			printer.ConfigLog(configLog, jsonFlag, true)
		}
	},
}

func init() {
	configsLogsCmd.Flags().StringP("project", "p", "", "enclave project (e.g. backend)")
	configsLogsCmd.Flags().StringP("config", "c", "", "enclave config (e.g. dev)")
	// TODO: hide this flag until the api supports it
	// configsLogsCmd.Flags().IntP("number", "n", 5, "max number of logs to display")
	configsCmd.AddCommand(configsLogsCmd)

	configsLogsGetCmd.Flags().String("log", "", "audit log id")
	configsLogsGetCmd.Flags().StringP("project", "p", "", "enclave project (e.g. backend)")
	configsLogsGetCmd.Flags().StringP("config", "c", "", "enclave config (e.g. dev)")
	configsLogsCmd.AddCommand(configsLogsGetCmd)

	configsLogsRollbackCmd.Flags().String("log", "", "audit log id")
	configsLogsRollbackCmd.Flags().StringP("project", "p", "", "enclave project (e.g. backend)")
	configsLogsRollbackCmd.Flags().StringP("config", "c", "", "enclave config (e.g. dev)")
	configsLogsRollbackCmd.Flags().Bool("silent", false, "disable text output")
	configsLogsCmd.AddCommand(configsLogsRollbackCmd)

	enclaveCmd.AddCommand(configsCmd)
}
