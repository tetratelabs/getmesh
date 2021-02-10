// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getistio/api"
	"github.com/tetratelabs/getistio/src/istioctl"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func Test_securityPatchCheckerImpl(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	locals := []*api.IstioDistribution{
		// non existent group
		{Version: "1.2.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0},
		// has a latest patch with security upgrade 1.7.6
		{Version: "1.7.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 10},
		{Version: "1.7.3", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 2},
		// 1.8.1 has a patch but not a security upgrade one
		{Version: "1.8.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 1},
		// has a security patch in 1.9.2 and a higher patch 1.9.10. So should be upgraded to 1.9.10
		{Version: "1.9.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0},
		// up-to-date
		{Version: "1.10.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0},
	}

	remotes := []*api.IstioDistribution{
		{Version: "1.7.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 12, IsSecurityPatch: false},
		{Version: "1.7.6", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 2, IsSecurityPatch: true},
		{Version: "1.8.2", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 1, IsSecurityPatch: false},
		{Version: "1.9.2", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0, IsSecurityPatch: true},
		{Version: "1.9.10", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0, IsSecurityPatch: false},
		{Version: "1.10.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0, IsSecurityPatch: true},
	}

	for _, d := range locals {
		ctlPath := istioctl.GetIstioctlPath(dir, d)
		suffix := strings.TrimSuffix(ctlPath, "/istioctl")
		require.NoError(t, os.MkdirAll(suffix, 0755))
		f, err := os.Create(ctlPath)
		require.NoError(t, err)
		_ = f.Close()
	}

	buf := logger.ExecuteWithLock(func() {
		require.NoError(t, securityPatchCheckerImpl(dir, &api.Manifest{
			IstioDistributions: remotes,
		}))
	})

	msg := buf.String()
	for _, exp := range []string{
		`[WARNING] The locally installed minor version 1.9-tetrate has a latest version 1.9.10-tetrate-v0 including security patches. We strongly recommend you to download 1.9.10-tetrate-v0 by "getistio fetch".`,
		`[WARNING] The locally installed minor version 1.2-tetrate is no longer supported by GetIstio. We recommend you use the higher minor versions in "getistio list" or remove with "getistio prune"`,
		`[WARNING] The locally installed minor version 1.7-tetrate has a latest version 1.7.6-tetrate-v2 including security patches. We strongly recommend you to download 1.7.6-tetrate-v2 by "getistio fetch".`,
	} {
		assert.Contains(t, msg, exp)
	}

	for _, nexp := range []string{"1.10", "1.8.2"} {
		assert.NotContains(t, msg, nexp)
	}

	t.Log(msg)
}

func Test_constructLatestVersionsMap(t *testing.T) {
	for _, c := range []struct {
		in  []*api.IstioDistribution
		exp map[string]*api.IstioDistribution
	}{
		{
			in: []*api.IstioDistribution{
				{Version: "1.8.10", FlavorVersion: 5, Flavor: "tetratefips"},
				{Version: "1.7.10", FlavorVersion: 1, Flavor: "tetrate"},
				{Version: "1.7.8", FlavorVersion: 1, Flavor: "tetrate"},
			},
			exp: map[string]*api.IstioDistribution{
				"1.8-tetratefips": {FlavorVersion: 5, Flavor: "tetratefips", Version: "1.8.10"},
				"1.7-tetrate":     {FlavorVersion: 1, Flavor: "tetrate", Version: "1.7.10"},
			},
		},
		{
			in: []*api.IstioDistribution{
				{Version: "1.8.10", FlavorVersion: 5, Flavor: "tetratefips"},
				{Version: "1.7.10", FlavorVersion: 1, Flavor: "tetrate"},
				{Version: "1.7.8", FlavorVersion: 3, Flavor: "tetratefips"},
			},
			exp: map[string]*api.IstioDistribution{
				"1.8-tetratefips": {FlavorVersion: 5, Flavor: "tetratefips", Version: "1.8.10"},
				"1.7-tetrate":     {FlavorVersion: 1, Flavor: "tetrate", Version: "1.7.10"},
				"1.7-tetratefips": {FlavorVersion: 3, Flavor: "tetratefips", Version: "1.7.8"},
			},
		},
	} {
		actual, err := constructLatestVersionsMap(c.in)
		require.NoError(t, err)
		assert.Equal(t, c.exp, actual)
	}
}

func Test_findSecurityPatchUpgrade(t *testing.T) {
	remotes := []*api.IstioDistribution{
		{Version: "1.7.6", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 2, IsSecurityPatch: true},
		{Version: "1.7.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 10, IsSecurityPatch: false},
		{Version: "1.8.2", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 1, IsSecurityPatch: false},
		{Version: "1.9.2", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0, IsSecurityPatch: true},
		{Version: "1.9.10", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0, IsSecurityPatch: false},
	}

	for _, c := range []struct {
		base                    *api.IstioDistribution
		exp                     *api.IstioDistribution
		expIncludeSecurityPatch bool
	}{
		{
			base:                    &api.IstioDistribution{Version: "1.2.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0},
			exp:                     nil,
			expIncludeSecurityPatch: false,
		},
		{
			base: &api.IstioDistribution{Version: "1.7.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 10},
			exp: &api.IstioDistribution{
				Version: "1.7.6", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 2, IsSecurityPatch: true,
			},
			expIncludeSecurityPatch: true,
		},
		{
			base: &api.IstioDistribution{Version: "1.9.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0},
			exp: &api.IstioDistribution{
				Version: "1.9.10", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0,
			},
			expIncludeSecurityPatch: true,
		},
	} {
		g, err := c.base.Group()
		require.NoError(t, err)
		actual, includeSP, err := findSecurityPatchUpgrade(c.base, g, remotes)
		require.NoError(t, err)
		assert.Equal(t, c.exp, actual)
		assert.Equal(t, c.expIncludeSecurityPatch, includeSP)
	}
}
