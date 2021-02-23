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
	"github.com/tetratelabs/getistio/src/istioctl"
	"github.com/tetratelabs/getistio/src/util/logger"
)

type switchFlags struct {
	name, version, flavor string
	flavorVersion         int64
}

func newSwitchCmd(homedir string) *cobra.Command {
	var flag switchFlags

	cmd := &cobra.Command{
		Use:   "switch <istio version | flavor | flavor-version>|<istio version full name>",
		Short: "Switch the active istioctl to a specified version",
		Long:  `Switch the active istioctl to a specified version`,
		Example: `# switch the active istioctl version to version=1.7.4, flavor=tetrate and flavor-version=1
$ getistio switch --version 1.7.4 --flavor tetrate --flavor-version=1, 
$ getistio switch --name 1.7.4-tetrate-v1
switch also supports to change only one version|flavor|flavorVersion flag and follow the rest settings in active version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := switchParse(homedir, &flag)
			if err != nil {
				return err
			}
			return switchExec(homedir, d)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(&flag.name, "name", "", "", "Name of istioctl, e.g. 1.9.0-istio-v0")
	flags.StringVarP(&flag.version, "version", "", "", "Version of istioctl, e.g. 1.7.4")
	flags.StringVarP(&flag.flavor, "flavor", "", "", "Flavor of istioctl, e.g. \"tetrate\" or \"tetratefips\" or \"istio\"")
	flags.Int64VarP(&flag.flavorVersion, "flavor-version", "", -1, "Version of the flavor, e.g. 1")

	return cmd
}

// if set name, it should only parse name to distro
// if version, flavor and version are all set, just parse it to distro
// if there exists active distro, switch with only one or two command will use the active distro setting for unset command
// if there are no active distro exists, switch with only one or two command will use the default distro setting for unset command
// if all commands are not set, use active setting if there has otherwise use default version
// default version: latest version, default flavor: tetrate, default flavorversion: 0
func switchParse(homedir string, flags *switchFlags) (*api.IstioDistribution, error) {
	if len(flags.name) != 0 {
		d, err := api.IstioDistributionFromString(flags.name)
		if err != nil {
			return nil, fmt.Errorf("cannot parse given name to %s istio distribution", flags.name)
		}
		return d, nil
	}

	// assumption there exists at least one distribution, thus currDistro cannot be nil
	currDistro, _ := istioctl.GetCurrentExecutable(homedir)
	return switchHandleDistro(currDistro, flags)
}

func switchHandleDistro(curr *api.IstioDistribution, flags *switchFlags) (*api.IstioDistribution, error) {
	var version, flavor string
	var flavorVersion int64

	if curr != nil {
		version, flavor, flavorVersion = curr.Version, curr.Flavor, curr.FlavorVersion
	}

	if len(flags.version) != 0 {
		version = flags.version
	}
	if len(flags.flavor) != 0 {
		flavor = flags.flavor
	}
	if flags.flavorVersion != -1 {
		flavorVersion = flags.flavorVersion
	}

	if curr == nil && (len(flags.version) == 0 || len(flags.flavor) == 0 || flags.flavorVersion == -1){
		return nil, fmt.Errorf("cannot infer the target version, no active distribution exists")
	}
	
	return &api.IstioDistribution{
		Version:       version,
		Flavor:        flavor,
		FlavorVersion: flavorVersion,
	}, nil
}

func switchExec(homedir string, distribution *api.IstioDistribution) error {
	if err := istioctl.Switch(homedir, distribution); err != nil {
		return err
	}
	logger.Infof("istioctl switched to %s now\n", distribution.ToString())
	return nil
}
