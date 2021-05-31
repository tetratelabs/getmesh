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

package getmesh

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/api"
)

func TestSetIstioVersion(t *testing.T) {
	GlobalConfigMux.Lock()
	defer GlobalConfigMux.Unlock()
	home, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	d := &api.IstioDistribution{
		Version:       "1.8.1",
		Flavor:        api.IstioDistributionFlavorTetrate,
		FlavorVersion: 0,
	}

	require.NoError(t, SetIstioVersion(home, d))

	b, err := ioutil.ReadFile(getConfigPath(home))
	require.NoError(t, err)
	var actual Config
	require.NoError(t, json.Unmarshal(b, &actual))
	assert.Equal(t, d, actual.IstioDistribution)
}

func TestSetDefaultHub(t *testing.T) {
	GlobalConfigMux.Lock()
	defer GlobalConfigMux.Unlock()
	home, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	hub := "gcr.io/istio-testing"
	require.NoError(t, SetDefaultHub(home, hub))
	b, err := ioutil.ReadFile(getConfigPath(home))
	require.NoError(t, err)
	var actual Config
	require.NoError(t, json.Unmarshal(b, &actual))
	assert.Equal(t, hub, actual.DefaultHub)
}

func TestInitConfig(t *testing.T) {
	GlobalConfigMux.Lock()
	defer GlobalConfigMux.Unlock()
	t.Run("exists", func(t *testing.T) {
		home, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(home)

		d := &api.IstioDistribution{
			Version:       "1.8.1",
			Flavor:        api.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}

		require.NoError(t, SetIstioVersion(home, d))
		currentConfig = Config{} // clear

		require.NoError(t, InitConfig(home))
		assert.Equal(t, currentConfig.IstioDistribution, d)
	})

	t.Run("non-exists", func(t *testing.T) {
		home, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(home)

		currentConfig = Config{IstioDistribution: &api.IstioDistribution{}}
		require.NoError(t, InitConfig(home))
		assert.Nil(t, currentConfig.IstioDistribution)
	})

}

func Test_getConfigPath(t *testing.T) {
	home := "this_is_home"
	assert.Equal(t, filepath.Join(home, "config.json"), getConfigPath(home))
}
