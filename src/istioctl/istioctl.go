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

package istioctl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/tetratelabs/getistio/api"
	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/util/logger"
)

// This is the message put into stdout of "istioctl version" command when there are no istiod running in istio-system,
// i.e. the case where istio is not installed.
// And this is used for handling that case in "getistio version" and "getistio check-upgrade" commands.
// https://github.com/istio/istio/blob/593b8777047af29ed1307a8a0a96aa6481fb2664/pkg/kube/client.go#L698
const IstioVersionNoPodRunningMsg = "no running Istio pods in \"istio-system\""

var (
	istioDirSuffix     = "istio"
	istioctlPathFormat = filepath.Join(istioDirSuffix, "%s/bin/istioctl")
)

func GetIstioctlPath(homeDir string, distribution *api.IstioDistribution) string {
	path := fmt.Sprintf(istioctlPathFormat, distribution.ToString())
	return filepath.Join(homeDir, path)
}

func GetFetchedVersions(homedir string) ([]*api.IstioDistribution, error) {
	istioDir := filepath.Join(homedir, istioDirSuffix)
	ditros, _ := ioutil.ReadDir(istioDir) // intentionally ignore the error

	ret := make([]*api.IstioDistribution, 0, len(ditros))
	for _, raw := range ditros {
		if !raw.IsDir() {
			continue
		}

		d, err := api.IstioDistributionFromString(raw.Name())
		if err != nil {
			continue
		}
		ret = append(ret, d)
	}

	return ret, nil
}

func PrintFetchedVersions(homeDir string) error {
	curr, _ := GetCurrentExecutable(homeDir)
	istioDir := filepath.Join(homeDir, istioDirSuffix)
	ditros, err := ioutil.ReadDir(istioDir)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %v", istioDir, err)
	}

	if len(ditros) == 0 {
		logger.Infof("No Istioctl installed yet")
		return nil
	}

	for _, dist := range ditros {
		if !dist.IsDir() {
			continue
		}

		name := dist.Name()
		if strings.Contains(name, curr.ToString()) {
			logger.Infof(name + " (Active)\n")
		} else {
			logger.Infof(name + "\n")
		}
	}
	return nil
}

func removeAll(homeDir string, current *api.IstioDistribution) error {
	istioDir := filepath.Join(homeDir, istioDirSuffix)
	ditros, err := ioutil.ReadDir(istioDir)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %v", istioDir, err)
	}

	for _, dist := range ditros {
		if !dist.IsDir() {
			continue
		}

		name := dist.Name()
		if name == current.ToString() {
			continue
		}

		if err := os.RemoveAll(filepath.Join(istioDir, name)); err != nil {
			return fmt.Errorf("failed to remove %s: %w", name, err)
		}
	}
	return nil
}

// entrypoint for prune cmd
func Remove(homeDir string, target, current *api.IstioDistribution) error {
	if target == nil {
		return removeAll(homeDir, current)
	} else if current != nil && target.Equal(current) {
		logger.Infof("we skip removing %s since it is the current active version\n",
			target.ToString())
		return nil
	}

	if err := checkExist(homeDir, target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("we skip removing %s since it does not exist in your system",
				target.ToString())
		}
		return fmt.Errorf("error checking existence of %s: %w",
			target.ToString(), err)
	}

	p := filepath.Join(homeDir, istioDirSuffix, target.ToString())
	if err := os.RemoveAll(p); err != nil {
		return fmt.Errorf("failed to remove %s: %w", target.ToString(), err)
	}
	return nil
}

func checkExist(homeDir string, distribution *api.IstioDistribution) error {
	// check if the istio version already fetched
	path := GetIstioctlPath(homeDir, distribution)
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("istioctl not fetched for %s. Please run `getistio fetch`: %w",
				distribution.ToString(), os.ErrNotExist)
		}
		return fmt.Errorf("error checking istioctl: %v", err)
	}
	return nil
}

func GetCurrentExecutable(homeDir string) (*api.IstioDistribution, error) {
	conf := getistio.GetActiveConfig()
	if err := checkExist(homeDir, conf.IstioDistribution); err != nil {
		return nil, fmt.Errorf("check exist failed: %w", err)
	}
	return conf.IstioDistribution, nil
}

