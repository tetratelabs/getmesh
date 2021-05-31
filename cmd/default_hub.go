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

	"github.com/tetratelabs/getmesh/src/getmesh"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

func newSetDefaultHubCmd(homedir string) *cobra.Command {
	var (
		removeFlag bool
		setFlag    string
		showFlag   bool
	)
	cmd := &cobra.Command{
		Use:   "default-hub",
		Short: `Set or Show the default hub passed to "getmesh istioctl install" via "--set hub=" e.g. docker.io/istio`,
		Long:  `Set or Show the default hub (root for Istio docker image paths) passed to "getmesh istioctl install" via "--set hub="  e.g. docker.io/istio`,
		Example: `# Set the default hub to docker.io/istio
$ getmesh default-hub --set docker.io/istio

# Show the configured default hub
$ getmesh default-hub --show

# Remove the configured default hub
$ getmesh default-hub --remove
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return defaultHubCheckFlags(removeFlag, setFlag, showFlag)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if setFlag != "" {
				return defaultHubHandleSet(homedir, setFlag)
			} else if removeFlag {
				return defaultHubHandleRemove(homedir)
			} else if showFlag {
				defaultHubHandleShow(getmesh.GetActiveConfig().DefaultHub)
			}
			defaultHubHandleShow(getmesh.GetActiveConfig().DefaultHub)
			return nil
		},
	}
	cmd.Flags().BoolVar(&removeFlag, "remove", false, "remove the configured default hub")
	cmd.Flags().StringVar(&setFlag, "set", "", "pass the location of hub, e.g. --set gcr.io/istio-testing")
	cmd.Flags().BoolVar(&showFlag, "show", false, "set to show the current default hub value")
	return cmd
}

var errDefaultHubArgCheck = errors.New("please provide exactly one of --remove, --set and --show flags for \"getmesh default-hub\" command")

func defaultHubCheckFlags(remove bool, setValue string, show bool) error {
	if setValue != "" {
		if remove || show {
			return errDefaultHubArgCheck
		}
		return nil
	}

	if remove {
		if show {
			return errDefaultHubArgCheck
		}
		return nil
	}

	if !show {
		return errDefaultHubArgCheck
	}
	return nil
}

func defaultHubHandleSet(homdir, setValue string) error {
	if err := getmesh.SetDefaultHub(homdir, setValue); err != nil {
		return err
	}
	logger.Infof("The default hub is now set to %s\n", setValue)
	return nil
}

func defaultHubHandleShow(current string) {
	if current == "" {
		logger.Infof("The default hub is not set yet. Istioctl's default value is used for \"getmesh istioctl install\" command\n")
	} else {
		logger.Infof("The current default hub is set to %s\n", current)
	}
}

func defaultHubHandleRemove(homdir string) error {
	if err := getmesh.SetDefaultHub(homdir, ""); err != nil {
		return err
	}
	logger.Infof("The default hub is removed. Now Istioctl's default value is used for \"getmesh istioctl install\" command\n")
	return nil
}
