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
	"strings"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getmesh/internal/istioctl"
	"github.com/tetratelabs/getmesh/internal/manifest"
	"github.com/tetratelabs/getmesh/internal/manifestchecker"
	"github.com/tetratelabs/getmesh/internal/util/logger"
)

type fetchFlags struct {
	name, version, flavor string
	flavorVersion         int64
}

func newFetchCmd(homedir string) *cobra.Command {
	var flag fetchFlags

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch istioctl of the specified version, flavor and flavor-version available in \"getmesh list\" command",
		Long:  `Fetch istioctl of the specified version, flavor and flavor-version available in "getmesh list" command`,
		Example: `# Fetch the latest "tetrate flavored" istioctl of version=1.8
$ getmesh fetch --version 1.8

# Fetch the latest istioctl in version=1.7 and flavor=tetratefips
$ getmesh fetch --version 1.7 --flavor tetratefips

# Fetch the latest istioctl of version=1.7, flavor=tetrate and flavor-version=0
$ getmesh fetch --version 1.7 --flavor tetrate --flavor-version 0

# Fetch the istioctl of version=1.7.4 flavor=tetrate flavor-version=0
$ getmesh fetch --version 1.7.4 --flavor tetrate --flavor-version 0

# Fetch the istioctl of version=1.7.4 flavor=tetrate flavor-version=0 using name
$ getmesh fetch --name 1.7.4-tetrate-v0

# Fetch the latest istioctl of version=1.7.4 and flavor=tetratefips
$ getmesh fetch --version 1.7.4 --flavor tetratefips

# Fetch the latest "tetrate flavored" istioctl of version=1.7.4
$ getmesh fetch --version 1.7.4

# Fetch the istioctl of version=1.8.3 flavor=istio flavor-version=0
$ getmesh fetch --version 1.8.3 --flavor istio



# Fetch the latest "tetrate flavored" istioctl
$ getmesh fetch

As you can see the above examples:
- If --flavor-versions is not given, it defaults to the latest flavor version in the list
	If the value does not have patch version, "1.7" or "1.8" for example, then we fallback to the latest patch version in that minor version. 
- If --flavor is not given, it defaults to "tetrate" flavor.
- If --versions is not given, it defaults to the latest version of "tetrate" flavor.


For more information, please refer to "getmesh list --help" command.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := manifest.FetchManifest()
			if err != nil {
				return fmt.Errorf("error fetching manifest: %v", err)
			}

			if err := manifestchecker.Check(ms); err != nil {
				return err
			}

			d, err := fetchParams(&flag, ms)
			if err != nil {
				return err
			}

			err = istioctl.Fetch(homedir, d, ms)
			if err != nil {
				return err
			}

			var notes string
			for _, n := range d.ReleaseNotes {
				notes += "- " + n + "\n"
			}

			if len(notes) > 0 {
				logger.Infof("For more information about %s, please refer to the release notes: \n%s\n",
					d.String(), notes)
			}

			return switchExec(homedir, d)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(&flag.name, "name", "", "", "Name of distribution, e.g. 1.9.0-istio-v0")
	flags.StringVarP(&flag.version, "version", "", "", "Version of istioctl e.g. \"--version 1.7.4\". When --name flag is set, this will not be used.")
	flags.StringVarP(&flag.flavor, "flavor", "", "",
		"Flavor of istioctl, e.g. \"--flavor tetrate\" or --flavor tetratefips\" or --flavor istio\". When --name flag is set, this will not be used.")
	flags.Int64VarP(&flag.flavorVersion, "flavor-version", "", -1,
		"Version of the flavor, e.g. \"--version 1\". When --name flag is set, this will not be used.")
	return cmd
}

func fetchParams(flags *fetchFlags,
	ms *manifest.Manifest) (*manifest.IstioDistribution, error) {
	if len(flags.name) != 0 {
		d, err := manifest.IstioDistributionFromString(flags.name)
		if err != nil {
			return nil, fmt.Errorf("cannot parse given name %s to istio distribution", flags.name)
		}
		return d, nil
	}
	if flags.flavor != manifest.IstioDistributionFlavorTetrate &&
		flags.flavor != manifest.IstioDistributionFlavorTetrateFIPS &&
		flags.flavor != manifest.IstioDistributionFlavorIstio {
		flags.flavor = manifest.IstioDistributionFlavorTetrate
		logger.Infof("fallback to the %s flavor since --flavor flag is not given or not supported\n", flags.flavor)
	}
	if len(flags.version) == 0 {
		for _, m := range ms.IstioDistributions {
			if m.Flavor == flags.flavor {
				return m, nil
			}
		}
	}

	ret := &manifest.IstioDistribution{Version: flags.version, Flavor: flags.flavor, FlavorVersion: flags.flavorVersion}

	if strings.Count(flags.version, ".") == 1 {
		// In the case where patch version is not given,
		// we find the latest patch version
		var (
			latest *manifest.IstioDistribution
			prev   *semver.Version
		)

		v, err := semver.NewVersion(flags.version)
		if err != nil {
			return nil, err
		}

		for _, d := range ms.IstioDistributions {
			cur, err := semver.NewVersion(d.Version)
			if err != nil {
				return nil, err
			}

			if d.Flavor == ret.Flavor && cur.Minor() == v.Minor() && (prev == nil || cur.GreaterThan(prev)) {
				prev = cur
				latest = d
			}
		}

		if latest == nil {
			return nil, fmt.Errorf("invalid version %s", ret.Version)
		}

		ret.Version = latest.Version
		logger.Infof("fallback to %s which is the latest patch version in the given verion minor %s\n",
			ret.Version, flags.version)
	}

	if ret.FlavorVersion < 0 {
		// search the latest flavor version in this flavor
		var found bool
		for _, m := range ms.IstioDistributions {
			if m.Version == ret.Version && m.Flavor == ret.Flavor {
				ret.FlavorVersion = m.FlavorVersion
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unsupported version=%s and flavor=%s", ret.Version, ret.Flavor)
		}
		logger.Infof("fallback to the flavor %d version which is the latest one in %s-%s\n",
			ret.FlavorVersion, ret.Version, ret.Flavor)
	}

	return ret, nil
}
