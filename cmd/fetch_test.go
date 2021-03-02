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

	"github.com/stretchr/testify/assert"

	"github.com/tetratelabs/getistio/api"
)

func Test_fetchParams(t *testing.T) {
	type tc struct {
		flag *fetchFlags
		mf   *api.Manifest
		exp  *api.IstioDistribution
	}

	for i, c := range []tc{
		{
			// no args -> fall back to the latest tetrate flavor
			flag: &fetchFlags{flavorVersion: -1},
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
				},
			},
			exp: &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 0, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			// all given
			flag: &fetchFlags{version: "1.7.3", flavorVersion: 100, flavor: api.IstioDistributionFlavorTetrate},
			exp:  &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			// given name
			flag: &fetchFlags{name: "1.7.3-tetrate-v100"},
			exp:  &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			// flavor not given
			flag: &fetchFlags{version: "1.7.3", flavorVersion: 100},
			exp:  &api.IstioDistribution{Version: "1.7.3", FlavorVersion: 100, Flavor: api.IstioDistributionFlavorTetrate},
		},
		{
			//  flavorVersion not given -> fall back to the latest flavor version
			flag: &fetchFlags{version: "1.7.3", flavor: api.IstioDistributionFlavorTetrateFIPS, flavorVersion: -1},
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
			flag: &fetchFlags{flavor: api.IstioDistributionFlavorIstio, flavorVersion: 0},
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
			flag: &fetchFlags{flavor: api.IstioDistributionFlavorIstio, flavorVersion: -1},
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
			flag: &fetchFlags{version: "1.7.3", flavor: api.IstioDistributionFlavorTetrateFIPS, flavorVersion: -1},
			mf: &api.Manifest{
				IstioDistributions: []*api.IstioDistribution{
					{Version: "1.7.3", FlavorVersion: 50, Flavor: api.IstioDistributionFlavorTetrate},
				},
			},
		},
		{
			// flavor, flavorVersion not given -> fall back to the latest tetrate flavor
			flag: &fetchFlags{version: "1.7.3", flavorVersion: -1},
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
			flag: &fetchFlags{version: "1.7", flavor: api.IstioDistributionFlavorTetrateFIPS, flavorVersion: -1},
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
			flag: &fetchFlags{version: "1.7", flavorVersion: 0},
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
			actual, err := fetchParams(c.flag, c.mf)
			if c.exp == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.exp, actual)
			}
		})

	}
}
