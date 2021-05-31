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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/api"
	"github.com/tetratelabs/getmesh/src/getmesh"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

func Test_endOfLifeChecker(t *testing.T) {
	home, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	getmesh.GlobalConfigMux.Lock()
	defer getmesh.GlobalConfigMux.Unlock()

	m := &api.Manifest{
		IstioDistributions: []*api.IstioDistribution{
			{Version: "1.8.1", Flavor: api.IstioDistributionFlavorTetrate, FlavorVersion: 0},
			{Version: "1.9.10", Flavor: api.IstioDistributionFlavorTetrateFIPS, FlavorVersion: 0},
			{Version: "1.9.0", Flavor: api.IstioDistributionFlavorIstio, FlavorVersion: 0},
		},
		IstioMinorVersionsEolDates: map[string]string{
			"1.7": "2020-10-10",
			"1.6": "2020-10-10",
		},
	}

	raw, err := json.Marshal(m)
	require.NoError(t, err)

	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer f.Close()

	_, err = f.Write(raw)
	require.NoError(t, err)
	require.NoError(t, os.Setenv("GETMESH_TEST_MANIFEST_PATH", f.Name()))

	t.Run("ok version", func(t *testing.T) {
		require.NoError(t, getmesh.SetIstioVersion(home, &api.IstioDistribution{Version: "1.8.1"}))
		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, endOfLifeCheckerImpl(m, time.Now()))
		})

		assert.Equal(t, "", buf.String())
	})

	t.Run("ok time", func(t *testing.T) {
		require.NoError(t, getmesh.SetIstioVersion(home, &api.IstioDistribution{Version: "1.7.1"}))
		buf := logger.ExecuteWithLock(func() {
			now := time.Date(2020, 9, 5, 0, 0, 0, 0, time.Local)
			require.NoError(t, endOfLifeCheckerImpl(m, now))
		})

		assert.Equal(t, "", buf.String())
	})

	t.Run("warn", func(t *testing.T) {
		now := time.Date(2020, 11, 5, 0, 0, 0, 0, time.Local)
		exp := `[WARNING] Your current active minor version %s is reaching the end of life on 2020-10-10. We strongly recommend you to upgrade to the available higher minor versions: 1.8.1-tetrate-v0, 1.9.10-tetratefips-v0, 1.9.0-istio-v0.`
		for _, c := range []struct {
			version, minorVersion string
		}{{version: "1.7.1", minorVersion: "1.7"}, {version: "1.6.100", minorVersion: "1.6"}} {
			t.Run(c.version, func(t *testing.T) {
				require.NoError(t, getmesh.SetIstioVersion(home, &api.IstioDistribution{Version: c.version}))
				buf := logger.ExecuteWithLock(func() {
					require.NoError(t, endOfLifeCheckerImpl(m, now))
				})

				assert.Contains(t, buf.String(), fmt.Sprintf(exp, c.minorVersion))
				t.Log(buf.String())
			})

		}
	})
}
