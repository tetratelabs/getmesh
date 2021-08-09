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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/internal/manifest"
)

func Test_fetchParams(t *testing.T) {
	type tc struct {
		flag *fetchFlags
		mf   *manifest.Manifest
		exp  *manifest.IstioDistribution
	}

	for i, c := range []tc{
		{
			// no args -> fall back to the latest tetrate flavor
			flag: &fetchFlags{flavorVersion: -1},
			mf: &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrate},
				},
			},
			exp: &manifest.IstioDistribution{Version: "1.7.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrate},
		},
		{
			// all given
			flag: &fetchFlags{version: "1.7.3", flavorVersion: 100, flavor: manifest.IstioDistributionFlavorTetrate},
			exp:  &manifest.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrate},
		},
		{
			// given name
			flag: &fetchFlags{name: "1.7.3-tetrate-v100"},
			exp:  &manifest.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrate},
		},
		{
			// flavor not given
			flag: &fetchFlags{version: "1.7.3", flavorVersion: 100},
			exp:  &manifest.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrate},
		},
		{
			//  flavorVersion not given -> fall back to the latest flavor version
			flag: &fetchFlags{version: "1.7.3", flavor: manifest.IstioDistributionFlavorTetrateFIPS, flavorVersion: -1},
			mf: &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 50, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.7.3", FlavorVersion: 10000000, Flavor: manifest.IstioDistributionFlavorTetrate},
				},
			},
			exp: &manifest.IstioDistribution{Version: "1.7.3", FlavorVersion: 50, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
		},
		{
			//  version not given -> choose the latest version given flavor in manifest
			flag: &fetchFlags{flavor: manifest.IstioDistributionFlavorIstio, flavorVersion: 0},
			mf: &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.8.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorIstio},
				},
			},
			exp: &manifest.IstioDistribution{Version: "1.8.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorIstio},
		},
		{
			//  version and flavor version  not given -> choose the latest version given flavor in manifest
			flag: &fetchFlags{flavor: manifest.IstioDistributionFlavorIstio, flavorVersion: -1},
			mf: &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.8.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorIstio},
				},
			},
			exp: &manifest.IstioDistribution{Version: "1.8.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorIstio},
		},
		{
			//  flavorVersion not given -> not found error
			flag: &fetchFlags{version: "1.7.3", flavor: manifest.IstioDistributionFlavorTetrateFIPS, flavorVersion: -1},
			mf: &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 50, Flavor: manifest.IstioDistributionFlavorTetrate},
				},
			},
		},
		{
			// flavor, flavorVersion not given -> fall back to the latest tetrate flavor
			flag: &fetchFlags{version: "1.7.3", flavorVersion: -1},
			mf: &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.7.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrate},
				},
			},
			exp: &manifest.IstioDistribution{Version: "1.7.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrate},
		},
		{
			// patch version is not given in 'version', so should fallback to the latest patch version in the minor version
			flag: &fetchFlags{version: "1.7", flavor: manifest.IstioDistributionFlavorTetrateFIPS, flavorVersion: -1},
			mf: &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrate},
					{Version: "1.7.1", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.8.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrate},
				},
			},
			exp: &manifest.IstioDistribution{Version: "1.7.1", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
		},
		{
			// patch version is not given in 'version', so should fallback to the latest patch version in the minor version
			flag: &fetchFlags{version: "1.7", flavorVersion: 0},
			mf: &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{
					{Version: "1.7.100", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrate},
					{Version: "1.7.20", FlavorVersion: 20, Flavor: manifest.IstioDistributionFlavorTetrate},
					{Version: "1.7.1", FlavorVersion: 1, Flavor: manifest.IstioDistributionFlavorTetrate},
					{Version: "1.7.1", FlavorVersion: 100, Flavor: manifest.IstioDistributionFlavorTetrateFIPS},
					{Version: "1.8.3", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrate},
				},
			},
			exp: &manifest.IstioDistribution{Version: "1.7.100", FlavorVersion: 0, Flavor: manifest.IstioDistributionFlavorTetrate},
		},
	} {
		t.Run(fmt.Sprintf("%d-th case", i), func(t *testing.T) {
			actual, err := fetchParams(c.flag, c.mf)
			if c.exp == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.exp, actual)
			}
		})

	}
}
