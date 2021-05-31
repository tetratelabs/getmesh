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
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/api"
	"github.com/tetratelabs/getmesh/src/getmesh"
	"github.com/tetratelabs/getmesh/src/manifest"
)

func Test_switchParse(t *testing.T) {
	home, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	manifest.GlobalManifestURLMux.Lock()
	defer manifest.GlobalManifestURLMux.Unlock()

	m := &api.Manifest{
		IstioDistributions: []*api.IstioDistribution{
			{
				Version:       "1.7.6",
				Flavor:        api.IstioDistributionFlavorIstio,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.5",
				Flavor:        api.IstioDistributionFlavorIstio,
				FlavorVersion: 0,
			},
		},
	}

	raw, err := json.Marshal(m)
	require.NoError(t, err)

	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer f.Close()

	_, err = f.Write(raw)
	require.NoError(t, err)

	require.NoError(t, os.Setenv("getmesh_TEST_MANIFEST_PATH", f.Name()))
	defer func() {
		require.NoError(t, os.Setenv("getmesh_TEST_MANIFEST_PATH", ""))
	}()

	// set up active distro
	d := &api.IstioDistribution{
		Version:       "1.7.6",
		Flavor:        api.IstioDistributionFlavorTetrate,
		FlavorVersion: 0,
	}

	require.NoError(t, getmesh.SetIstioVersion(home, d))
	require.NoError(t,
		os.MkdirAll(strings.TrimSuffix(istioctl.getmeshctlPath(home, d), "/istioctl"), 0755))

	f, err = os.Create(istioctl.getmeshctlPath(home, d))
	require.NoError(t, err)
	defer f.Close()

	t.Run("ok", func(t *testing.T) {
		flag := &switchFlags{version: "1.7.6", flavor: "istio", flavorVersion: 0}
		distro, err := switchParse(home, flag)
		require.NoError(t, err)
		exp := &api.IstioDistribution{Version: "1.7.6", Flavor: "istio", FlavorVersion: 0}
		assert.Equal(t, distro, exp)
	})
	t.Run("name", func(t *testing.T) {
		flag := &switchFlags{name: "1.8.3-istio-v0"}
		distro, err := switchParse(home, flag)
		require.NoError(t, err)
		exp := &api.IstioDistribution{Version: "1.8.3", Flavor: "istio", FlavorVersion: 0}
		assert.Equal(t, distro, exp)
	})
	t.Run("group", func(t *testing.T) {
		flag := &switchFlags{version: "1.7", flavor: "istio", flavorVersion: 0}
		distro, err := switchParse(home, flag)
		require.NoError(t, err)
		exp := &api.IstioDistribution{Version: "1.7.6", Flavor: "istio", FlavorVersion: 0}
		assert.Equal(t, distro, exp)
	})
}

func Test_switchHandleDistro(t *testing.T) {
	for _, c := range []struct {
		curr  *api.IstioDistribution
		flags *switchFlags
		exp   *api.IstioDistribution
	}{
		{
			curr:  &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			flags: &switchFlags{version: "1.8.3", flavor: "istio", flavorVersion: 1},
			exp:   &api.IstioDistribution{Version: "1.8.3", Flavor: "istio", FlavorVersion: 1},
		},
		{
			curr:  &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			flags: &switchFlags{version: "", flavor: "istio", flavorVersion: 1},
			exp:   &api.IstioDistribution{Version: "1.7.6", Flavor: "istio", FlavorVersion: 1},
		},
		{
			curr:  &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			flags: &switchFlags{version: "1.8.3", flavor: "", flavorVersion: -1},
			exp:   &api.IstioDistribution{Version: "1.8.3", Flavor: "tetratefips", FlavorVersion: 0},
		},
		{
			curr:  &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			flags: &switchFlags{version: "", flavor: "", flavorVersion: -1},
			exp:   &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
		},
	} {
		v, err := switchHandleDistro(c.curr, c.flags)
		require.NoError(t, err)
		assert.Equal(t, c.exp, v)
	}
}
