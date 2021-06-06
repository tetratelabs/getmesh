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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/api"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

func TestFetchManifest(t *testing.T) {
	manifest := &api.Manifest{
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
				Version:       "1.7.7",
				Flavor:        api.IstioDistributionFlavorIstio,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.5",
				Flavor:        api.IstioDistributionFlavorTetrate,
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
		delete(expIstioVersions, a.ToString())
	}
	assert.Equal(t, map[string]struct{}{}, expIstioVersions)
}

func TestPrintManifest(t *testing.T) {
	t.Run("nil-current", func(t *testing.T) {
		manifest := &api.Manifest{
			IstioDistributions: []*api.IstioDistribution{
				{
					Version:       "1.7.6",
					Flavor:        api.IstioDistributionFlavorTetrate,
					FlavorVersion: 0,
					K8SVersions:   []string{"1.16"},
				},
				{
					Version:       "1.7.5",
					Flavor:        api.IstioDistributionFlavorTetrate,
					FlavorVersion: 0,
					K8SVersions:   []string{"1.16"},
				},
			},
		}

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, PrintManifest(manifest, nil))
		})
		assert.Equal(t, `ISTIO VERSION	FLAVOR 	FLAVOR VERSION	K8S VERSIONS 
    1.7.6    	tetrate	      0       	    1.16    	
    1.7.5    	tetrate	      0       	    1.16    	
`,
			buf.String())
	})

	t.Run("non-nil-current", func(t *testing.T) {
		current := &api.IstioDistribution{
			Version:       "1.7.6",
			Flavor:        api.IstioDistributionFlavorTetrateFIPS,
			FlavorVersion: 0,
			K8SVersions:   []string{"1.16"},
		}
		manifest := &api.Manifest{
			IstioDistributions: []*api.IstioDistribution{
				{
					Version:       "1.8.3",
					Flavor:        api.IstioDistributionFlavorIstio,
					FlavorVersion: 0,
					K8SVersions:   []string{"1.18"},
				},
				{
					Version:       "1.7.6",
					Flavor:        api.IstioDistributionFlavorTetrateFIPS,
					FlavorVersion: 0,
					K8SVersions:   []string{"1.16"},
				},
				{
					Version:       "1.7.5",
					Flavor:        api.IstioDistributionFlavorTetrate,
					FlavorVersion: 0,
					K8SVersions:   []string{"1.16"},
				},
			},
		}

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, PrintManifest(manifest, current))
		})

		assert.Equal(t, `ISTIO VERSION	  FLAVOR   	FLAVOR VERSION	K8S VERSIONS 
    1.8.3    	   istio   	      0       	    1.18    	
   *1.7.6    	tetratefips	      0       	    1.16    	
    1.7.5    	  tetrate  	      0       	    1.16    	
`,
			buf.String())
	})
}
