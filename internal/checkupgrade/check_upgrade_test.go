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

package checkupgrade

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	istioversion "istio.io/pkg/version"

	"github.com/tetratelabs/getmesh/internal/manifest"
	"github.com/tetratelabs/getmesh/internal/util/logger"
)

func TestIstioVersion(t *testing.T) {
	in := istioversion.Version{
		MeshVersion: &istioversion.MeshInfo{istioversion.ServerInfo{Component: "pilot", Info: istioversion.BuildInfo{
			Version: "1.7.4-tetrate-v0",
		}}},
		DataPlaneVersion: &[]istioversion.ProxyInfo{{
			ID:           "istio-ingressgateway-55f67b4b7f-hdx47.istio-system",
			IstioVersion: "1.7.4-tetrate-v0",
		}},
		ClientVersion: &istioversion.BuildInfo{
			Version: "1.7.4-tetrate-v0",
		},
	}

	require.NoError(t, IstioVersion(in, &manifest.Manifest{
		IstioDistributions: []*manifest.IstioDistribution{
			{Version: "1.7.4", FlavorVersion: 0, Flavor: "tetrate"},
			{Version: "1.6.10", FlavorVersion: 0, Flavor: "tetrate"},
		},
	}))

	require.Equal(t, ErrIssueFound, IstioVersion(in, &manifest.Manifest{
		IstioDistributions: []*manifest.IstioDistribution{{Version: "1.7.4", Flavor: "tetrate", FlavorVersion: 100}},
	}))
	require.Equal(t, ErrIssueFound, IstioVersion(in, &manifest.Manifest{
		IstioDistributions: []*manifest.IstioDistribution{{Version: "1.7.5", Flavor: "tetrate", FlavorVersion: 0}},
	}))
}

func Test_printSummary(t *testing.T) {
	t.Run("only client", func(t *testing.T) {
		in := istioversion.Version{
			ClientVersion: &istioversion.BuildInfo{
				Version: "1.7.4-tetrate-v0",
			},
		}

		buf := logger.ExecuteWithLock(func() {
			printSummary(in)
		})

		actual := buf.String()
		require.Equal(t, fmt.Sprintf("active istioctl version: %s\n\n", in.ClientVersion.Version), actual)
	})

	t.Run("full", func(t *testing.T) {
		t.Run("single version", func(t *testing.T) {
			in := istioversion.Version{
				ClientVersion: &istioversion.BuildInfo{
					Version: "1.7.4-tetrate-v0",
				},
				MeshVersion: &istioversion.MeshInfo{
					istioversion.ServerInfo{Component: "pilot", Info: istioversion.BuildInfo{
						Version: "1.7.4-tetrate-v0",
					}},
				},
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.4-tetrate-v0"},
					{IstioVersion: "1.7.4-tetrate-v0"},
				},
			}
			buf := logger.ExecuteWithLock(func() {
				printSummary(in)
			})

			actual := buf.String()
			require.Equal(t, `active istioctl version: 1.7.4-tetrate-v0
data plane version: 1.7.4-tetrate-v0 (2 proxies)
control plane version: 1.7.4-tetrate-v0

`, actual)
		})

		t.Run("multiple versions", func(t *testing.T) {
			in := istioversion.Version{
				ClientVersion: &istioversion.BuildInfo{
					Version: "1.7.4-tetrate-v0",
				},
				MeshVersion: &istioversion.MeshInfo{
					istioversion.ServerInfo{Component: "pilot", Info: istioversion.BuildInfo{
						Version: "1.7.4-tetrate-v0",
					}},
					istioversion.ServerInfo{Component: "pilot", Info: istioversion.BuildInfo{
						Version: "1.8.1-tetrate-v0",
					}},
				},
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.4-tetrate-v0"},
					{IstioVersion: "1.7.4-tetrate-v0"},
					{IstioVersion: "1.8.4-tetrate-v0"},
				},
			}
			buf := logger.ExecuteWithLock(func() {
				printSummary(in)
			})

			actual := buf.String()
			require.Equal(t, `active istioctl version: 1.7.4-tetrate-v0
data plane version: 1.7.4-tetrate-v0 (2 proxies), 1.8.4-tetrate-v0 (1 proxies)
control plane version: 1.7.4-tetrate-v0, 1.8.1-tetrate-v0

`, actual)
		})
	})
}

