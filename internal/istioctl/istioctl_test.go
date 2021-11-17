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
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/tetratelabs/getmesh/internal/test"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/internal/getmesh"
	"github.com/tetratelabs/getmesh/internal/manifest"
	"github.com/tetratelabs/getmesh/internal/util/logger"
)

func TestGetFetchedVersions(t *testing.T) {
	dir := test.TempDir(t, "", "")

	exp := map[string]struct{}{}
	for _, v := range []string{
		"20.1.1", "1.2.4", "1.7.3",
	} {
		d := &manifest.IstioDistribution{
			Version:       v,
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}
		ctlPath := GetIstioctlPath(dir, d)
		exp[d.String()] = struct{}{}

		suffix := strings.TrimSuffix(ctlPath, "/istioctl")
		require.NoError(t, os.MkdirAll(suffix, 0755))
		f, err := os.Create(ctlPath)
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}

	actual, err := GetFetchedVersions(dir)
	require.NoError(t, err)
	for _, a := range actual {
		delete(exp, a.String())
	}

	require.Empty(t, exp)
}

func TestPrintFetchedVersions(t *testing.T) {
	getmesh.GlobalConfigMux.Lock()
	defer getmesh.GlobalConfigMux.Unlock()

	t.Run("ok", func(t *testing.T) {
		dir := test.TempDir(t, "", "")

		d := &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		require.NoError(t, getmesh.SetIstioVersion(dir, d))
		for _, v := range []string{
			"20.1.1", "1.2.4", "1.7.3",
		} {
			ctlPath := GetIstioctlPath(dir, &manifest.IstioDistribution{
				Version:       v,
				Flavor:        manifest.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			})
			suffix := strings.TrimSuffix(ctlPath, "/istioctl")
			require.NoError(t, os.MkdirAll(suffix, 0755))
			f, err := os.Create(ctlPath)
			require.NoError(t, err)
			require.NoError(t, f.Close())
		}

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, PrintFetchedVersions(dir))
		})
		exp := `1.2.4-tetrate-v0
1.7.3-tetrate-v0 (Active)
20.1.1-tetrate-v0
`
		require.Equal(t, exp, buf.String())
	})
}

func TestGetCurrentExecutable(t *testing.T) {
	t.Run("non exist", func(t *testing.T) {
		getmesh.GlobalConfigMux.Lock()
		defer getmesh.GlobalConfigMux.Unlock()
		d := &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		dir := test.TempDir(t, "", "")
		require.NoError(t, getmesh.SetIstioVersion(dir, d))
		_, err := GetCurrentExecutable(dir)
		require.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		getmesh.GlobalConfigMux.Lock()
		defer getmesh.GlobalConfigMux.Unlock()

		dir := test.TempDir(t, "", "")

		d := &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		require.NoError(t, getmesh.SetIstioVersion(dir, d))
		require.NoError(t,
			os.MkdirAll(strings.TrimSuffix(GetIstioctlPath(dir, d), "/istioctl"), 0755))

		f, err := os.Create(GetIstioctlPath(dir, d))
		require.NoError(t, err)
		defer f.Close()

		actual, err := GetCurrentExecutable(dir)
		require.NoError(t, err)
		require.Equal(t, "1.7.3-tetrate-v0", actual.String())
	})
}

func Test_getmeshctlPath(t *testing.T) {
	d := &manifest.IstioDistribution{
		Version:       "1.7.3",
		Flavor:        manifest.IstioDistributionFlavorTetrate,
		FlavorVersion: 0,
	}
	require.Equal(t, "tmpdir/istio/1.7.3-tetrate-v0/bin/istioctl",
		GetIstioctlPath("tmpdir", d))
}

func Test_removeAll(t *testing.T) {
	t.Run("non exist", func(t *testing.T) {
		require.Error(t, removeAll("non-exist", nil))
	})

	t.Run("ok", func(t *testing.T) {
		dir := test.TempDir(t, "", "")

		distros := []*manifest.IstioDistribution{
			{Version: "1.7.1", Flavor: "tetrate", FlavorVersion: 1},
			{Version: "1.7.2", Flavor: "tetrate", FlavorVersion: 1},
			{Version: "1.7.3", Flavor: "tetrate", FlavorVersion: 1},
		}

		// create dirs
		for _, d := range distros {
			ctlPath := GetIstioctlPath(dir, d)
			suffix := strings.TrimSuffix(ctlPath, "/istioctl")
			require.NoError(t, os.MkdirAll(suffix, 0755))
			f, err := os.Create(ctlPath)
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// should exist
			require.NoError(t, checkExist(dir, d))
		}

		current := distros[0]
		require.NoError(t, removeAll(dir, current))

		// should not exist
		for _, d := range distros[1:] {
			require.Error(t, checkExist(dir, d))
		}

		//should exist
		require.NoError(t, checkExist(dir, current))
	})
}

