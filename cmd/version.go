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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getmesh/src/getmesh"
	"github.com/tetratelabs/getmesh/src/istioctl"
	"github.com/tetratelabs/getmesh/src/util"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

func newVersionCmd(homedir, getmeshVersion string) *cobra.Command {
	var remote bool
	sanitizer := regexp.MustCompile("(?m)[\r\n]+^.*client\\sversion.*$")

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the versions of getmesh cli, running Istiod, Envoy, and the active istioctl",
		Long:  `Show the versions of getmesh cli, running Istiod, Envoy, and the active istioctl`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if getmesh.GetActiveConfig().IstioDistribution == nil {
				return errors.New("please fetch Istioctl by `getmesh fetch` beforehand")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cur, err := istioctl.GetCurrentExecutable(homedir)
			if err != nil {
				return err
			}

			logger.Infof("getmesh version: %s\nactive istioctl: %s\n", getmeshVersion, cur.ToString())
			k8sCLient, err := util.GetK8sClient()
			if err != nil {
				logger.Infof("no active Kubernetes clusters found\n")
			} else {
				_, err = k8sCLient.ServerVersion()
				if err != nil {
					logger.Infof("cannot retrieve Kubernetes cluster server information\n")
				}
			}

			if err == nil && remote {
				w := new(bytes.Buffer)
				as := []string{"version", "--remote=true"}
				if err := istioctl.ExecWithWriters(homedir, as, w, nil); err != nil {
					return fmt.Errorf("error executing istioctl: %v", err)
				}

				msg := w.String()
				if !strings.Contains(msg, istioctl.IstioVersionNoPodRunningMsg) {
					logger.Infof(sanitizer.ReplaceAllString(msg, ""))
				} else {
					logger.Infof(istioctl.IstioVersionNoPodRunningMsg + "\n")
				}
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&remote, "remote", "", true, "Use --remote=false to suppress control plane and data plane check")
	return cmd
}
