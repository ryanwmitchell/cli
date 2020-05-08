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
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/DopplerHQ/cli/pkg/configuration"
	"github.com/DopplerHQ/cli/pkg/http"
	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/printer"
	"github.com/DopplerHQ/cli/pkg/utils"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the Doppler CLI for Enclave",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		silent := utils.GetBoolFlag(cmd, "silent")
		promptUser := !utils.GetBoolFlag(cmd, "no-prompt")
		scope := cmd.Flag("scope").Value.String()
		localConfig := configuration.LocalConfig(cmd)
		scopedConfig := configuration.Get(scope)

		utils.RequireValue("token", localConfig.Token.Value)


		currentProject := localConfig.EnclaveProject.Value
		selectedProject := ""

		switch localConfig.EnclaveProject.Source {
		case models.FlagSource.String():
			selectedProject = localConfig.EnclaveProject.Value
		case models.EnvironmentSource.String():
			utils.Log(valueFromEnvironmentNotice("ENCLAVE_PROJECT"))
			selectedProject = localConfig.EnclaveProject.Value
		default:
			projects, httpErr := http.GetProjects(localConfig.APIHost.Value, utils.GetBool(localConfig.VerifyTLS.Value, true), localConfig.Token.Value)
			if !httpErr.IsNil() {
				utils.HandleError(httpErr.Unwrap(), httpErr.Message)
			}
			if len(projects) == 0 {
				utils.HandleError(errors.New("you do not have access to any projects"))
			}

			selectedProject = selectProject(projects, scopedConfig.EnclaveProject.Value, promptUser)
			if selectedProject == "" {
				utils.HandleError(errors.New("Invalid project"))
			}
		}

		selectedConfiguredProject := selectedProject == currentProject
		selectedConfig := ""

		switch localConfig.EnclaveConfig.Source {
		case models.FlagSource.String():
			selectedConfig = localConfig.EnclaveConfig.Value
		case models.EnvironmentSource.String():
			utils.Log(valueFromEnvironmentNotice("ENCLAVE_CONFIG"))
			selectedConfig = localConfig.EnclaveConfig.Value
		default:
			configs, apiError := http.GetConfigs(localConfig.APIHost.Value, utils.GetBool(localConfig.VerifyTLS.Value, true), localConfig.Token.Value, selectedProject)
			if !apiError.IsNil() {
				utils.HandleError(apiError.Unwrap(), apiError.Message)
			}
			if len(configs) == 0 {
				utils.HandleError(errors.New("your project does not have any configs"))
			}

			selectedConfig = selectConfig(configs, selectedConfiguredProject, scopedConfig.EnclaveConfig.Value, promptUser)
			if selectedConfig == "" {
				utils.HandleError(errors.New("Invalid config"))
			}
		}

		configToSave := map[string]string{
			models.ConfigEnclaveProject.String(): selectedProject,
			models.ConfigEnclaveConfig.String():  selectedConfig,
		}
		configuration.Set(scope, configToSave)

		if !silent {
			// do not fetch the LocalConfig since we do not care about env variables or cmd flags
			conf := configuration.Get(scope)
			valuesToPrint := []string{models.ConfigEnclaveConfig.String(), models.ConfigEnclaveProject.String()}
			printer.ScopedConfigValues(conf, valuesToPrint, models.ScopedPairs(&conf), utils.OutputJSON, false, false)
		}
	},
}

func selectProject(projects []models.ProjectInfo, prevConfiguredProject string, promptUser bool) string {
	var options []string
	var defaultOption string
	for _, val := range projects {
		option := val.Name + " (" + val.ID + ")"
		options = append(options, option)

		if val.ID == prevConfiguredProject {
			defaultOption = option
		}
	}

	if !promptUser {
		utils.HandleError(errors.New("project must be specified via --project flag or ENCLAVE_PROJECT environment variable when using --no-prompt"))
	}

	prompt := &survey.Select{
		Message: "Select a project:",
		Options: options,
	}
	if defaultOption != "" {
		prompt.Default = defaultOption
	}

	selectedProject := ""
	err := survey.AskOne(prompt, &selectedProject)
	if err != nil {
		utils.HandleError(err)
	}

	for _, val := range projects {
		if strings.HasSuffix(selectedProject, "("+val.ID+")") {
			return val.ID
		}
	}

	return ""
}

func selectConfig(configs []models.ConfigInfo, selectedConfiguredProject bool, prevConfiguredConfig string, promptUser bool) string {
	var options []string
	var defaultOption string
	for _, val := range configs {
		option := val.Name
		options = append(options, option)

		// make previously selected config the default when re-using the previously selected project
		if selectedConfiguredProject && val.Name == prevConfiguredConfig {
			defaultOption = val.Name
		}
	}

	if !promptUser {
		utils.HandleError(errors.New("config must be specified via --config flag or ENCLAVE_CONFIG environment variable when using --no-prompt"))
	}

	prompt := &survey.Select{
		Message: "Select a config:",
		Options: options,
	}
	if defaultOption != "" {
		prompt.Default = defaultOption
	}

	selectedConfig := ""
	err := survey.AskOne(prompt, &selectedConfig)
	if err != nil {
		utils.HandleError(err)
	}

	return selectedConfig
}

func valueFromEnvironmentNotice(name string) string {
	return fmt.Sprintf("Using %s from the environment. To disable this, use --no-read-env.", name)
}

func init() {
	setupCmd.Flags().StringP("project", "p", "", "enclave project (e.g. backend)")
	setupCmd.Flags().StringP("config", "c", "", "enclave config (e.g. dev)")
	setupCmd.Flags().Bool("silent", false, "disable text output")
	setupCmd.Flags().Bool("no-prompt", false, "do not prompt for information. if the project or config is not specified, an error will be thrown.")
	enclaveCmd.AddCommand(setupCmd)
}
