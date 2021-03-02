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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getistio/api"
	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func TestGetFetchedVersions(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	exp := map[string]struct{}{}
	for _, v := range []string{
		"20.1.1", "1.2.4", "1.7.3",
	} {
		d := &api.IstioDistribution{
			Version:       v,
			Flavor:        api.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}
		ctlPath := GetIstioctlPath(dir, d)
		exp[d.ToString()] = struct{}{}

		suffix := strings.TrimSuffix(ctlPath, "/istioctl")
		require.NoError(t, os.MkdirAll(suffix, 0755))
		f, err := os.Create(ctlPath)
		require.NoError(t, err)
		_ = f.Close()
	}

	actual, err := GetFetchedVersions(dir)
	require.NoError(t, err)
	for _, a := range actual {
		delete(exp, a.ToString())
	}

	require.Empty(t, exp)
}

func TestPrintFetchedVersions(t *testing.T) {
	getistio.GlobalConfigMux.Lock()
	defer getistio.GlobalConfigMux.Unlock()

	t.Run("ok", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		d := &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        api.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		require.NoError(t, getistio.SetIstioVersion(dir, d))
		for _, v := range []string{
			"20.1.1", "1.2.4", "1.7.3",
		} {
			ctlPath := GetIstioctlPath(dir, &api.IstioDistribution{
				Version:       v,
				Flavor:        api.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			})
			suffix := strings.TrimSuffix(ctlPath, "/istioctl")
			require.NoError(t, os.MkdirAll(suffix, 0755))
			f, err := os.Create(ctlPath)
			require.NoError(t, err)
			_ = f.Close()
		}

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, PrintFetchedVersions(dir))
		})
		exp := `1.2.4-tetrate-v0
1.7.3-tetrate-v0 (Active)
20.1.1-tetrate-v0
`
		assert.Equal(t, exp, buf.String())
	})
}

func TestGetCurrentExecutable(t *testing.T) {
	t.Run("non exist", func(t *testing.T) {
		getistio.GlobalConfigMux.Lock()
		defer getistio.GlobalConfigMux.Unlock()
		d := &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        api.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(dir)
		require.NoError(t, getistio.SetIstioVersion(dir, d))
		_, err = GetCurrentExecutable(dir)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		getistio.GlobalConfigMux.Lock()
		defer getistio.GlobalConfigMux.Unlock()

		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		d := &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        api.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		require.NoError(t, getistio.SetIstioVersion(dir, d))
		require.NoError(t,
			os.MkdirAll(strings.TrimSuffix(GetIstioctlPath(dir, d), "/istioctl"), 0755))

		f, err := os.Create(GetIstioctlPath(dir, d))
		require.NoError(t, err)
		defer f.Close()

		actual, err := GetCurrentExecutable(dir)
		assert.NoError(t, err)
		assert.Equal(t, "1.7.3-tetrate-v0", actual.ToString())
	})
}

func Test_GetIstioctlPath(t *testing.T) {
	d := &api.IstioDistribution{
		Version:       "1.7.3",
		Flavor:        api.IstioDistributionFlavorTetrate,
		FlavorVersion: 0,
	}
	assert.Equal(t, "tmpdir/istio/1.7.3-tetrate-v0/bin/istioctl",
		GetIstioctlPath("tmpdir", d))
}

func Test_removeAll(t *testing.T) {
	t.Run("non exist", func(t *testing.T) {
		require.Error(t, removeAll("non-exist", nil))
	})

	t.Run("ok", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		distros := []*api.IstioDistribution{
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
			f.Close()

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
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	t.Run("skip", func(t *testing.T) {
		d := &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "tetrate",
			FlavorVersion: 1,
		}

		err := Remove(dir, d, d)
		require.NoError(t, err)
	})

	t.Run("non exist", func(t *testing.T) {
		d := &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "non-exist",
			FlavorVersion: 1,
		}

		err := Remove(dir, d, nil)
		require.Error(t, err)
		assert.Equal(t, "we skip removing 1.7.3-non-exist-v1 since it does not exist in your system", err.Error())
	})

	t.Run("specific", func(t *testing.T) {
		target := &api.IstioDistribution{
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
		require.NoError(t, Remove(dir, target, &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "current",
			FlavorVersion: 1,
		}))

		// should not exist
		require.Error(t, checkExist(dir, target))
	})
}

