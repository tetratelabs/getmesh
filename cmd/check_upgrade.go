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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	istioversion "istio.io/pkg/version"

	"github.com/tetratelabs/getmesh/src/checkupgrade"
	"github.com/tetratelabs/getmesh/src/getmesh"
	"github.com/tetratelabs/getmesh/src/istioctl"
	"github.com/tetratelabs/getmesh/src/manifest"
	"github.com/tetratelabs/getmesh/src/manifestchecker"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

func newCheckCmd(homedir string) *cobra.Command {
	return &cobra.Command{
		Use:   "check-upgrade",
		Short: "Check if there are patches available in the current minor version",
		Long:  `Check if there are patches available in the current minor version, e.g. 1.7-tetrate: 1.7.4-tetrate-v1 -> 1.7.5-tetrate-v1`,
		Example: `# example output
$ getmesh check-upgrade
...
- Your data plane running in multiple minor versions: 1.7-tetrate, 1.8-tetrate
- Your control plane running in multiple minor versions: 1.6-tetrate, 1.8-tetrate
- The minor version 1.6-tetrate is not supported by Tetrate.io. We recommend you use the trusted minor versions in "getmesh list"
- There is the available patch for the minor version 1.7-tetrate. We recommend upgrading all 1.7-tetrate versions -> 1.7.4-tetrate-v1
- There is the available patch for the minor version 1.8-tetrate which includes **security upgrades**. We strongly recommend upgrading all 1.8-tetrate versions -> 1.8.1-tetrate-v1

In the above example, we call names in the form of x.y-${flavor} "minor version", where x.y is Istio's upstream minor and ${flavor} is the flavor of the distribution.
Please refer to 'getmesh fetch --help' or 'getmesh list --help' for more information.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if getmesh.GetActiveConfig().IstioDistribution == nil {
				return errors.New("please fetch Istioctl by `getmesh fetch` beforehand")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ms, err := manifest.FetchManifest()
			if err != nil {
				return fmt.Errorf(" failed to fetch manifests")
			}

			if err := manifestchecker.Check(ms); err != nil {
				return err
			}

			w := new(bytes.Buffer)
			if err := istioctl.ExecWithWriters(homedir, []string{"version", "-o", "json"}, w, nil); err != nil {
				return fmt.Errorf("error executing istioctl: %v", err)
			}

			msg := w.String()
			if strings.Contains(msg, istioctl.IstioVersionNoPodRunningMsg) {
				logger.Infof(istioctl.IstioVersionNoPodRunningMsg + "\n")
				return nil
			}

			var iv istioversion.Version
			if err := json.Unmarshal(w.Bytes(), &iv); err != nil {
				return fmt.Errorf("failed to parse istio version results: %v: %s", err, w.Bytes())
			}

			if err := checkupgrade.IstioVersion(iv, ms); err != nil && err != checkupgrade.ErrIssueFound {
				return fmt.Errorf("failed to check Istio version: %v", err)
			} else if err == checkupgrade.ErrIssueFound {
				os.Exit(1)
			}
			return nil
		},
	}
}