func Test_printgetmeshCheck(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		for _, c := range []istioversion.Version{
			{
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.xxxxxx"},
				},
			},
			{
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.1"},
				},
				MeshVersion: &istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.xxxxxxxxxx"}},
				},
			},
			{
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.1-tetrate-v1"},
				},
				MeshVersion: &istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetrate-v1"}},
				},
			},
		} {
			err := printgetmeshCheck(c, &manifest.Manifest{
				IstioDistributions: []*manifest.IstioDistribution{{Version: "1.4.aaaa"}}})
			require.Error(t, err)
			t.Log(err)
		}
	})

	t.Run("only upstream versions", func(t *testing.T) {
		for i, c := range []istioversion.Version{
			{
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.1"},
					{IstioVersion: "1.6.1"},
				},
			},
			{
				MeshVersion: &istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.1"}},
					{Info: istioversion.BuildInfo{Version: "1.20.1"}},
				},
			},
			{
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.1"},
					{IstioVersion: "1.6.1"},
				},
				MeshVersion: &istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.1"}},
					{Info: istioversion.BuildInfo{Version: "1.20.1"}},
				},
			},
		} {
			t.Run(fmt.Sprintf("%d-th", i), func(t *testing.T) {
				buf := logger.ExecuteWithLock(func() {
					err := printgetmeshCheck(c, &manifest.Manifest{})
					require.NoError(t, err)
				})

				actual := buf.String()
				require.Contains(t, actual, "Please install distributions with tetrate flavor listed in `getmesh list` command")
				require.Contains(t, actual, "nothing to check")
				t.Log(actual)
			})
		}
	})

	t.Run("multiple minor versions", func(t *testing.T) {
		for i, c := range []istioversion.Version{
			{
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.1-tetrate-v1"},
					{IstioVersion: "1.6.1-tetrate-v1"},
				},
			},
			{
				MeshVersion: &istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetratefips-v1"}},
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetrate-v1"}},
				},
			},
			{
				MeshVersion: &istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetratefips-v1"}},
					{Info: istioversion.BuildInfo{Version: "1.20.1-tetrate-v1"}},
				},
			},
			{
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.7.1-tetrate-v1"},
					{IstioVersion: "1.6.1-tetrate-v1"},
				},
				MeshVersion: &istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetrate-v1"}},
					{Info: istioversion.BuildInfo{Version: "1.7.2-tetrate-v1"}},
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetrate-v3"}},
				},
			},
			{
				DataPlaneVersion: &[]istioversion.ProxyInfo{
					{IstioVersion: "1.6.1-tetrate-v1"},
					{IstioVersion: "1.6.1-tetrate-v2"},
				},
				MeshVersion: &istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetrate-v1"}},
					{Info: istioversion.BuildInfo{Version: "1.7.2-tetrate-v1"}},
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetratefips-v3"}},
				},
			},
		} {

			t.Run(fmt.Sprintf("%d-th", i), func(t *testing.T) {
				buf := logger.ExecuteWithLock(func() {
					err := printgetmeshCheck(c, &manifest.Manifest{})
					require.Error(t, err)
				})

				actual := buf.String()
				require.Contains(t, actual, "running in multiple minor versions")
				t.Log(actual)
			})
		}
	})

	t.Run("full", func(t *testing.T) {
		for i, c := range []struct {
			iv  istioversion.Version
			ds  []*manifest.IstioDistribution
			exp []string
		}{
			{
				ds: []*manifest.IstioDistribution{
					{Version: "1.7.10", Flavor: "tetrate", FlavorVersion: 1},
					{Version: "1.7.8", Flavor: "tetrate", FlavorVersion: 2},
					{Version: "1.6.10", Flavor: "tetrate", FlavorVersion: 10},
				},
				iv: istioversion.Version{
					MeshVersion: &istioversion.MeshInfo{
						{Info: istioversion.BuildInfo{Version: "1.7.8-tetrate-v2"}},
						{Info: istioversion.BuildInfo{Version: "1.6.10-tetrate-v10"}},
					},
				},
				exp: []string{`- Your control plane running in multiple minor versions: 1.6-tetrate, 1.7-tetrate`,
					`- There is the available patch for the minor version 1.7-tetrate. We recommend upgrading all 1.7-tetrate versions -> 1.7.10-tetrate-v1`,
					`- 1.6.10-tetrate-v10 is the latest version in 1.6-tetrate`,
				},
			},
			{
				ds: []*manifest.IstioDistribution{
					{Version: "1.8.10", Flavor: "tetrate", FlavorVersion: 10},
					{Version: "1.7.10", Flavor: "tetrate", FlavorVersion: 10},
					{Version: "1.7.8", Flavor: "tetrate", FlavorVersion: 10},
				},
				iv: istioversion.Version{
					MeshVersion: &istioversion.MeshInfo{
						{Info: istioversion.BuildInfo{Version: "1.8.1-tetrate-v10"}},
						{Info: istioversion.BuildInfo{Version: "1.6.1-tetrate-v10"}}, // not supported
					},
					DataPlaneVersion: &[]istioversion.ProxyInfo{
						{IstioVersion: "1.7.1-tetrate-v10"},
						{IstioVersion: "1.8.3-tetrate-v10"},
					},
				},
				exp: []string{`- Your data plane running in multiple minor versions: 1.7-tetrate, 1.8-tetrate`,
					`- Your control plane running in multiple minor versions: 1.6-tetrate, 1.8-tetrate`,
					` There is the available patch for the minor version 1.7-tetrate. We recommend upgrading all 1.7-tetrate versions -> 1.7.10-tetrate-v10`,
					` There is the available patch for the minor version 1.8-tetrate. We recommend upgrading all 1.8-tetrate versions -> 1.8.10-tetrate-v10`,
					` The minor version 1.6-tetrate is no longer supported by getmesh. We recommend you use the higher minor versions in "getmesh list"`,
				},
			},
			{
				ds: []*manifest.IstioDistribution{
					{Version: "1.8.10", Flavor: "tetratefips", FlavorVersion: 1},
					{Version: "1.7.10", Flavor: "tetrate", FlavorVersion: 2},
					{Version: "1.7.8", Flavor: "tetrate", FlavorVersion: 3},
				},
				iv: istioversion.Version{
					MeshVersion: &istioversion.MeshInfo{
						{Info: istioversion.BuildInfo{Version: "1.8.1-tetratefips-v3000"}},
					},
					DataPlaneVersion: &[]istioversion.ProxyInfo{
						{IstioVersion: "1.7.1-tetrate-v10000"},
					},
				},
				exp: []string{`- Your data plane running in the minor version 1.7-tetrate but control plane in 1.8-tetratefips`,
					`- There is the available patch for the minor version 1.7-tetrate. We recommend upgrading all 1.7-tetrate versions -> 1.7.10-tetrate-v2`,
					`- There is the available patch for the minor version 1.8-tetratefips. We recommend upgrading all 1.8-tetratefips versions -> 1.8.10-tetratefips-v1`,
				},
			},
			{
				ds: []*manifest.IstioDistribution{
					{Version: "1.8.10", FlavorVersion: 5, Flavor: "tetratefips"},
					{Version: "1.7.10", FlavorVersion: 1, Flavor: "tetratefips"},
					{Version: "1.7.8", FlavorVersion: 1, Flavor: "tetratefips"},
				},
				iv: istioversion.Version{
					MeshVersion: &istioversion.MeshInfo{
						{Info: istioversion.BuildInfo{Version: "1.8.1-tetratefips-v1000"}},
					},
					DataPlaneVersion: &[]istioversion.ProxyInfo{
						{IstioVersion: "1.8.10-tetratefips-v1"},
					},
				},
				exp: []string{`- There is the available patch for the minor version 1.8-tetratefips. We recommend upgrading all 1.8-tetratefips versions -> 1.8.10-tetratefips-v5`},
			},
		} {
			t.Run(fmt.Sprintf("%d-th", i), func(t *testing.T) {
				buf := logger.ExecuteWithLock(func() {
					err := printgetmeshCheck(c.iv, &manifest.Manifest{IstioDistributions: c.ds})
					require.Error(t, err)
				})

				actual := buf.String()
				for _, e := range c.exp {
					require.Contains(t, actual, e)
				}
				t.Log(actual)
			})
		}
	})
}

