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

package e2e

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/getmesh/internal/test"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		log.Fatal(err)
	}

	// Set up manifest
	if err := os.Setenv("GETMESH_TEST_MANIFEST_PATH", "site/manifest.json"); err != nil {
		log.Fatal(err)
	}

	// Setup the latest istioctl
	cmd := exec.Command("./getmesh", "fetch")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func getmeshFetchRequire(t *testing.T, version, flavor, flavorVersion string) {
	require.NoError(t, exec.Command("./getmesh", "fetch", "--flavor", flavor, "--version", version, "--flavor-version", flavorVersion).Run())
}

func getmeshListRequire(t *testing.T, version, flavor, flavorVersion string) {
	cmd := exec.Command("./getmesh", "show")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	require.Contains(t, buf.String(), fmt.Sprintf("%s-%s-v%s", version, flavor, flavorVersion))
}

func getmeshListRequireNot(t *testing.T, version, flavor, flavorVersion string) {
	cmd := exec.Command("./getmesh", "show")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	require.NotContains(t, buf.String(), fmt.Sprintf("%s-%s-v%s", version, flavor, flavorVersion))
}

func TestFetch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GETMESH_HOME", home)

	cmd := exec.Command("./getmesh", "fetch", "--version=1.12.4", "--flavor=tetrate", "--flavor-version=0")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run(), buf.String())
	require.Contains(t, buf.String(), `For more information about 1.12.4-tetrate-v0, please refer to the release notes: 
- https://istio.io/latest/news/releases/1.12.x/announcing-1.12.4/

istioctl switched to 1.12.4-tetrate-v0 now
`)

	// not listed version should be error
	cmd = exec.Command("./getmesh", "fetch", "--version=1.70000000000000000000.4")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.Error(t, cmd.Run())

	cmd = exec.Command("./getmesh", "fetch", "--version=1.70000000000000000000.4", "--flavor-version=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.Error(t, cmd.Run())

	// fetch without version
	cmd = exec.Command("./getmesh", "fetch", "--flavor=tetrate", "--flavor-version=0")
	buf = new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	require.Contains(t, buf.String(), `-tetrate-v0 now`)

	// fetch with single flavor flag
	cmd = exec.Command("./getmesh", "fetch", "--flavor=tetrate")
	buf = new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	require.Contains(t, buf.String(), `-tetrate-v0 now`)

	// fetch another version
	cmd = exec.Command("./getmesh", "fetch", "--version=1.13.2")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())

	// check the active istioctl has been changed to the last fetched one
	cmd = exec.Command("./getmesh", "show")
	buf = new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	require.Contains(t, buf.String(), `1.13.2-tetrate-v0 (Active)`)
}

func TestPrune(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GETMESH_HOME", home)

	t.Run("specific", func(t *testing.T) {
		var version = "1.13.2"
		var flavor = "tetrate"
		var flavorVersion = strconv.Itoa(0)

		// fetch the target.
		getmeshFetchRequire(t, version, flavor, flavorVersion)
		// fetch another target since the target should not be active.
		getmeshFetchRequire(t, "1.10.3", flavor, flavorVersion)

		// check existence
		getmeshListRequire(t, version, flavor, flavorVersion)

		// prune
		cmd := exec.Command("./getmesh", "prune", "--version", version,
			"--flavor", flavor, "--flavor-version", flavorVersion)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())

		// check non-existence
		getmeshListRequireNot(t, version, flavor, flavorVersion)
	})

	t.Run("all", func(t *testing.T) {
		distros := []struct{ version, flavor, flavorVersion string }{
			{
				version:       "1.12.4",
				flavor:        "tetrate",
				flavorVersion: strconv.Itoa(0),
			},
			{
				version:       "1.13.2",
				flavor:        "tetrate",
				flavorVersion: strconv.Itoa(0),
			},
			{
				version:       "1.10.3",
				flavor:        "tetrate",
				flavorVersion: strconv.Itoa(0),
			},
		}
		for _, d := range distros {
			getmeshFetchRequire(t, d.version, d.flavor, d.flavorVersion)
			getmeshListRequire(t, d.version, d.flavor, d.flavorVersion)
		}

		// prune all except the active one
		cmd := exec.Command("./getmesh", "prune")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())

		// check if non-active ones are removed.
		for _, d := range distros[:len(distros)-1] {
			getmeshListRequireNot(t, d.version, d.flavor, d.flavorVersion)
		}

		// check if the active one is not removed.
		active := distros[len(distros)-1]
		getmeshListRequire(t, active.version, active.flavor, active.flavorVersion)
	})
}

