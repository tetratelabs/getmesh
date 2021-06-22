// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getistio/src/util"
)

func Execute(version, homeDir string) {
	cmd := NewRoot(version, homeDir)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		handleUnknowns(cmd, err)
		os.Exit(1)
	}
}

// print help message on subcommand when encountered unknown commands or flags
func handleUnknowns(root *cobra.Command, err error) {
	if strings.Contains(err.Error(), "unknown flag") ||
		strings.Contains(err.Error(), "unknown command") {
		calledArgs := []*cobra.Command{}
		getCommandCallPath(root, &calledArgs)
		if len(calledArgs) == 0 {
			if err = root.Help(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			if err = calledArgs[len(calledArgs)-1].Help(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

// retrieve the call path given root command
func getCommandCallPath(cmd *cobra.Command, calledArgs *[]*cobra.Command) {
	if cmd == nil {
		return
	}
	*calledArgs = append(*calledArgs, cmd)
	if !cmd.HasSubCommands() {
		return
	}
	for _, sub := range cmd.Commands() {
		// if subcommand get called it will not be ""
		// https://pkg.go.dev/github.com/spf13/cobra#Command.CalledAs
		if sub.CalledAs() != "" {
			getCommandCallPath(sub, calledArgs)
		}
	}

}

func NewRoot(version, homeDir string) *cobra.Command {
	cmd := &cobra.Command{
		SilenceUsage:      true,
		SilenceErrors:     true,
		Use:               "getistio",
		DisableAutoGenTag: true,
		Short:             `GetIstio is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.`,
		Long:              `GetIstio is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.`,
	}

	cmd.AddCommand(newIstioCmd(homeDir))
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSwitchCmd(homeDir))
	cmd.AddCommand(newFetchCmd(homeDir))
	cmd.AddCommand(newVersionCmd(homeDir, version))
	cmd.AddCommand(newCheckCmd(homeDir))
	cmd.AddCommand(newShowCmd(homeDir))
	cmd.AddCommand(newConfigValidateCmd(homeDir))
	cmd.AddCommand(newGenCACmd())
	cmd.AddCommand(newUpgradeCmd(version))
	cmd.AddCommand(newPruneCmd(homeDir))
	cmd.AddCommand(newSetDefaultHubCmd(homeDir))

	cmd.PersistentFlags().StringVarP(&util.KubeConfig, "kubeconfig", "c", "", "Kubernetes configuration file")
	return cmd
}