func Test_getLatestPatchInManifestMsg(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		ms := []*manifest.IstioDistribution{
			{Version: "1.aaaa"},
			{Version: "1.7.8"},
			{Version: "1.7.9"},
		}
		_, ok, err := getLatestPatchInManifestMsg(&manifest.IstioDistribution{}, &manifest.Manifest{IstioDistributions: ms})
		require.Error(t, err)
		require.False(t, ok)
	})

	t.Run("not supported minot version", func(t *testing.T) {
		ms := []*manifest.IstioDistribution{
			{Version: "1.8.10", FlavorVersion: 20, Flavor: "tetrate"},
			{Version: "1.7.8", FlavorVersion: 10, Flavor: "tetratefips"},
		}
		msg, ok, err := getLatestPatchInManifestMsg(&manifest.IstioDistribution{
			Version: "1.0.100",
			Flavor:  "tetrate", FlavorVersion: 10,
		}, &manifest.Manifest{IstioDistributions: ms})
		require.NoError(t, err)
		require.False(t, ok)
		require.Contains(t, msg, "The minor version 1.0-tetrate is no longer supported by getmesh.")
		require.Contains(t, msg, "getmesh list")
		t.Log(msg)

		msg, ok, err = getLatestPatchInManifestMsg(&manifest.IstioDistribution{
			Version: "1.10.100", Flavor: "tetratefips", FlavorVersion: 10,
		}, &manifest.Manifest{IstioDistributions: ms})
		require.NoError(t, err)
		require.False(t, ok)
		require.Contains(t, msg, "The minor version 1.10-tetratefips is no longer supported by getmesh.")
		require.Contains(t, msg, "getmesh list")
		t.Log(msg)

	})

	t.Run("updated", func(t *testing.T) {
		ms := []*manifest.IstioDistribution{
			{Version: "1.8.10", Flavor: "tetrate", FlavorVersion: 10},
			{Version: "1.7.20", Flavor: "tetratefips", FlavorVersion: 1},
		}

		msg, ok, err := getLatestPatchInManifestMsg(&manifest.IstioDistribution{
			Version: "1.8.10",
			Flavor:  "tetrate", FlavorVersion: 10,
		}, &manifest.Manifest{IstioDistributions: ms})
		require.NoError(t, err)
		require.True(t, ok)
		require.Contains(t, msg, "- 1.8.10-tetrate-v10 is the latest version in 1.8-tetrate")
		t.Log(msg)

		msg, ok, err = getLatestPatchInManifestMsg(&manifest.IstioDistribution{
			Version: "1.7.20",
			Flavor:  "tetratefips", FlavorVersion: 1,
		}, &manifest.Manifest{IstioDistributions: ms})
		require.NoError(t, err)
		require.True(t, ok)
		require.Contains(t, msg, "- 1.7.20-tetratefips-v1 is the latest version in 1.7-tetratefips")
		t.Log(msg)
	})

	t.Run("recommend upgrade", func(t *testing.T) {
		ms := []*manifest.IstioDistribution{
			{Version: "1.8.10", FlavorVersion: 10, Flavor: "tetrate"},
			{Version: "1.8.10", FlavorVersion: 5, Flavor: "tetrate"},
			{Version: "1.8.5", FlavorVersion: 20, Flavor: "tetrate"},
			{Version: "1.7.20", FlavorVersion: 30, Flavor: "tetratefips", IsSecurityPatch: true},
			{Version: "1.7.1", FlavorVersion: 40, Flavor: "tetratefips"},
		}

		for i, v := range []*manifest.IstioDistribution{
			{Version: "1.8.1", Flavor: "tetrate", FlavorVersion: 1000},
			{Version: "1.8.5", Flavor: "tetrate", FlavorVersion: 20},
			{Version: "1.8.10", Flavor: "tetrate", FlavorVersion: 1},
			{Version: "1.8.10", Flavor: "tetrate", FlavorVersion: 5},
		} {
			t.Run(fmt.Sprintf("%d-th_1.8", i), func(t *testing.T) {
				actual, ok, err := getLatestPatchInManifestMsg(v, &manifest.Manifest{IstioDistributions: ms})
				require.NoError(t, err)
				require.False(t, ok)
				require.Contains(t, actual, "all 1.8-tetrate versions -> 1.8.10-tetrate-v10")
				t.Log(actual)
			})

		}

		for i, v := range []*manifest.IstioDistribution{
			{Version: "1.7.1", Flavor: "tetratefips", FlavorVersion: 1000},
			{Version: "1.7.20", Flavor: "tetratefips", FlavorVersion: 1},
		} {
			t.Run(fmt.Sprintf("%d-th_1.8", i), func(t *testing.T) {
				actual, ok, err := getLatestPatchInManifestMsg(v, &manifest.Manifest{IstioDistributions: ms})
				require.NoError(t, err)
				require.False(t, ok)
				require.Contains(t, actual, "which includes **security upgrades**. We strongly recommend upgrading all 1.7-tetratefips versions -> 1.7.20-tetratefips-v30")
				t.Log(actual)
			})
		}
	})
}

