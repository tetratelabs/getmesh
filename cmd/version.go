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

	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/istioctl"
	"github.com/tetratelabs/getistio/src/util"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func newVersionCmd(homedir, getIstioVersion string) *cobra.Command {
	var remote bool
	sanitizer := regexp.MustCompile("(?m)[\r\n]+^.*client\\sversion.*$")

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the versions of GetIstio cli, running Istiod, Envoy, and the active istioctl",
		Long:  `Show the versions of GetIstio cli, running Istiod, Envoy, and the active istioctl`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if getistio.GetActiveConfig().IstioDistribution == nil {
				return errors.New("please fetch Istioctl by `getistio fetch` beforehand")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cur, err := istioctl.GetCurrentExecutable(homedir)
			if err != nil {
				return err
			}

			logger.Infof("getistio version: %s\nactive istioctl: %s\n", getIstioVersion, cur.ToString())
			k8sCLient, err := util.GetK8sClient()
			if err != nil {
				logger.Infof("no active Kubernetes clusters found\n")
			} else {
				v, err := k8sCLient.ServerVersion()
				if err != nil {
					logger.Infof("cannot retrieve Kubernetes cluster server information\n")
				} else {
					logger.Infof("active kubernetes cluster run in %s platform in version %s\n", v.Platform, v.GitVersion)
				}

			}

			if remote {
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

			latest, err := getistio.LatestVersion()
			if err != nil {
				return fmt.Errorf("failed to check GetIstio's latest version: %v", err)
			}

			if latest != getIstioVersion {
				logger.Infof(
					"\nThe latest GetIstio of version %s is available. Please run 'getistio update' to install %s\n",
					latest, latest)
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&remote, "remote", "", true, "Use --remote=false to suppress control plane and data plane check")
	return cmd
}
