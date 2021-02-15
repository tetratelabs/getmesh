package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifest_GetManifestEOLDates(t *testing.T) {
	// This test virtually validates the HEAD manifest.json's `istio_minor_versions_eof_dates` field.
	raw, err := ioutil.ReadFile("../manifest.json")
	require.NoError(t, err)

	var ms Manifest
	require.NoError(t, json.Unmarshal(raw, &ms))

	actual, err := ms.GetEOLDates()
	require.NoError(t, err)
	t.Log(actual)
}

func Test_parseManifestEOLDate(t *testing.T) {
	t.Run("ng", func(t *testing.T) {
		for _, d := range []string{
			"2020-", "2020-01-", "2020-01",
		} {
			_, err := parseManifestEOLDate(d)
			require.Error(t, err)
		}
	})

	t.Run("ok", func(t *testing.T) {
		for _, c := range []struct {
			in                        string
			expDay, expMonth, expYear int
		}{
			{in: "2021-01-12", expYear: 2021, expMonth: 1, expDay: 12},
			{in: "2022-12-01", expYear: 2022, expMonth: 12, expDay: 1},
			{in: "2022-03-01", expYear: 2022, expMonth: 3, expDay: 1},
			{in: "2020-11-21", expYear: 2020, expMonth: 11, expDay: 21},
		} {

			a, err := parseManifestEOLDate(c.in)
			require.NoError(t, err)
			y, m, d := a.Date()
			assert.Equal(t, c.expYear, y)
			assert.Equal(t, c.expMonth, int(m))
			assert.Equal(t, c.expDay, d)
		}
	})
}

