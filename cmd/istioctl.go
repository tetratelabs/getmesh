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
	"os"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getistio/api"
	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/istioctl"
	"github.com/tetratelabs/getistio/src/manifest"
	"github.com/tetratelabs/getistio/src/util"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func newIstioCmd(homedir string) *cobra.Command {
	return &cobra.Command{
		Use:   "istioctl <args...>",
		Short: "Execute istioctl with given arguments",
		Long:  `Execute istioctl with given arguments where the version of istioctl is set by "getsitio fetch or switch"`,
		Example: `# install Istio with the default profile
getistio istioctl install --set profile=default

# check versions of Istio data plane, control plane, and istioctl
getistio istioctl version`,
		PreRunE: func(_ *cobra.Command, args []string) error {
			cur := getistio.GetActiveConfig().IstioDistribution
			if cur == nil {
				return errors.New("please fetch Istioctl by `getistio fetch` beforehand")
			}
			newArgs, err := istioctlArgChecks(args, cur, getistio.GetActiveConfig().DefaultHub)
			if err != nil {
				return err
			}
			// precheck inspects a Kubernetes cluster for istio
			return istioK8scompatibilityCheck(homedir, newArgs)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := istioctl.Exec(homedir, args); err != nil {
				return fmt.Errorf("error executing istioctl: %v", err)
			}
			return nil
		},

		// verify on whether istiod and CRDs are installed correctly
		PostRunE: func(_ *cobra.Command, args []string) error {
			args = istioctlParseVerifyInstallArgs(args)
			if len(args) > 0 {
				if err := istioctl.Exec(homedir, args); err != nil {
					return fmt.Errorf("error executing istioctl: %v", err)
				}
			}
			return nil
		},
		// Here we parse flags around istioctl as arguments
		DisableFlagParsing: true,
	}
}

func istioctlArgChecks(args []string, currentDistro *api.IstioDistribution, defaultHub string) ([]string, error) {
	// Sanitize args.
	out := istioctlPreProcessArgs(args)

	// Walk thourgh args and search if it has 1) install command, 2) "--set hub=..." parameter.
	var (
		prev            string
		hasInstallCMD   bool
		hasHubParameter bool
	)
	for _, a := range out {
		if a == "install" {
			hasInstallCMD = true
			m, err := manifest.FetchManifest()
			if err != nil {
				return nil, err
			}
			ok, err := currentDistro.ExistInManifest(m)
			if err != nil {
				return nil, err
			} else if !ok {
				logger.Warnf("Your active istioctl of version %s is deprecated. "+
					"We recommend you use the supported distribution listed in \"getistio list\" command. \n", currentDistro.ToString())
				p := promptui.Prompt{
					Label:     "Proceed",
					IsConfirm: true,
				}
				if _, err := p.Run(); err != nil {
					// error returned when it's not confirmed
					return nil, err
				}
			}

			err = istioctlPatchVersionCheck(currentDistro, m)
			if err != nil {
				return nil, err
			}
		}

		// Search "--set hub=..." args.
		if prev == "--set" || prev == "-s" {
			if kv := strings.SplitN(a, "=", 2); len(kv) == 2 && kv[0] == "hub" {
				hasHubParameter = true
			}
		}
		prev = a
	}

	// Insert the default hub set by "getistio default-hub --set".
	if hasInstallCMD && !hasHubParameter {
		if defaultHub != "" {
			out = append(out, "--set", fmt.Sprintf("hub=%s", defaultHub))
		}
	}
	return out, nil
}

// check on whether the current version is the latest patch given current group version
func istioctlPatchVersionCheck(current *api.IstioDistribution, ms *api.Manifest) error {
	latestPatch, _, err := api.GetLatestDistribution(current, ms)
	if err != nil {
		logger.Warnf("unable to check latest trusted patch version, %v", err)
		// should allow user to proceed further
		return nil
	}

	// this case handled before because we already when they exists in manifest
	// here we handle it to avoid nil pointer comparision
	if latestPatch == nil {
		return nil
	}

	if !current.Equal(latestPatch) {
		logger.Warnf("your current patch version %s is not the latest version %s. "+
			"We recommend you fetch the latest version through \"getistio fetch\" command, "+
			"and switch to the latest version through \"getistio switch\" command \n", current.Version, latestPatch.Version)

		p := promptui.Prompt{
			Label:     "Proceed",
			IsConfirm: true,
		}
		if _, err := p.Run(); err != nil {
			// error returned when it's not confirmed
			return err
		}
	}
	return nil
}