func TestRemove(t *testing.T) {
	dir := test.TempDir(t, "", "")

	t.Run("skip", func(t *testing.T) {
		d := &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "tetrate",
			FlavorVersion: 1,
		}

		err := Remove(dir, d, d)
		require.NoError(t, err)
	})

	t.Run("non exist", func(t *testing.T) {
		d := &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "non-exist",
			FlavorVersion: 1,
		}

		err := Remove(dir, d, nil)
		require.Error(t, err)
		require.Equal(t, "we skip removing 1.7.3-non-exist-v1 since it does not exist in your system", err.Error())
	})

	t.Run("specific", func(t *testing.T) {
		target := &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "exist",
			FlavorVersion: 1,
		}
		ctlPath := GetIstioctlPath(dir, target)
		suffix := strings.TrimSuffix(ctlPath, "/istioctl")
		require.NoError(t, os.MkdirAll(suffix, 0755))
		f, err := os.Create(ctlPath)
		require.NoError(t, err)
		defer f.Close()

		// should exist
		require.NoError(t, checkExist(dir, target))

		// remove
		require.NoError(t, Remove(dir, target, &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "current",
			FlavorVersion: 1,
		}))

		// should not exist
		require.Error(t, checkExist(dir, target))
	})
}

func Test_checkExists(t *testing.T) {

	t.Run("exist", func(t *testing.T) {
		dir := test.TempDir(t, "", "")
		d := &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		ctlPath := GetIstioctlPath(dir, d)
		suffix := strings.TrimSuffix(ctlPath, "/istioctl")
		require.NoError(t, os.MkdirAll(suffix, 0755))
		f, err := os.Create(ctlPath)
		require.NoError(t, err)
		require.NoError(t, f.Close())
		require.NoError(t, checkExist(dir, d))
	})

	t.Run("non exist", func(t *testing.T) {
		dir := test.TempDir(t, "", "")
		d := &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "non-exist",
			FlavorVersion: 0,
		}
		require.Error(t, checkExist(dir, d))
	})
}

func TestSwitch(t *testing.T) {
	getmesh.GlobalConfigMux.Lock()
	defer getmesh.GlobalConfigMux.Unlock()

	dir := test.TempDir(t, "", "")

	t.Run("exist", func(t *testing.T) {
		d := &manifest.IstioDistribution{
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		for _, v := range []string{
			"20.1.1", "1.7.3",
		} {
			d.Version = v
			ctlPath := GetIstioctlPath(dir, d)
			suffix := strings.TrimSuffix(ctlPath, "/istioctl")
			require.NoError(t, os.MkdirAll(suffix, 0755))
			f, err := os.Create(ctlPath)
			require.NoError(t, err)
			require.NoError(t, f.Close())
		}

		d.Version = "1.7.3"
		require.NoError(t, getmesh.SetIstioVersion(dir, d))
		d.Version = "20.1.1"
		require.NoError(t, Switch(dir, d))
		require.Equal(t, d, getmesh.GetActiveConfig().IstioDistribution)
	})

	t.Run("non-exist", func(t *testing.T) {
		require.Error(t, Switch(dir, &manifest.IstioDistribution{
			Version:       "0.1.1",
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}))
		require.Error(t, Switch(dir, &manifest.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "non-exist",
			FlavorVersion: 0,
		}))
	})
}

