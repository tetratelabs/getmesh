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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getistio/api"
)

func Test_switchHandleDistro(t *testing.T) {
	for _, c := range []struct {
		curr                            *api.IstioDistribution
		latest, flagVersion, flagFlavor string
		flagFlavorVersion               int
		exp                             *api.IstioDistribution
	}{
		{
			curr:              nil,
			latest:            "1.9.0",
			flagVersion:       "1.8.3",
			flagFlavor:        "istio",
			flagFlavorVersion: 1,
			exp:               &api.IstioDistribution{Version: "1.8.3", Flavor: "istio", FlavorVersion: 1},
		},
		{
			curr:              nil,
			latest:            "1.9.0",
			flagVersion:       "",
			flagFlavor:        "istio",
			flagFlavorVersion: 1,
			exp:               &api.IstioDistribution{Version: "1.9.0", Flavor: "istio", FlavorVersion: 1},
		},
		{
			curr:              nil,
			latest:            "1.9.0",
			flagVersion:       "1.8.3",
			flagFlavor:        "",
			flagFlavorVersion: -1,
			exp:               &api.IstioDistribution{Version: "1.8.3", Flavor: "tetrate", FlavorVersion: 0},
		},
		{
			curr:              nil,
			latest:            "1.9.0",
			flagVersion:       "",
			flagFlavor:        "",
			flagFlavorVersion: -1,
			exp:               &api.IstioDistribution{Version: "1.9.0", Flavor: "tetrate", FlavorVersion: 0},
		},

		{
			curr:              &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			latest:            "1.9.0",
			flagVersion:       "1.8.3",
			flagFlavor:        "istio",
			flagFlavorVersion: 1,
			exp:               &api.IstioDistribution{Version: "1.8.3", Flavor: "istio", FlavorVersion: 1},
		},
		{
			curr:              &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			latest:            "1.9.0",
			flagVersion:       "",
			flagFlavor:        "istio",
			flagFlavorVersion: 1,
			exp:               &api.IstioDistribution{Version: "1.7.6", Flavor: "istio", FlavorVersion: 1},
		},
		{
			curr:              &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			latest:            "1.9.0",
			flagVersion:       "1.8.3",
			flagFlavor:        "",
			flagFlavorVersion: -1,
			exp:               &api.IstioDistribution{Version: "1.8.3", Flavor: "tetratefips", FlavorVersion: 0},
		},
		{
			curr:              &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
			latest:            "1.9.0",
			flagVersion:       "",
			flagFlavor:        "",
			flagFlavorVersion: -1,
			exp:               &api.IstioDistribution{Version: "1.7.6", Flavor: "tetratefips", FlavorVersion: 0},
		},
	} {
		v, err := switchHandleDistro(c.curr, c.latest, c.flagVersion, c.flagFlavor, c.flagFlavorVersion)
		require.NoError(t, err)
		assert.Equal(t, c.exp, v)
	}
}