func TestIstioDistribution_ToString(t *testing.T) {
	for _, c := range []struct {
		in  *IstioDistribution
		exp string
	}{
		{
			in: &IstioDistribution{
				Version:       "1.7.5",
				Flavor:        IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			exp: "1.7.5-tetrate-v0",
		},
		{
			in: &IstioDistribution{
				Version:       "1.7.7",
				Flavor:        IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 15,
			},
			exp: "1.7.7-tetratefips-v15",
		},
		{
			in: &IstioDistribution{
				Version:       "1.8.3",
				Flavor:        IstioDistributionFlavorIstio,
				FlavorVersion: 0,
			},
			exp: "1.8.3-istio-v0",
		},
	} {
		assert.Equal(t, c.exp, c.in.ToString())
	}
}

func TestIstioDistributionEqual(t *testing.T) {
	operand := &IstioDistribution{
		Version:       "1.7.7",
		Flavor:        IstioDistributionFlavorTetrateFIPS,
		FlavorVersion: 15,
	}

	for _, c := range []struct {
		in  *IstioDistribution
		exp bool
	}{
		{
			in: &IstioDistribution{
				Version:       "1.7.5",
				Flavor:        IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			exp: false,
		},
		{
			in: &IstioDistribution{
				Version:       "1.7.7",
				Flavor:        IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 15,
			},
			exp: true,
		},
		{
			in: &IstioDistribution{
				Version:       "1.8.3",
				Flavor:        IstioDistributionFlavorIstio,
				FlavorVersion: 0,
			},
			exp: false,
		},
	} {
		assert.Equal(t, c.exp, c.in.Equal(operand))
	}
}

func TestIstioDistribution_ExistInManifest(t *testing.T) {
	ms := &Manifest{
		IstioDistributions: []*IstioDistribution{
			{
				Version:       "1.8.1",
				Flavor:        IstioDistributionFlavorTetrate,
				FlavorVersion: 10,
				K8SVersions:   []string{"1.16"},
			},
			{
				Version:       "1.7.5",
				Flavor:        IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
				K8SVersions:   []string{"1.16"},
			},
		},
	}

	d := &IstioDistribution{
		Version:       "1.8.1",
		Flavor:        IstioDistributionFlavorTetrate,
		FlavorVersion: 10,
	}

	ok, err := d.ExistInManifest(ms)
	require.NoError(t, err)
	assert.True(t, ok)

	d.FlavorVersion--
	ok, err = d.ExistInManifest(ms)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestIstioDistribution_Group(t *testing.T) {
	for _, c := range []struct {
		exp string
		in  *IstioDistribution
	}{
		{exp: "1.3-tetrate", in: &IstioDistribution{Version: "1.3.1", Flavor: "tetrate"}},
		{exp: "1.7-tetratefips", in: &IstioDistribution{Version: "1.7.6", Flavor: "tetratefips"}},
		{exp: "1.8-istio", in: &IstioDistribution{Version: "1.8.3", Flavor: "istio"}},
	} {
		actual, err := c.in.Group()
		require.NoError(t, err)
		assert.Equal(t, c.exp, actual)
	}
}

func TestIstioDistribution_IsUpstream(t *testing.T) {
	assert.True(t, (&IstioDistribution{Flavor: "istio"}).IsUpstream())
	assert.False(t, (&IstioDistribution{Flavor: "tetrate"}).IsUpstream())
	assert.False(t, (&IstioDistribution{Flavor: "tetratefips"}).IsUpstream())
}

func TestIstioDistribution_GreaterThan(t *testing.T) {
	base := &IstioDistribution{Version: "1.7.30", FlavorVersion: 40}
	t.Run("true", func(t *testing.T) {
		for _, c := range []*IstioDistribution{
			{Version: base.Version, FlavorVersion: base.FlavorVersion - 1},
			{Version: "1.7.20", FlavorVersion: base.FlavorVersion},
		} {
			actual, err := base.GreaterThan(c)
			require.NoError(t, err)
			assert.True(t, actual)
		}
	})

	t.Run("false", func(t *testing.T) {
		for _, c := range []*IstioDistribution{
			{Version: base.Version, FlavorVersion: base.FlavorVersion + 1},
			{Version: "1.7.50", FlavorVersion: base.FlavorVersion},
		} {
			actual, err := base.GreaterThan(c)
			require.NoError(t, err)
			assert.False(t, actual)
		}
	})
}

func TestIstioDistribution_Equal(t *testing.T) {
	base := &IstioDistribution{Version: "1.2.3", Flavor: "tetrate", FlavorVersion: 40}
	t.Run("true", func(t *testing.T) {
		assert.True(t, base.Equal(&IstioDistribution{Version: "1.2.3", Flavor: "tetrate", FlavorVersion: 40}))
	})

	t.Run("false", func(t *testing.T) {
		for i, c := range []*IstioDistribution{
			{Version: "100.2.3", Flavor: "tetrate", FlavorVersion: 4},
			{Version: "1.200.3", Flavor: "tetrate", FlavorVersion: 4},
			{Version: "1.2.300", Flavor: "tetrate", FlavorVersion: 4},
			{Version: "1.2.3", Flavor: "tetratefips", FlavorVersion: 4},
			{Version: "1.2.3", Flavor: "tetrate", FlavorVersion: 1},
			{Version: "1.2.3", Flavor: "istio", FlavorVersion: 4},
		} {
			require.False(t, base.Equal(c), fmt.Sprintf("%d-th", i))
		}
	})
}

func TestIstioDistributionFromString(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		for _, c := range []struct {
			in  string
			exp *IstioDistribution
		}{
			{in: "1.7.3-tetrate-v0",
				exp: &IstioDistribution{Version: "1.7.3", Flavor: "tetrate", FlavorVersion: 0}},
			{in: "1.1000.3-tetratefips-v1",
				exp: &IstioDistribution{Version: "1.1000.3", Flavor: "tetratefips", FlavorVersion: 1}},
			{in: "1.8.3-istio-v0",
				exp: &IstioDistribution{Version: "1.8.3", Flavor: "istio", FlavorVersion: 0}},
			{in: "1.7.30-tetratefips-v100",
				exp: &IstioDistribution{Version: "1.7.30", Flavor: "tetratefips", FlavorVersion: 100}},
			{in: "2001.7.3-tetrate-v0",
				exp: &IstioDistribution{Version: "2001.7.3", Flavor: "tetrate", FlavorVersion: 0}},
			{in: "2001.7.300-tetratefips-v10",
				exp: &IstioDistribution{Version: "2001.7.300", Flavor: "tetratefips", FlavorVersion: 10}},
		} {
			v, err := IstioDistributionFromString(c.in)
			require.NoError(t, err, c.in, c.in)
			assert.Equal(t, c.exp, v)
		}
	})

	t.Run("ng", func(t *testing.T) {
		for _, in := range []string{
			"1.6", "1.7.113r",
			"1.6.7-", "1.7.113r-tetrate-v1", "1.7.113-tetrate",
			"1.6.7-tetrate-v", "1.7.113-tetrate-",
		} {
			_, err := IstioDistributionFromString(in)
			assert.Error(t, err, in)
		}
	})
}

func Test_parseFlavor(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		for _, c := range []struct {
			in, flavor    string
			flavorVersion int64
		}{
			{in: "tetrate-v0", flavor: "tetrate", flavorVersion: 0},
			{in: "tetratefips-v100", flavor: "tetratefips", flavorVersion: 100},
			{in: "istio-v0", flavor: "istio", flavorVersion: 0},
		} {
			flavor, flavorVersion, err := parseFlavor(c.in)
			require.NoError(t, err)
			assert.Equal(t, c.flavor, flavor)
			assert.Equal(t, c.flavorVersion, flavorVersion)
		}
	})

	t.Run("ng", func(t *testing.T) {
		for _, c := range []string{
			"1.7.10.", "1.7.", "tetrate-",
			"tetrate-v", "v1",
		} {
			_, _, err := parseFlavor(c)
			require.Error(t, err)
			t.Log(err.Error())
		}
	})
}

