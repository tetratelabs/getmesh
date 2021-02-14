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

	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/manifest"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available Istio distributions built by Tetrate",
		Long:  `List available Istio distributions built by Tetrate`,
		Example: `$ getistio list

ISTIO VERSION	FLAVOR 	FLAVOR VERSION	 K8S VERSIONS
   *1.8.2    	tetrate	      0       	1.16,1.17,1.18
    1.8.1    	tetrate	      0       	1.16,1.17,1.18
    1.7.6    	tetrate	      0       	1.16,1.17,1.18
    1.7.5    	tetrate	      0       	1.16,1.17,1.18
    1.7.4    	tetrate	      0       	1.16,1.17,1.18

'*' indicates the currently active istioctl version.

The following is the explanation of each column:

[ISTIO VERSION]
The official tagged version of Istio on which the distribution is built.

[FLAVOR]
The kind of the distribution. As of now, there are three flavors "tetrate",
"tetratefips" and "istio".

- "tetrate" flavor equals the official Istio except it is built by Tetrate.
- "tetratefips" flavor is FIPS-compliant, and can be used for installing FIPS-compliant control plain and data plain built by Tetrate.
- "istio" flavor is the upstream build.

[FLAVOR VERSION]
The flavor's version. A flavor version 0 maps to the distribution that is built on 
exactly the same source code of the corresponding official Istio version.

[K8S VERSIONS]
Supported k8s versions for the distribution
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := manifest.FetchManifest()
			if err != nil {
				return fmt.Errorf("error fetching manifest: %v", err)
			}

			if err := manifest.PrintManifest(ms, getistio.GetActiveConfig().IstioDistribution); err != nil {
				return fmt.Errorf("error executing istioctl: %v", err)
			}
			return nil
		},
	}
}
