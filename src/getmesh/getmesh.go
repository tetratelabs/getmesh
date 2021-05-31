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

package getmesh

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/tetratelabs/getmesh/src/util/logger"
)

const (
	downloadShellURL           = "https://dl.getistio.io/public/raw/files/download.sh"
	downloadShellTestURLEnvKey = "GETMESH_TEST_DOWNLOAD_SHELL_URL"
)

func getDownloadShellURL() string {
	if url := os.Getenv(downloadShellTestURLEnvKey); len(url) != 0 {
		return url
	}
	return downloadShellURL
}

var latestVersionPrefixes = []string{
	"GETMESH_LATEST_VERSION=\"",
	// TODO: delete after the next release.
	"GETISTIO_LATEST_VERSION=\"",
}

func LatestVersion() (string, error) {
	url := getDownloadShellURL()
	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching donwload.sh: %v", err)
	}

	defer res.Body.Close()
	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading fetched donwload.sh: %v ", err)
	}

	var ret string
	r := bufio.NewScanner(bytes.NewReader(raw))
	for r.Scan() && ret == "" {
		for _, prefix := range latestVersionPrefixes {
			if line := r.Text(); strings.Contains(line, prefix) {
				ret = strings.TrimPrefix(line, prefix)
				ret = strings.TrimSuffix(ret, "\"")
				break
			}
		}
	}

	if len(ret) == 0 {
		return "", fmt.Errorf("latest version not found in donwload script. This is a bug in GetMesh")
	}
	return ret, nil
}

func Update(currentVersion string) error {
	l, err := LatestVersion()
	if err != nil {
		return fmt.Errorf("failed to get the latest version of getmesh: %w", err)
	}

	if l == currentVersion {
		logger.Infof("Your getmesh version is up-to-date: %s\n", currentVersion)
		return nil
	}

	url := getDownloadShellURL()
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch donwload.sh: %w", err)
	}
	defer res.Body.Close()

	cmd := exec.Command("bash", "-")
	cmd.Stdin = res.Body
	cmd.Stdout = logger.GetWriter()
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "FETCH_LATEST_ISTIOCTL=false")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	logger.Infof("\ngetmesh successfully updated from %s to %s!\n", currentVersion, l)
	return nil
}