func TestGetLatestDistribution(t *testing.T) {
	tests := []struct {
		name        string
		maniest     *Manifest
		current     *IstioDistribution
		wants       *IstioDistribution
		wantsSecure bool
	}{
		{
			name: "ok",
			maniest: &Manifest{
				IstioDistributions: []*IstioDistribution{
					{
						Version:         "1.8.3",
						Flavor:          IstioDistributionFlavorIstio,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.17"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.8.2",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.7.6",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.7.5",
						Flavor:          IstioDistributionFlavorTetrate,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
				},
			},
			current: &IstioDistribution{
				Version:         "1.8.2",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   0,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wants: &IstioDistribution{
				Version:         "1.8.2",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   0,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wantsSecure: false,
		},
		{
			name: "old patch",
			maniest: &Manifest{
				IstioDistributions: []*IstioDistribution{
					{
						Version:         "1.8.2",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.7.6",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: true,
					},
					{
						Version:         "1.7.5",
						Flavor:          IstioDistributionFlavorTetrate,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.8.3",
						Flavor:          IstioDistributionFlavorIstio,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
				},
			},
			current: &IstioDistribution{
				Version:         "1.7.5",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   0,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wants: &IstioDistribution{
				Version:         "1.7.6",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   0,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: true,
			},
			wantsSecure: true,
		},
		{
			name: "old flavor",
			maniest: &Manifest{
				IstioDistributions: []*IstioDistribution{
					{
						Version:         "1.7.6",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   1,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.7.6",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.7.5",
						Flavor:          IstioDistributionFlavorTetrate,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.8.3",
						Flavor:          IstioDistributionFlavorIstio,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.18"},
						IsSecurityPatch: false,
					},
				},
			},
			current: &IstioDistribution{
				Version:         "1.7.6",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   0,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wants: &IstioDistribution{
				Version:         "1.7.6",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   1,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wantsSecure: false,
		},
		{
			name: "secure patch true case",
			maniest: &Manifest{
				IstioDistributions: []*IstioDistribution{
					{
						Version:         "1.8.2",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   1,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: true,
					},
					{
						Version:         "1.7.3",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   1,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.7.2",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: true,
					},
					{
						Version:         "1.8.4",
						Flavor:          IstioDistributionFlavorIstio,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.17"},
						IsSecurityPatch: true,
					},
				},
			},
			current: &IstioDistribution{
				Version:         "1.7.1",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   0,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wants: &IstioDistribution{
				Version:         "1.7.3",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   1,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wantsSecure: true,
		},
		{
			name: "secure patch false case",
			maniest: &Manifest{
				IstioDistributions: []*IstioDistribution{
					{
						Version:         "1.7.3",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   1,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.7.2",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: true,
					},
				},
			},
			current: &IstioDistribution{
				Version:         "1.7.3",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   0,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wants: &IstioDistribution{
				Version:         "1.7.3",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   1,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wantsSecure: false,
		},
		{
			name: "nil case",
			maniest: &Manifest{
				IstioDistributions: []*IstioDistribution{
					{
						Version:         "1.7.3",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   1,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: false,
					},
					{
						Version:         "1.7.2",
						Flavor:          IstioDistributionFlavorTetrateFIPS,
						FlavorVersion:   0,
						K8SVersions:     []string{"1.16"},
						IsSecurityPatch: true,
					},
				},
			},
			current: &IstioDistribution{
				Version:         "1.8.2",
				Flavor:          IstioDistributionFlavorTetrateFIPS,
				FlavorVersion:   0,
				K8SVersions:     []string{"1.16"},
				IsSecurityPatch: false,
			},
			wants:       nil,
			wantsSecure: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			latest, isSecure, err := GetLatestDistribution(test.current, test.maniest)
			assert.NoError(t, err)
			if latest == nil {
				assert.Equal(t, latest, test.wants)
			} else {
				assert.True(t, latest.Equal(test.wants))
			}
			assert.Equal(t, test.wantsSecure, isSecure)
		})
	}
}
