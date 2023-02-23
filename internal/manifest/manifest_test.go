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

package manifest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/internal/util/logger"
)

func TestFetchManifest(t *testing.T) {
	manifest := &Manifest{
		IstioDistributions: []*IstioDistribution{
			{
				Version:       "1.7.6",
				Flavor:        IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.6",
				Flavor:        IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.7",
				Flavor:        IstioDistributionFlavorIstio,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.5",
				Flavor:        IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
		},
	}

	raw, err := json.Marshal(manifest)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(raw)
	}))
	defer ts.Close()

	actual, err := fetchManifest(ts.URL)
	require.NoError(t, err)

	expIstioVersions := map[string]struct{}{
		"1.7.7-istio-v0":       {},
		"1.7.6-tetrate-v0":     {},
		"1.7.5-tetrate-v0":     {},
		"1.7.6-tetratefips-v0": {},
	}

	for _, a := range actual.IstioDistributions {
		delete(expIstioVersions, a.String())
	}
	require.Equal(t, map[string]struct{}{}, expIstioVersions)
}

func TestPrintManifest(t *testing.T) {
	t.Run("nil-current", func(t *testing.T) {
		manifest := &Manifest{
			IstioDistributions: []*IstioDistribution{
				{
					Version:       "1.7.6",
					Flavor:        IstioDistributionFlavorTetrate,
					FlavorVersion: 0,
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

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, PrintManifest(manifest, nil))
		})
		require.Equal(t, `ISTIO VERSION	FLAVOR 	FLAVOR VERSION	K8S VERSIONS	END OF LIFE
    1.7.6    	tetrate	      0       	    1.16    	
    1.7.5    	tetrate	      0       	    1.16    	
`,
			buf.String())
	})

	t.Run("non-nil-current", func(t *testing.T) {
		current := &IstioDistribution{
			Version:       "1.7.6",
			Flavor:        IstioDistributionFlavorTetrateFIPS,
			FlavorVersion: 0,
			K8SVersions:   []string{"1.16"},
		}
		manifest := &Manifest{
			IstioDistributions: []*IstioDistribution{
				{
					Version:       "1.8.3",
					Flavor:        IstioDistributionFlavorIstio,
					FlavorVersion: 0,
					K8SVersions:   []string{"1.18"},
				},
				{
					Version:       "1.7.6",
					Flavor:        IstioDistributionFlavorTetrateFIPS,
					FlavorVersion: 0,
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

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, PrintManifest(manifest, current))
		})

		require.Equal(t, `ISTIO VERSION	  FLAVOR   	FLAVOR VERSION	K8S VERSIONS	END OF LIFE
    1.8.3    	   istio   	      0       	    1.18    	
   *1.7.6    	tetratefips	      0       	    1.16    	
    1.7.5    	  tetrate  	      0       	    1.16    	
`,
			buf.String())
	})
}
