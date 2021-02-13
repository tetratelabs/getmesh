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
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getistio/api"
	"github.com/tetratelabs/getistio/src/istioctl"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func newSwitchCmd(homedir string) *cobra.Command {
	var (
		flagVersion       string
		flagFlavor        string
		flagFlavorVersion int
	)

	cmd := &cobra.Command{
		Use:   "switch <istio version>",
		Short: "Switch the active istioctl to a specified version",
		Long:  `Switch the active istioctl to a specified version`,
		Example: `# switch the active istioctl version to version=1.7.4, flavor=tetrate and flavor-version=1
$ getistio switch --version 1.7.4 --flavor tetrate --flavor-version=1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d := &api.IstioDistribution{
				Version:       flagVersion,
				Flavor:        flagFlavor,
				FlavorVersion: int64(flagFlavorVersion),
			}

			return switchExec(homedir, d)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(&flagVersion, "version", "", "", "Version of istioctl e.g. 1.7.4")
	flags.StringVarP(&flagFlavor, "flavor", "", "", "Flavor of istioctl, e.g. \"tetrate\" or \"tetratefips\"")
	flags.IntVarP(&flagFlavorVersion, "flavor-version", "", -1, "Version of the flavor, e.g. 1")

	_ = cmd.MarkFlagRequired("version")
	_ = cmd.MarkFlagRequired("flavor")
	_ = cmd.MarkFlagRequired("flavor-version")

	return cmd
}

func switchExec(homedir string, distribution *api.IstioDistribution) error {
	if err := istioctl.Switch(homedir, distribution); err != nil {
		return err
	}
	logger.Infof("istioctl switched to %s now\n", distribution.ToString())

	return nil
}