func Test_checkExists(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	t.Run("exist", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(dir)
		d := &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        api.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		ctlPath := GetIstioctlPath(dir, d)
		suffix := strings.TrimSuffix(ctlPath, "/istioctl")
		require.NoError(t, os.MkdirAll(suffix, 0755))
		f, err := os.Create(ctlPath)
		require.NoError(t, err)
		f.Close()
		assert.NoError(t, checkExist(dir, d))
	})

	t.Run("non exist", func(t *testing.T) {
		assert.Error(t, checkExist(dir, &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "non-exist",
			FlavorVersion: 0,
		}))
	})
}

func TestSwitch(t *testing.T) {
	getistio.GlobalConfigMux.Lock()
	defer getistio.GlobalConfigMux.Unlock()

	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	t.Run("exist", func(t *testing.T) {
		d := &api.IstioDistribution{
			Flavor:        api.IstioDistributionFlavorTetrate,
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
			f.Close()
		}

		d.Version = "1.7.3"
		require.NoError(t, getistio.SetIstioVersion(dir, d))
		d.Version = "20.1.1"
		assert.NoError(t, Switch(dir, d))
		assert.Equal(t, d, getistio.GetActiveConfig().IstioDistribution)
	})

	t.Run("non-exist", func(t *testing.T) {
		assert.Error(t, Switch(dir, &api.IstioDistribution{
			Version:       "0.1.1",
			Flavor:        api.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}))
		assert.Error(t, Switch(dir, &api.IstioDistribution{
			Version:       "1.7.3",
			Flavor:        "non-exist",
			FlavorVersion: 0,
		}))
	})
}

func TestExec(t *testing.T) {
	getistio.GlobalConfigMux.Lock()
	defer getistio.GlobalConfigMux.Unlock()

	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	d := &api.IstioDistribution{
		Version:       "0.0.1",
		Flavor:        api.IstioDistributionFlavorTetrate,
		FlavorVersion: 0,
	}

	ctlPath := GetIstioctlPath(dir, d)
	suffix := strings.TrimSuffix(ctlPath, "/istioctl")
	require.NoError(t, getistio.SetIstioVersion(dir, d))
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
	assert.Equal(t, buf.String(), "istioctl")
}

