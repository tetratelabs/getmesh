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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getmesh/internal/configvalidator"
	"github.com/tetratelabs/getmesh/internal/getmesh"
)

func newConfigValidateCmd(homedir string) *cobra.Command {
	var flagNS, flagOutputThreshold string

	cmd := &cobra.Command{
		Use:   "config-validate <file/directory>...",
		Short: "Validate the current Istio configurations in your cluster",
		Long: `Validate the current Istio configurations in your cluster just like 'istioctl analyze'. Inspect all namespaces by default.
If the <file/directory> is specified, we analyze the effect of applying these yaml files against the current cluster.`,
		Example: `# validating a local manifest against the current cluster
$ getmesh config-validate my-app.yaml another-app.yaml

# validating local manifests in a directory against the current cluster in a specific namespace
$ getmesh config-validate -n bookinfo my-manifest-dir/

NAME                        	RESOURCE TYPE 	ERROR CODE	SEVERITY	MESSAGE
httpbin                     	Service       	IST0108   	Warning 	[my-manifest-dir/service.yaml:1] Unknown annotation: networking.istio.io/non-exist

# for all namespaces
$ getmesh config-validate

NAMESPACE               NAME                    RESOURCE TYPE           ERROR CODE      SEVERITY        MESSAGE
default                 bookinfo-gateway        Gateway                 IST0101         Error           Referenced selector not found: "app=nonexisting"
bookinfo                default                 Peerauthentication      KIA0505         Error           Destination Rule disabling namespace-wide mTLS is missing
bookinfo                bookinfo-gateway        Gateway                 KIA0302         Warning         No matching workload found for gateway selector in this namespace

# for a specific namespace
$ getmesh config-validate -n bookinfo

NAME                    RESOURCE TYPE           ERROR CODE      SEVERITY        MESSAGE
bookinfo-gateway        Gateway                 IST0101         Error           Referenced selector not found: "app=nonexisting"
bookinfo-gateway        Gateway                 KIA0302         Warning         No matching workload found for gateway selector in this namespace

# for a specific namespace with Error as threshold for validation
$ getmesh config-validate -n bookinfo --output-threshold Error

NAME                    RESOURCE TYPE           ERROR CODE      SEVERITY        MESSAGE
bookinfo-gateway        Gateway                 IST0101         Error           Referenced selector not found: "app=nonexisting"

The following is the explanation of each column:
[NAMESPACE]
namespace of the resource

[NAME]
name of the resource

[RESOURCE TYPE]
resource type, i.e. kind, of the resource

[ERROR CODE]
The error code of the found issue which is prefixed by 'IST' or 'KIA'. Please refer to
- https://istio.io/latest/docs/reference/config/analysis/ for 'IST' error codes
- https://kiali.io/documentation/latest/validations/ for 'KIA' error codes

[SEVERITY] the severity of the found issue

[MESSAGE] the detailed message of the found issue`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if getmesh.GetActiveConfig().IstioDistribution == nil {
				return errors.New("please fetch Istioctl by `getmesh fetch` beforehand")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			validator, err := configvalidator.New(homedir, flagNS, flagOutputThreshold, args)
			if err != nil {
				return err
			}

			err = validator.Validate()
			if err == configvalidator.ErrConfigIssuesFound {
				os.Exit(1)
			}
			return err
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(&flagNS, "namespace", "n", "", "namespace for config validation")
	flags.StringVarP(&flagOutputThreshold, "output-threshold", "",
		configvalidator.SeverityLevelInfo.Name,
		fmt.Sprintf("severity level of analysis at which to display messages. Valid values: %v",
			configvalidator.SeverityNames))

	return cmd
}