func TestShow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GETMESH_HOME", home)

	distros := []struct{ version, flavor, flavorVersion string }{
		{
			version:       "1.12.4",
			flavor:        "tetrate",
			flavorVersion: strconv.Itoa(0),
		},
		{
			version:       "1.13.2",
			flavor:        "tetrate",
			flavorVersion: strconv.Itoa(0),
		},
		{
			version:       "1.10.3",
			flavor:        "tetrate",
			flavorVersion: strconv.Itoa(0),
		},
	}
	for _, d := range distros {
		getmeshFetchRequire(t, d.version, d.flavor, d.flavorVersion)
		getmeshListRequire(t, d.version, d.flavor, d.flavorVersion)
	}

	cmd := exec.Command("./getmesh", "show")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	exp := `1.10.3-tetrate-v0 (Active)
1.12.4-tetrate-v0
1.13.2-tetrate-v0`
	require.Contains(t, buf.String(), exp)
}

func TestSwitch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GETMESH_HOME", home)

	distros := []struct{ version, flavor, flavorVersion string }{
		{
			version:       "1.13.2",
			flavor:        "tetrate",
			flavorVersion: strconv.Itoa(0),
		},
		{
			version:       "1.10.3",
			flavor:        "tetrate",
			flavorVersion: strconv.Itoa(0),
		},
		{
			version:       "1.12.4",
			flavor:        "tetrate",
			flavorVersion: strconv.Itoa(0),
		},
	}
	for _, d := range distros {
		getmeshFetchRequire(t, d.version, d.flavor, d.flavorVersion)
		getmeshListRequire(t, d.version, d.flavor, d.flavorVersion)
	}

	t.Run("full", func(t *testing.T) {
		for _, v := range []string{"1.10.3", "1.12.4"} {
			{
				cmd := exec.Command("./getmesh", "switch",
					"--version", v, "--flavor", "tetrate", "--flavor-version=0",
				)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				require.NoError(t, cmd.Run())
			}
			{
				cmd := exec.Command("./getmesh", "istioctl", "version")
				buf := new(bytes.Buffer)
				cmd.Stdout = buf
				cmd.Stderr = os.Stderr
				require.NoError(t, cmd.Run())
				require.Contains(t, buf.String(), v)
			}
		}
	})
	t.Run("name", func(t *testing.T) {
		cmd := exec.Command("./getmesh", "switch",
			"--version", "1.10.3", "--flavor", "tetrate", "--flavor-version=0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())

		cmd = exec.Command("./getmesh", "istioctl", "version")
		buf := new(bytes.Buffer)
		cmd.Stdout = buf
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())
		require.Contains(t, buf.String(), "1.10.3-tetrate-v0")

		cmd = exec.Command("./getmesh", "switch",
			"--name", "1.12.4-tetrate-v0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())

		cmd = exec.Command("./getmesh", "istioctl", "version")
		buf = new(bytes.Buffer)
		cmd.Stdout = buf
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())
		require.Contains(t, buf.String(), "1.12.4-tetrate-v0")
	})
	t.Run("active", func(t *testing.T) {
		cmd := exec.Command("./getmesh", "fetch",
			"--version=1.12.4", "--flavor=istio", "--flavor-version=0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())

		cmd = exec.Command("./getmesh", "istioctl", "version")
		buf := new(bytes.Buffer)
		cmd.Stdout = buf
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())
		require.Contains(t, buf.String(), "1.12.4")
		require.NotContains(t, buf.String(), "1.12.4-tetrate-v0")

		cmd = exec.Command("./getmesh", "switch",
			"--flavor=tetrate",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())

		cmd = exec.Command("./getmesh", "istioctl", "version")
		buf = new(bytes.Buffer)
		cmd.Stdout = buf
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())
		require.Contains(t, buf.String(), "1.12.4-tetrate-v0")
	})
}

