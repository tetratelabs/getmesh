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

	"github.com/tetratelabs/getistio/src/istioctl"
	"github.com/tetratelabs/getistio/src/manifest"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func newFetchCmd(homedir string) *cobra.Command {
	var (
		flagVersion       string
		flagFlavor        string
		flagFlavorVersion int
	)

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch istioctl of the specified version, flavor and flavor-version available in \"getistio list\" command",
		Long:  `Fetch istioctl of the specified version, flavor and flavor-version available in "getistio list" command`,
		Example: `# Fetch the latest "tetrate flavored" istioctl of version=1.8
$ getistio fetch --version 1.8

# Fetch the latest istioctl in version=1.7 and flavor=tetratefips
$ getistio fetch --version 1.7 --flavor tetratefips

# Fetch the latest istioctl of version=1.7, flavor=tetrate and flavor-version=0
$ getistio fetch --version 1.7 --flavor tetrate --flavor-version 0

# Fetch the istioctl of version=1.7.4 flavor=tetrate flavor-version=0
$ getistio fetch --version 1.7.4 --flavor tetrate --flavor-version 0

# Fetch the latest istioctl of version=1.7.4 and flavor=tetratefips
$ getistio fetch --version 1.7.4 --flavor tetratefips

# Fetch the latest "tetrate flavored" istioctl of version=1.7.4
$ getistio fetch --version 1.7.4



# Fetch the latest "tetrate flavored" istioctl
$ getistio fetch

As you can see the above examples:
- If --flavor-versions is not given, it defaults to the latest flavor version in the list
	If the value does not have patch version, "1.7" or "1.8" for example, then we fallback to the latest patch version in that minor version. 
- If --flavor is not given, it defaults to "tetrate" flavor.
- If --versions is not given, it defaults to the latest version of "tetrate" flavor.


For more information, please refer to "getistio list --help" command.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := manifest.FetchManifest()
			if err != nil {
				return fmt.Errorf("error fetching manifest: %v", err)
			}

			d, err := istioctl.Fetch(homedir, flagVersion, flagFlavor, flagFlavorVersion, ms)
			if err != nil {
				return err
			}

			var notes string
			for _, n := range d.ReleaseNotes {
				notes += "- " + n + "\n"
			}

			if len(notes) > 0 {
				logger.Infof("For more information about %s, please refer to the release notes: \n%s\n",
					d.ToString(), notes)
			}

			return switchExec(homedir, d)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(&flagVersion, "version", "", "", "Version of istioctl e.g. \"--version 1.7.4\"")
	flags.StringVarP(&flagFlavor, "flavor", "", "", "Flavor of istioctl, e.g. \"--flavor tetrate\" or --flavor tetratefips\"")
	flags.IntVarP(&flagFlavorVersion, "flavor-version", "", -1, "Version of the flavor, e.g. \"--version 1\"")
	return cmd
}