func Test_getMultipleMinorVersionRunningMsg(t *testing.T) {
	for _, c := range []struct {
		t   string
		mvs map[string]*manifest.IstioDistribution
		exp string
	}{
		{
			t: "control plane", mvs: map[string]*manifest.IstioDistribution{"1.6-tetrate": {}, "1.7-tetratefips": {}},
			exp: "- Your control plane running in multiple minor versions: 1.6-tetrate, 1.7-tetratefips\n",
		},
		{
			t: "data plane", mvs: map[string]*manifest.IstioDistribution{"1.6-tetrate": {}, "1.7-tetrate": {}, "1.9-tetratefips": {}, "2.1-tetratefips": {}},
			exp: "- Your data plane running in multiple minor versions: 1.6-tetrate, 1.7-tetrate, 1.9-tetratefips, 2.1-tetratefips\n",
		},
	} {
		require.Equal(t, c.exp, getMultipleMinorVersionRunningMsg(c.t, c.mvs))
	}
}

func Test_getControlPlaneVersions(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		actual, err := getControlPlaneVersions(nil)
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run("error", func(t *testing.T) {
		in := istioversion.MeshInfo{{Info: istioversion.BuildInfo{
			Version: "1.3123131",
		}}}
		_, err := getControlPlaneVersions(&in)
		require.Error(t, err)
	})

	t.Run("upstream version", func(t *testing.T) {
		in := istioversion.MeshInfo{{Info: istioversion.BuildInfo{
			Version: "1.7.1",
		}}}

		buf := logger.ExecuteWithLock(func() {
			_, err := getControlPlaneVersions(&in)
			require.NoError(t, err)
		})

		require.Contains(t, buf.String(), "Please install distributions with tetrate flavor listed in")
		t.Log(buf.String())
	})

	t.Run("ok", func(t *testing.T) {
		for i, c := range []struct {
			in  istioversion.MeshInfo
			exp map[string]*manifest.IstioDistribution
		}{
			{
				in: istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.3-tetrate-v1"}},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate": {Version: "1.7.3", Flavor: "tetrate", FlavorVersion: 1},
				},
			},
			{
				in: istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.3-tetrate-v1"}},
					{Info: istioversion.BuildInfo{Version: "1.6.1-tetratefips-v2"}},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate":     {Version: "1.7.3", Flavor: "tetrate", FlavorVersion: 1},
					"1.6-tetratefips": {Version: "1.6.1", Flavor: "tetratefips", FlavorVersion: 2},
				},
			},
			{
				in: istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.1-tetrate-v10"}},
					{Info: istioversion.BuildInfo{Version: "1.7.3-tetrate-v2"}},
					{Info: istioversion.BuildInfo{Version: "1.7.7-tetrate-v4"}},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate": {Version: "1.7.1", Flavor: "tetrate", FlavorVersion: 10},
				},
			},
			{
				in: istioversion.MeshInfo{
					{Info: istioversion.BuildInfo{Version: "1.7.9-tetrate-v100"}},
					{Info: istioversion.BuildInfo{Version: "1.7.3-tetrate-v10"}},
					{Info: istioversion.BuildInfo{Version: "1.6.5-tetratefips-v10"}},
					{Info: istioversion.BuildInfo{Version: "1.6.1-tetratefips-v100"}},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate":     {Version: "1.7.3", Flavor: "tetrate", FlavorVersion: 10},
					"1.6-tetratefips": {Version: "1.6.1", Flavor: "tetratefips", FlavorVersion: 100},
				},
			},
		} {
			t.Run(fmt.Sprintf("%d-th", i), func(t *testing.T) {
				actual, err := getControlPlaneVersions(&c.in)
				require.NoError(t, err)
				require.Equal(t, c.exp, actual)
			})
		}
	})
}

