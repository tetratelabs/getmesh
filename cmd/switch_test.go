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

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/internal/getmesh"
	"github.com/tetratelabs/getmesh/internal/istioctl"
	"github.com/tetratelabs/getmesh/internal/manifest"
)

func Test_switchParse(t *testing.T) {
	home, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	manifest.GlobalManifestURLMux.Lock()
	defer manifest.GlobalManifestURLMux.Unlock()

	m := &manifest.Manifest{
		IstioDistributions: []*manifest.IstioDistribution{
			{
				Version:       "1.7.6",
				Flavor:        manifest.IstioDistributionFlavorIstio,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.5",
				Flavor:        manifest.IstioDistributionFlavorIstio,
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

	t.Setenv("GETMESH_TEST_MANIFEST_PATH", f.Name())

	// set up active distro
	d := &manifest.IstioDistribution{
		Version:       "1.7.6",
		Flavor:        manifest.IstioDistributionFlavorTetrate,
		FlavorVersion: 0,
	}

	require.NoError(t, getmesh.SetIstioVersion(home, d))
	require.NoError(t,
		os.MkdirAll(strings.TrimSuffix(istioctl.GetIstioctlPath(home, d), "/istioctl"), 0755))

	f, err = os.Create(istioctl.GetIstioctlPath(home, d))
	require.NoError(t, err)
	defer f.Close()

	t.Run("ok", func(t *testing.T) {
		flag := &switchFlags{version: "1.7.6", flavor: "istio", flavorVersion: 0}
		distro, err := switchParse(home, flag)
		require.NoError(t, err)
		exp := &manifest.IstioDistribution{Version: "1.7.6", Flavor: "istio", FlavorVersion: 0}
		require.Equal(t, distro, exp)
	})
	t.Run("name", func(t *testing.T) {
		flag := &switchFlags{name: "1.8.3-istio-v0"}
		distro, err := switchParse(home, flag)
		require.NoError(t, err)
		exp := &manifest.IstioDistribution{Version: "1.8.3", Flavor: "istio", FlavorVersion: 0}
		require.Equal(t, distro, exp)
	})
	t.Run("group", func(t *testing.T) {
		flag := &switchFlags{version: "1.7", flavor: "istio", flavorVersion: 0}
		distro, err := switchParse(home, flag)
		require.NoError(t, err)
		exp := &manifest.IstioDistribution{Version: "1.7.6", Flavor: "istio", FlavorVersion: 0}
		require.Equal(t, distro, exp)
	})
}

func Test_switchHandleDistro(t *testing.T) {
	for _, c := range []struct {
		curr  *manifest.IstioDistribution
		flags *switchFlags
		exp   *manifest.IstioDistribution
	}{
		{
			curr:  &manifest.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			flags: &switchFlags{version: "1.8.3", flavor: "istio", flavorVersion: 1},
			exp:   &manifest.IstioDistribution{Version: "1.8.3", Flavor: "istio", FlavorVersion: 1},
		},
		{
			curr:  &manifest.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			flags: &switchFlags{version: "", flavor: "istio", flavorVersion: 1},
			exp:   &manifest.IstioDistribution{Version: "1.7.6", Flavor: "istio", FlavorVersion: 1},
		},
		{
			curr:  &manifest.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			flags: &switchFlags{version: "1.8.3", flavor: "", flavorVersion: -1},
			exp:   &manifest.IstioDistribution{Version: "1.8.3", Flavor: "tetratefips", FlavorVersion: 0},
		},
		{
			curr:  &manifest.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			flags: &switchFlags{version: "", flavor: "", flavorVersion: -1},
			exp:   &manifest.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
		},
	} {
		v, err := switchHandleDistro(c.curr, c.flags)
		require.NoError(t, err)
		require.Equal(t, c.exp, v)
	}
}