// E2E that requires k8s.
func TestE2E_requirek8s(t *testing.T) {
	defer func() {
		// Clean up.
		cmd := exec.Command("kubectl", "delete", "-f", "./e2e/testdata/")
		_ = cmd.Run()
		cmd = exec.Command("./getmesh", "istioctl", "x", "--purge")
		_ = cmd.Run()
	}()

	t.Run("istioctl_install", istioctlInstall)
	t.Run("version", versionTest)
	t.Run("check-upgrade", checkUpgrade)
	t.Run("config-validate", configValidate)
}

func istioctlInstall(t *testing.T) {
	cmd := exec.Command("./getmesh", "istioctl",
		"install", "--set", "profile=default", "-y")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	actual := buf.String()

	// istioctl x precheck
	require.Contains(t, actual, "No issues found when checking the cluster. Istio is safe to install or upgrade")
}

func versionTest(t *testing.T) {
	t.Run("remote", func(t *testing.T) {
		for _, args := range [][]string{
			{"version", "--remote=true"},
			{"version"},
		} {
			cmd := exec.Command("./getmesh", args...)
			buf := new(bytes.Buffer)
			cmd.Stdout = buf
			cmd.Stderr = os.Stderr
			require.NoError(t, cmd.Run())
			actual := buf.String()
			require.Contains(t, actual, "getmesh version: dev")
			require.Contains(t, actual, "active istioctl")
			// latest version is available
			require.Contains(t, actual, "control plane version")
			require.Contains(t, actual, "data plane version")
			fmt.Println(actual)
		}

	})
	t.Run("local", func(t *testing.T) {
		cmd := exec.Command("./getmesh", "version", "--remote=false")
		buf := new(bytes.Buffer)
		cmd.Stdout = buf
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())
		actual := buf.String()
		require.Contains(t, actual, "active istioctl")
		// latest version is available
		require.NotContains(t, actual, "control plane version")
		require.NotContains(t, actual, "data plane version")
	})
	t.Run("unknown cluster", func(t *testing.T) {
		cmd := exec.Command("./getmesh", "version", "-c", "unknown.yaml")
		buf := new(bytes.Buffer)
		cmd.Stdout = buf
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())
		actual := buf.String()
		require.Contains(t, actual, "active istioctl")
		require.Contains(t, actual, "no active Kubernetes clusters found")
	})
}