func Switch(homeDir string, distribution *api.IstioDistribution) error {
	if err := checkExist(homeDir, distribution); err != nil {
		return err
	}
	return getistio.SetIstioVersion(homeDir, distribution)
}

// getistio istioctl
func Exec(homeDir string, args []string) error {
	return ExecWithWriters(homeDir, args, nil, nil)
}

func ExecWithWriters(homeDir string, args []string, stdout, stderr io.Writer) error {
	conf := getistio.GetActiveConfig()
	if err := checkExist(homeDir, conf.IstioDistribution); err != nil {
		return err
	}
	path := GetIstioctlPath(homeDir, conf.IstioDistribution)
	cmd := exec.Command(path, args...)

	if stdout != nil {
		cmd.Stdout = stdout
	} else {
		cmd.Stdout = os.Stdout
	}

	if stderr != nil {
		cmd.Stderr = stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func Fetch(homeDir string, name, version, flavor string, flavorVersion int, ms *api.Manifest) (*api.IstioDistribution, error) {
	d, err := processFetchParams(name, version, flavor, flavorVersion, ms)
	if err != nil {
		return nil, err
	}

	return d, fetch(homeDir, d, ms)
}

func processFetchParams(name, version, flavor string, flavorVersion int,
	ms *api.Manifest) (*api.IstioDistribution, error) {
	if len(name) != 0 {
		d, err := api.IstioDistributionFromString(name)
		if err != nil {
			return nil, fmt.Errorf("cannot parse given name to %s istio distribution", name)
		}
		return d, nil
	}
	if flavor != api.IstioDistributionFlavorTetrate && flavor != api.IstioDistributionFlavorTetrateFIPS && flavor != api.IstioDistributionFlavorIstio {
		flavor = api.IstioDistributionFlavorTetrate
		logger.Infof("fallback to the %s flavor since --flavor flag is not given or not supported\n", flavor)
	}
	if len(version) == 0 {
		for _, m := range ms.IstioDistributions {
			if m.Flavor == flavor {
				return m, nil
			}
		}
	}

	ret := &api.IstioDistribution{Version: version, Flavor: flavor, FlavorVersion: int64(flavorVersion)}

	if strings.Count(version, ".") == 1 {
		// In the case where patch version is not given,
		// we find the latest patch version
		var (
			latest *api.IstioDistribution
			prev   *semver.Version
		)

		v, err := semver.NewVersion(version)
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
		logger.Infof("fallback to %s which is the latest patch version in the given verion minor %s\n", ret.Version, version)
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

func fetch(homeDir string, targetDistribution *api.IstioDistribution, ms *api.Manifest) error {
	var found bool
	for _, m := range ms.IstioDistributions {
		found = m.Equal(targetDistribution)
		if found {
			targetDistribution.ReleaseNotes = m.ReleaseNotes
			break
		}
	}

	if err := checkExist(homeDir, targetDistribution); err == nil {
		logger.Infof("%s already fetched: download skipped\n", targetDistribution.ToString())
		return nil
	}

	if !found {
		return fmt.Errorf("manifest not found for istioctl %s."+
			" Please check the supported istio versions and flavors by `getistio list`",
			targetDistribution.ToString())
	}

	return fetchIstioctl(homeDir, targetDistribution)
}

func fetchIstioctl(homeDir string, targetDistribution *api.IstioDistribution) error {
	dir := filepath.Join(homeDir, istioDirSuffix, targetDistribution.ToString())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cmd := exec.Command("sh", "-")
	cmd.Env = append(os.Environ(), fmt.Sprintf(
		"ISTIO_VERSION=%s", targetDistribution.Version),
		fmt.Sprintf("DISTRIBUTION_IDENTIFIER=%s", targetDistribution.ToString()),
	)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = bytes.NewBuffer([]byte(downloadScript))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while dowloading istio: %v", err)
	}

	if conf := getistio.GetActiveConfig(); conf.IstioDistribution == nil {
		if err := getistio.SetIstioVersion(homeDir, targetDistribution); err != nil {
			return fmt.Errorf("error switching to %s", conf.IstioDistribution.ToString())
		}
	}
	return nil
}
