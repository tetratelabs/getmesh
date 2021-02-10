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

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getistio/api"
	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/istioctl"
)

func newPruneCmd(homedir string) *cobra.Command {
	var (
		flagVersion       string
		flagFlavor        string
		flagFlavorVersion int
	)

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove specific istioctl installed, or all, except the active one",
		Long:  "Remove specific istioctl installed, or all, except the active one",
		Example: `# remove all the installed
$ getistio prune

# remove the specific distribution
$ getistio prune --version 1.7.4 --flavor tetrate --flavor-version 0
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := pruneCheckFlags(flagVersion, flagFlavor, flagFlavorVersion)
			if err != nil {
				return err
			}
			return istioctl.Remove(homedir, target, getistio.GetActiveConfig().IstioDistribution)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&flagVersion, "version", "", "", "Version of istioctl e.g. 1.7.4")
	flags.StringVarP(&flagFlavor, "flavor", "", "", "Flavor of istioctl, e.g. \"tetrate\" or \"tetratefips\"")
	flags.IntVarP(&flagFlavorVersion, "flavor-version", "", -1, "Version of the flavor, e.g. 1")
	return cmd
}

func pruneCheckFlags(flagVersion string, flagFlavor string, flagFlavorVersion int) (*api.IstioDistribution, error) {
	var target *api.IstioDistribution
	if flagFlavor != "" || flagFlavorVersion != -1 || flagVersion != "" {
		if flagFlavor == "" || flagFlavorVersion == -1 || flagVersion == "" {
			return nil, fmt.Errorf("all of \"--version\", \"--flavor\" and \"--flavor-version \" " +
				"flags must be given when removing a specific version")
		}
		target = &api.IstioDistribution{
			Version:       flagVersion,
			Flavor:        flagFlavor,
			FlavorVersion: int64(flagFlavorVersion),
		}
	}
	return target, nil
}