func TestFetch(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	ms := &api.Manifest{
		IstioDistributions: []*api.IstioDistribution{
			{
				Version:       "1.7.6",
				Flavor:        api.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.6",
				Flavor:        api.IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.5",
				Flavor:        api.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
		},
	}

	type tc struct {
		name, version, flavor string
		flavorVersion         int
	}

	t.Run("not-supported", func(t *testing.T) {
		for _, c := range []tc{
			{version: "1000.7.4", flavor: api.IstioDistributionFlavorTetrate},
			{version: "1.7.5", flavor: api.IstioDistributionFlavorTetrateFIPS},
			{version: "1.7.5", flavor: api.IstioDistributionFlavorTetrate, flavorVersion: 1},
		} {
			_, err = Fetch(dir, c.name, c.version, c.flavor, c.flavorVersion, ms)
			require.Error(t, err)
		}
	})

	t.Run("supported", func(t *testing.T) {
		for _, c := range []tc{
			{version: "1.7.5", flavor: api.IstioDistributionFlavorTetrate, flavorVersion: 0},
			{version: "1.7.6", flavor: api.IstioDistributionFlavorTetrate, flavorVersion: 0},
			{name: "1.7.6-tetrate-v0"},
		} {
			_, err = Fetch(dir, c.name, c.version, c.flavor, c.flavorVersion, ms)
			require.NoError(t, err)
		}
	})
}

func Test_processFetchParams(t *testing.T) {
	type tc struct {
		name, version, flavor string
		flavorVersion         int
		mf                    *api.Manifest
		exp                   *api.IstioDistribution
	}

	for i, c := range []tc{
		{
			// no args -> fall back to the latest tetrate flavor
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
				},
			},
			exp: &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			// all given
			version: "1.7.3", flavorVersion: 100, flavor: api.IstioDistributionFlavorTetrate,
			exp: &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			// given name
			name: "1.7.3-tetrate-v100",
			exp:  &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			// flavor not given
			version: "1.7.3", flavorVersion: 100,
			exp: &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			//  flavorVersion not given -> fall back to the latest flavor version
			version: "1.7.3", flavor: api.IstioDistributionFlavorTetrateFIPS, flavorVersion: -1,
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 50, Flavor: api.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.7.3", FlavorVersion: 10000000, Flavor: api.IstioDistributionFlavorTetrate},
				},
			},
			exp: &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 50, Flavor: api.IstioDistributionFlavorTetrateFIPS},
		},
		{
			//  version not given -> choose the latest version given flavor in manifest
			flavor: api.IstioDistributionFlavorIstio, flavorVersion: 0,
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.8.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorIstio},
				},
			},
			exp: &api.IstioDistribution{Version: "1.8.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorIstio},
		},
		{
			//  version and flavor version  not given -> choose the latest version given flavor in manifest
			flavor: api.IstioDistributionFlavorIstio, flavorVersion: -1,
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.8.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorIstio},
				},
			},
			exp: &api.IstioDistribution{Version: "1.8.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorIstio},
		},
		{
			//  flavorVersion not given -> not found error
			version: "1.7.3", flavor: api.IstioDistributionFlavorTetrateFIPS, flavorVersion: -1,
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 50, Flavor: api.IstioDistributionFlavorTetrate},
				},
			},
		},
		{
			// flavor, flavorVersion not given -> fall back to the latest tetrate flavor
			version: "1.7.3", flavorVersion: -1,
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
				},
			},
			exp: &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			// patch version is not given in 'version', so should fallback to the latest patch version in the minor version
			version: "1.7", flavorVersion: -1, flavor: api.IstioDistributionFlavorTetrateFIPS,
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrate},
					{Version: "1.7.1", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.8.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
				},
			},
			exp: &api.IstioDistribution{Version: "1.7.1", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrateFIPS},
		},
		{
			// patch version is not given in 'version', so should fallback to the latest patch version in the minor version
			version: "1.7", flavorVersion: 0,
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.100", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrate},
					{Version: "1.7.20", FlavorVersion: 20, Flavor: api.IstioDistributionFlavorTetrate},
					{Version: "1.7.1", FlavorVersion: 1, Flavor: api.IstioDistributionFlavorTetrate},
					{Version: "1.7.1", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.8.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
				},
			},
			exp: &api.IstioDistribution{Version: "1.7.100", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
		},
	} {
		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			actual, err := processFetchParams(c.name, c.version, c.flavor, c.flavorVersion, c.mf)
			if c.exp == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.exp, actual)
			}
		})

	}
}

func Test_fetch(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	t.Run("already exist", func(t *testing.T) {
		target := &api.IstioDistribution{
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
		require.NoError(t, fetch(dir, target, &api.Manifest{}))
	})

	t.Run("not found", func(t *testing.T) {
		mf := &api.Manifest{
			IstioDistributions: []*api.IstioDistribution{
				{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrateFIPS},
				{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
			},
		}

		for _, target := range []*api.IstioDistribution{
			{
				Version:       "1.7.4",
				Flavor:        api.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.3",
				Flavor:        api.IstioDistributionFlavorTetrate,
				FlavorVersion: 10,
			},
			{
				Version:       "1.7.3",
				Flavor:        api.IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 100,
			},
		} {
			require.Error(t, fetch(dir, target, mf))
		}
	})
}

func Test_fetchIstioctl(t *testing.T) {
	// This test virtually validates the HEAD istio distributions' existence in the HEAD manifest.json

	f, err := ioutil.ReadFile("../../manifest.json")
	require.NoError(t, err)
	var m api.Manifest
	require.NoError(t, json.Unmarshal(f, &m))

	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	for _, d := range m.IstioDistributions {
		require.NoError(t, fetchIstioctl(dir, d))
		ctlPath := GetIstioctlPath(dir, d)
		_, err = os.Stat(ctlPath)
		require.NoError(t, err)
		t.Log(ctlPath)
	}
}
