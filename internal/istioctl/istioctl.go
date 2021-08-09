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
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tetratelabs/getmesh/internal/getmesh"
	"github.com/tetratelabs/getmesh/internal/manifest"
	"github.com/tetratelabs/getmesh/internal/util/logger"
)

// This is the message put into stdout of "istioctl version" command when there are no istiod running in istio-system,
// i.e. the case where istio is not installed.
// And this is used for handling that case in "getmesh version" and "getmesh check-upgrade" commands.
// https://github.com/istio/istio/blob/593b8777047af29ed1307a8a0a96aa6481fb2664/pkg/kube/client.go#L698
const IstioVersionNoPodRunningMsg = "no running Istio pods in \"istio-system\""

var (
	istioDirSuffix                       = "istio"
	istioctlPathFormat                   = filepath.Join(istioDirSuffix, "%s/bin/istioctl")
	istioctlDownloadURLFormatWithArch    = "https://istio.tetratelabs.io/getmesh/files/istio-%s-%s-%s.tar.gz"
	istioctlDownloadURLFormatWithoutArch = "https://istio.tetratelabs.io/getmesh/files/istio-%s-%s.tar.gz"
)

func GetIstioctlPath(homeDir string, distribution *manifest.IstioDistribution) string {
	path := fmt.Sprintf(istioctlPathFormat, distribution.String())
	return filepath.Join(homeDir, path)
}

func GetFetchedVersions(homedir string) ([]*manifest.IstioDistribution, error) {
	istioDir := filepath.Join(homedir, istioDirSuffix)
	ditros, _ := ioutil.ReadDir(istioDir) // intentionally ignore the error

	ret := make([]*manifest.IstioDistribution, 0, len(ditros))
	for _, raw := range ditros {
		if !raw.IsDir() {
			continue
		}

		d, err := manifest.IstioDistributionFromString(raw.Name())
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
		if curr != nil && strings.Contains(name, curr.String()) {
			logger.Infof(name + " (Active)\n")
		} else {
			logger.Infof(name + "\n")
		}
	}
	return nil
}

func removeAll(homeDir string, current *manifest.IstioDistribution) error {
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
		if name == current.String() {
			continue
		}

		if err := os.RemoveAll(filepath.Join(istioDir, name)); err != nil {
			return fmt.Errorf("failed to remove %s: %w", name, err)
		}
	}
	return nil
}

// entrypoint for prune cmd
func Remove(homeDir string, target, current *manifest.IstioDistribution) error {
	if target == nil {
		return removeAll(homeDir, current)
	} else if current != nil && target.Equal(current) {
		logger.Infof("we skip removing %s since it is the current active version\n",
			target.String())
		return nil
	}

	if err := checkExist(homeDir, target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("we skip removing %s since it does not exist in your system",
				target.String())
		}
		return fmt.Errorf("error checking existence of %s: %w",
			target.String(), err)
	}

	p := filepath.Join(homeDir, istioDirSuffix, target.String())
	if err := os.RemoveAll(p); err != nil {
		return fmt.Errorf("failed to remove %s: %w", target.String(), err)
	}
	return nil
}

func checkExist(homeDir string, distribution *manifest.IstioDistribution) error {
	// check if the istio version already fetched
	path := GetIstioctlPath(homeDir, distribution)
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("istioctl not fetched for %s. Please run `getmesh fetch`: %w",
				distribution.String(), os.ErrNotExist)
		}
		return fmt.Errorf("error checking istioctl: %v", err)
	}
	return nil
}

func GetCurrentExecutable(homeDir string) (*manifest.IstioDistribution, error) {
	conf := getmesh.GetActiveConfig()
	if err := checkExist(homeDir, conf.IstioDistribution); err != nil {
		return nil, fmt.Errorf("check exist failed: %w", err)
	}
	return conf.IstioDistribution, nil
}

func Switch(homeDir string, distribution *manifest.IstioDistribution) error {
	if err := checkExist(homeDir, distribution); err != nil {
		return err
	}
	return getmesh.SetIstioVersion(homeDir, distribution)
}

// getmesh istioctl
func Exec(homeDir string, args []string) error {
	return ExecWithWriters(homeDir, args, nil, nil)
}

func ExecWithWriters(homeDir string, args []string, stdout, stderr io.Writer) error {
	conf := getmesh.GetActiveConfig()
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

func Fetch(homeDir string, target *manifest.IstioDistribution, ms *manifest.Manifest) error {
	var found bool
	for _, m := range ms.IstioDistributions {
		found = m.Equal(target)
		if found {
			target.ReleaseNotes = m.ReleaseNotes
			break
		}
	}

	if err := checkExist(homeDir, target); err == nil {
		logger.Infof("%s already fetched: download skipped\n", target.String())
		return nil
	}

	if !found {
		return fmt.Errorf("manifest not found for istioctl %s."+
			" Please check the supported istio versions and flavors by `getmesh list`",
			target.String())
	}

	return fetchIstioctl(homeDir, target)
}

func fetchIstioctl(homeDir string, targetDistribution *manifest.IstioDistribution) error {
	// Create dir
	dir := filepath.Join(homeDir, istioDirSuffix, targetDistribution.String(), "bin")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Construct URL from GOOS,GOARCH
	var url string
	if runtime.GOOS == "darwin" {
		url = fmt.Sprintf(istioctlDownloadURLFormatWithoutArch, targetDistribution.String(), runtime.GOOS)
	} else {
		url = fmt.Sprintf(istioctlDownloadURLFormatWithArch, targetDistribution.String(), runtime.GOOS, runtime.GOARCH)
	}

	// Download
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("404 not found for %s", url)
	}

	defer resp.Body.Close()

	// Read body
	gr, _ := gzip.NewReader(resp.Body)
	defer gr.Close()
	tr := tar.NewReader(gr)

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}

		if filepath.Base(h.Name) == "istioctl" {
			bin, err := io.ReadAll(tr)
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(dir, "istioctl"), bin, 0755); err != nil {
				return err
			}
		}
	}

	// Set active istioctl to the downloaded one
	if conf := getmesh.GetActiveConfig(); conf.IstioDistribution == nil {
		if err := getmesh.SetIstioVersion(homeDir, targetDistribution); err != nil {
			return fmt.Errorf("error switching to %s", conf.IstioDistribution.String())
		}
	}
	return nil
}