func checkUpgrade(t *testing.T) {

	cmd := exec.Command("./getmesh", "check-upgrade")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run(), buf.String())
	actual := buf.String()
	require.Contains(t, actual, "1.13.3-tetrate-v0 is the latest version in 1.13-tetrate")

	// change image to 1.8.1-tetrate-v0
	image := "containers.istio.tetratelabs.com/pilot:1.8.1-tetrate-v0"
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"discovery","image":"%s"}]}}}}`,
		image)
	cmd = exec.Command("kubectl", "patch", "deployment",
		"-nistio-system", "istiod", "-p", patch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	require.Eventually(t, func() bool {
		cmd := exec.Command("./getmesh", "check-upgrade")
		buf := new(bytes.Buffer)
		cmd.Stdout = buf
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		actual := buf.String()
		return strings.Contains(actual,
			"1.13.3-tetrate-v0 is the latest version in 1.13-tetrate")
	}, time.Minute, 3*time.Second)
}

func configValidate(t *testing.T) {
	cmd := exec.Command("kubectl", "apply", "-f", "./e2e/testdata/config-validate.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	time.Sleep(time.Second * 6)

	t.Run("all namespaces", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command("./getmesh", "config-validate")
		bufOut := new(bytes.Buffer)
		cmd.Stdout = bufOut
		cmd.Stderr = os.Stderr
		require.Error(t, cmd.Run())
		exps := []string{
			`KIA0505`, `Destination Rule disabling namespace-wide mTLS is missing`,
			`KIA1102`, `VirtualService is pointing to a non-existent gateway`,
		}

		out := bufOut.String()
		for _, exp := range exps {
			require.Contains(t, out, exp, exp)
		}
		fmt.Println(out)
	})

	t.Run("all namespaces with threshold", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command("./getmesh", "config-validate", "--output-threshold", "Error")
		bufOut := new(bytes.Buffer)
		cmd.Stdout = bufOut
		cmd.Stderr = os.Stderr
		require.Error(t, cmd.Run())

		out := bufOut.String()
		for _, exp := range []string{"Info", "Warning"} {
			require.NotContains(t, out, exp, exp)
		}
		fmt.Println(out)
	})

	t.Run("invalid kubeconfig", func(t *testing.T) {
		t.Parallel()
		// make a new location for config
		// TODO: misconfigured kubeconfig, i,e: unauthorized kubeconfig file
		f := test.TempFile(t, "", "")
		cmd := exec.Command("./getmesh", "config-validate", "--kubeconfig", f.Name())
		bufErr := new(bytes.Buffer)
		cmd.Stderr = bufErr
		require.Error(t, cmd.Run())

		out := bufErr.String()
		exp := fmt.Sprintf("error building config from kubeconfig located in %s", f.Name())
		require.Contains(t, out, exp)
	})

	t.Run("single namespace", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command("./getmesh", "config-validate", "-n", "bookinfo")
		bufOut := new(bytes.Buffer)
		cmd.Stdout = bufOut
		cmd.Stderr = os.Stderr
		require.Error(t, cmd.Run())

		exps := []string{
			`KIA0505`, `Destination Rule disabling namespace-wide mTLS is missing`,
			`KIA1102`, `VirtualService is pointing to a non-existent gateway`,
		}
		out := bufOut.String()
		for _, exp := range exps {
			require.Contains(t, out, exp, exp)
		}
		fmt.Println(out)
	})

	t.Run("healthy", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command("./getmesh", "config-validate", "-n", "healthy")
		bufOut := new(bytes.Buffer)
		cmd.Stdout = bufOut
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Run())

		out := bufOut.String()
		exp := "Configuration issues not found."
		require.Contains(t, out, exp, exp)
		fmt.Println(out)
	})

	t.Run("local file", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command("./getmesh",
			"config-validate", "-n", "invalid",
			"e2e/testdata/config-validate-local.yaml",
		)
		bufOut := new(bytes.Buffer)
		cmd.Stdout = bufOut
		cmd.Stderr = os.Stderr
		require.Error(t, cmd.Run())

		exps := []string{
			`ratings-bogus-weight-default`,
			`KIA1104`, `[e2e/testdata/config-validate-local.yaml] The weight is assumed to be 100 because there is only one route destination`,
		}
		out := bufOut.String()
		for _, exp := range exps {
			require.Contains(t, out, exp, exp)
		}
		fmt.Println(out)
	})

	t.Run("local directory", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command("./getmesh",
			"config-validate", "-n", "invalid",
			"e2e/testdata/config-validate-local",
		)
		bufOut := new(bytes.Buffer)
		cmd.Stdout = bufOut
		cmd.Stderr = os.Stderr
		require.Error(t, cmd.Run())

		exps := []string{
			`KIA1101`,
		}
		out := bufOut.String()
		for _, exp := range exps {
			require.Contains(t, out, exp, exp)
		}
		fmt.Println(out)
	})
}