func istioctlParsePreCheckArgs(args []string) []string {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return nil
		}
	}
	argument := []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation()}
	hasInstall := istioctlProcessChecksArgs(args, &argument, true)
	if !hasInstall {
		return nil
	}
	return argument
}

func istioK8scompatibilityCheck(homedir string, args []string) error {
	// if current istioctl does not support preCheck in either stable or experimental version
	// getistio will bypass preCheck
	if !istioctlHasPreCheckCommand(homedir) {
		return nil
	}

	args = istioctlParsePreCheckArgs(args)
	if len(args) == 0 {
		return nil
	}

	buf := new(bytes.Buffer)
	if err := istioctl.ExecWithWriters(homedir, args, os.Stdout, buf); err != nil {
		return fmt.Errorf("error executing istioctl: %v, %s", err, buf.String())
	}
	msg := buf.String()
	if msg != "" {
		// if given namespace already install istiod, user should choose install it or not by themselves
		logger.Infof(msg + "\n")
	}
	return nil
}

func istioctlHasPreCheckCommand(homedir string) bool {
	// precheck availability in stable istioctl
	buf := new(bytes.Buffer)
	if err := istioctl.ExecWithWriters(homedir, []string{"--help"}, buf, os.Stderr); err != nil {
		return false
	}
	help := buf.String()
	if strings.Contains(help, "precheck") {
		return true
	}

	// precheck availability in experimental istioctl
	buf = new(bytes.Buffer)
	if err := istioctl.ExecWithWriters(homedir, []string{"x", "--help"}, buf, os.Stderr); err != nil {
		return false
	}
	help = buf.String()
	return strings.Contains(help, "precheck")
}

func istioctlParseVerifyInstallArgs(args []string) []string {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return nil
		}
	}
	argument := []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation()}
	hasInstallCMD := istioctlProcessChecksArgs(args, &argument, false)
	if !hasInstallCMD {
		return nil
	}
	return argument
}

// preprocess corner cases in handcrafted install command parsing
func istioctlPreProcessArgs(args []string) []string {
	ret := []string{}
	// match string in format like --manifests=testfile -f=a --set=profile=demo -s=profile=demo
	// and supports to match with dir/ or values.value
	valid := regexp.MustCompile(`^((((\-\-|\-)([\w./]+))(=)([\w./]+))|(((\-\-set=)|(\-s=))([\w./]+=)([\w./]+)))$`)
	var prev string
	for _, arg := range args {
		if prev != "--set" && prev != "-s" && valid.MatchString(arg) {
			splited := strings.SplitN(arg, "=", 2)
			if len(splited) != 2 {
				continue
			}
			ret = append(ret, strings.TrimSpace(splited[0]), strings.TrimSpace(splited[1]))
		} else {
			ret = append(ret, strings.TrimSpace(arg))
		}
		prev = arg
	}
	return ret
}

// parse checks will parse install command line into precheck commands or verify install commands
// argument is the command pass to precheck and verify-install command and is passed by reference
func istioctlProcessChecksArgs(args []string, ourArg *[]string, precheck bool) bool {
	hasInstallCMD := false
	var prev string
	args = istioctlPreProcessArgs(args)
	for _, a := range args {
		if a == "install" {
			hasInstallCMD = true
		}
		switch prev {
		case "-f", "--filename", "--revision", "-r":
			*ourArg = append(*ourArg, prev, a)
		case "--set", "-s":
			kv := strings.SplitN(a, "=", 2)
			if len(kv) != 2 || strings.TrimSpace(kv[0]) != "values.global.istioNamespace" {
				prev = a
				continue
			}
			*ourArg = append(*ourArg, "--istioNamespace", kv[1])
		case "--manifests", "-d":
			if !precheck {
				*ourArg = append(*ourArg, prev, a)
			}
		default:
			// default do nothing
		}
		prev = a
	}
	return hasInstallCMD
}
