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

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func newSetDefaultHubCmd(homedir string) *cobra.Command {
	var (
		setFlag  string
		showFlag bool
	)
	cmd := &cobra.Command{
		Use:   "default-hub",
		Short: `Set or Show the default hub passed to "getistio istioctl install" via "--set hub=" e.g. docker.io/istio`,
		Long:  `Set or Show the default hub (root for Istio docker image paths) passed to "getistio istioctl install" via "--set hub="  e.g. docker.io/istio`,
		Example: `# Set the default hub to docker.io/istio
$ getistio default-hub --set docker.io/istio

# Show the current default hub
$ getistio default-hub --show
`,
		PreRunE: func(cmd *cobra.Command, args []string) error { return defaultHubCheckFlags(setFlag, showFlag) },
		RunE: func(cmd *cobra.Command, args []string) error {
			if setFlag != "" {
				return defaultHubHandleSet(homedir, setFlag)
			}
			defaultHubHandleShow(getistio.GetActiveConfig().DefaultHub)
			return nil
		},
	}
	cmd.Flags().StringVar(&setFlag, "set", "", "pass the URL of hub, e.g. --set gcr.io/istio-testing")
	cmd.Flags().BoolVar(&showFlag, "show", false, "set to show the current default hub value")
	return cmd
}

func defaultHubCheckFlags(setValue string, show bool) error {
	if (setValue != "" && show) || (setValue == "" && !show) {
		return errors.New("please provide exactly one of --set or --show flags for \"getistio default-hub\" command")
	}
	return nil
}

func defaultHubHandleSet(homdir, setValue string) error {
	if err := getistio.SetDefaultHub(homdir, setValue); err != nil {
		return err
	}
	logger.Infof("The default hub is now set to %s\n", setValue)
	return nil
}

func defaultHubHandleShow(current string) {
	if current == "" {
		logger.Infof("The default hub is not set yet. Istioctl's default value is used for \"getistio istioctl install\" command\n")
	} else {
		logger.Infof("The current default hub is set to %s\n", current)
	}
}