func Test_getDataPlaneVersions(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		actual, err := getDataPlaneVersions(nil)
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run("error", func(t *testing.T) {
		in := []istioversion.ProxyInfo{{IstioVersion: "1.3123131"}}
		_, err := getDataPlaneVersions(&in)
		require.Error(t, err)
	})

	t.Run("upstream version", func(t *testing.T) {
		in := []istioversion.ProxyInfo{{IstioVersion: "1.7.1"}}
		buf := logger.ExecuteWithLock(func() {
			_, err := getDataPlaneVersions(&in)
			require.NoError(t, err)
		})

		require.Contains(t, buf.String(), "Please install distributions with tetrate flavor listed in")
		t.Log(buf.String())
	})

	t.Run("ok", func(t *testing.T) {
		for i, c := range []struct {
			in  []istioversion.ProxyInfo
			exp map[string]*manifest.IstioDistribution
		}{
			{
				in: []istioversion.ProxyInfo{
					{IstioVersion: "1.7.3-tetrate-v0"},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate": {Version: "1.7.3", Flavor: "tetrate", FlavorVersion: 0},
				},
			},
			{
				in: []istioversion.ProxyInfo{
					{IstioVersion: "1.7.3"}, // to be ignored
					{IstioVersion: "1.6.1-tetrate-v100"},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.6-tetrate": {Version: "1.6.1", Flavor: "tetrate", FlavorVersion: 100},
				}},
			{
				in: []istioversion.ProxyInfo{
					{IstioVersion: "1.7.3-tetrate-v0"},
					{IstioVersion: "1.6.1-tetratefips-v1"},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate":     {Version: "1.7.3", Flavor: "tetrate", FlavorVersion: 0},
					"1.6-tetratefips": {Version: "1.6.1", Flavor: "tetratefips", FlavorVersion: 1},
				},
			},
			{
				in: []istioversion.ProxyInfo{
					{IstioVersion: "1.7.1-tetrate-v1"},
					{IstioVersion: "1.7.3-tetrate-v10"},
					{IstioVersion: "1.7.7-tetrate-v100"},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate": {Version: "1.7.1", Flavor: "tetrate", FlavorVersion: 1},
				},
			},
			{
				in: []istioversion.ProxyInfo{
					{IstioVersion: "1.7.1-tetrate-v1"},
					{IstioVersion: "1.7.1-tetrate-v100"},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate": {Version: "1.7.1", Flavor: "tetrate", FlavorVersion: 1},
				},
			},
			{
				in: []istioversion.ProxyInfo{
					{IstioVersion: "1.7.9-tetrate-v1"},
					{IstioVersion: "1.7.3-tetrate-v100"},
					{IstioVersion: "1.6.5-tetrate-v100"},
					{IstioVersion: "1.6.1-tetrate-v1"},
				},
				exp: map[string]*manifest.IstioDistribution{
					"1.7-tetrate": {Version: "1.7.3", Flavor: "tetrate", FlavorVersion: 100},
					"1.6-tetrate": {Version: "1.6.1", Flavor: "tetrate", FlavorVersion: 1},
				},
			},
		} {
			t.Run(fmt.Sprintf("%d-th", i), func(t *testing.T) {
				actual, err := getDataPlaneVersions(&c.in)
				require.NoError(t, err)
				require.Equal(t, c.exp, actual)
			})
		}
	})
}