func TestExec(t *testing.T) {
	getmesh.GlobalConfigMux.Lock()
	defer getmesh.GlobalConfigMux.Unlock()

	dir := test.TempDir(t, "", "")

	d := &manifest.IstioDistribution{
		Version:       "0.0.1",
		Flavor:        manifest.IstioDistributionFlavorTetrate,
		FlavorVersion: 0,
	}

	ctlPath := GetIstioctlPath(dir, d)
	suffix := strings.TrimSuffix(ctlPath, "/istioctl")
	require.NoError(t, getmesh.SetIstioVersion(dir, d))
	require.NoError(t, os.MkdirAll(suffix, 0755))
	f, err := os.Create(ctlPath + ".go")
	require.NoError(t, err)
	_, err = f.Write([]byte(`package main ; import "fmt" ;func main () {fmt.Print("istioctl")}`))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	cmd := exec.Command("go", "build", "-o", ctlPath, ctlPath+".go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())

	buf := new(bytes.Buffer)
	require.NoError(t, ExecWithWriters(dir, []string{"."}, buf, nil))
	require.Equal(t, buf.String(), "istioctl")
}

func TestFetch(t *testing.T) {
	dir := test.TempDir(t, "", "")
	ms := &manifest.Manifest{
		IstioDistributions: []*manifest.IstioDistribution{
			{
				Version:       "1.10.3",
				Flavor:        manifest.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			{
				Version:       "1.10.3",
				Flavor:        manifest.IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.6",
				Flavor:        manifest.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.6",
				Flavor:        manifest.IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.5",
				Flavor:        manifest.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
		},
	}
	t.Run("not-supported", func(t *testing.T) {
		for _, c := range []*manifest.IstioDistribution{
			{Version: "1000.7.4", Flavor: manifest.IstioDistributionFlavorTetrate},
			{Version: "1.7.5", Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
			{Version: "1.7.5", Flavor: manifest.IstioDistributionFlavorTetrate, FlavorVersion: 1},
		} {
			err := Fetch(dir, c, ms)
			require.Error(t, err)
		}
	})

	t.Run("supported", func(t *testing.T) {
		for _, c := range []*manifest.IstioDistribution{
			{Version: "1.10.3", Flavor: manifest.IstioDistributionFlavorTetrate, FlavorVersion: 0},
			{Version: "1.10.3", Flavor: manifest.IstioDistributionFlavorTetrateFIPS, FlavorVersion: 0},
		} {
			require.Error(t, checkExist(dir, c))
			err := Fetch(dir, c, ms)
			require.NoError(t, err)
			require.NoError(t, checkExist(dir, c))
		}
	})

	t.Run("already exist", func(t *testing.T) {
		target := &manifest.IstioDistribution{
			Version:       "111111111111",
			Flavor:        "noooooo",
			FlavorVersion: 1000,
		}
		ctlPath := GetIstioctlPath(dir, target)
		suffix := strings.TrimSuffix(ctlPath, "/istioctl")
		require.NoError(t, os.MkdirAll(suffix, 0755))
		f, err := os.Create(ctlPath)
		require.NoError(t, err)
		defer f.Close()
		require.NoError(t, Fetch(dir, target, &manifest.Manifest{}))
	})

	t.Run("not found", func(t *testing.T) {
		mf := &manifest.Manifest{
			IstioDistributions: []*manifest.IstioDistribution{
				{Version: "1.7.3", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
				{Version: "1.7.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrate},
			},
		}

		for _, target := range []*manifest.IstioDistribution{
			{
				Version:       "1.7.4",
				Flavor:        manifest.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.3",
				Flavor:        manifest.IstioDistributionFlavorTetrate,
				FlavorVersion: 10,
			},
			{
				Version:       "1.7.3",
				Flavor:        manifest.IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 100,
			},
		} {
			require.Error(t, Fetch(dir, target, mf))
		}
	})
}

func TestFetchIstioctlURL(t *testing.T) {
	istioDistribution := &manifest.IstioDistribution{
		Version:       "1.7.6",
		Flavor:        manifest.IstioDistributionFlavorTetrate,
		FlavorVersion: 0,
	}

	tests := map[string]struct {
		istioDistribution *manifest.IstioDistribution
		goos              string
		goarch            string
		want              string
	}{
		"linux-amd64": {
			istioDistribution: istioDistribution,
			goos:              "linux",
			goarch:            "amd64",
			want:              "https://istio.tetratelabs.io/getmesh/files/istio-1.7.6-tetrate-v0-linux-amd64.tar.gz",
		},
		"linux-arm64": {
			istioDistribution: istioDistribution,
			goos:              "linux",
			goarch:            "arm64",
			want:              "https://istio.tetratelabs.io/getmesh/files/istio-1.7.6-tetrate-v0-linux-arm64.tar.gz",
		},
		"darwin-arm64": {
			istioDistribution: istioDistribution,
			goos:              "darwin",
			goarch:            "arm64",
			want:              "https://istio.tetratelabs.io/getmesh/files/istio-1.7.6-tetrate-v0-osx-arm64.tar.gz",
		},
		"darwin-amd64": {
			istioDistribution: istioDistribution,
			goos:              "darwin",
			goarch:            "amd64",
			want:              "https://istio.tetratelabs.io/getmesh/files/istio-1.7.6-tetrate-v0-osx.tar.gz", // No arch
		},
		"madeupoos-madeuparch": { //  Check that follows os-arch convention
			istioDistribution: istioDistribution,
			goos:              "madeupoos",
			goarch:            "madeuparch",
			want:              "https://istio.tetratelabs.io/getmesh/files/istio-1.7.6-tetrate-v0-madeupoos-madeuparch.tar.gz",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := fetchIstioctlURL(tc.istioDistribution, tc.goos, tc.goarch)
			require.Equal(t, tc.want, got)
		})
	}
}
