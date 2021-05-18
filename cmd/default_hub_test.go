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

package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func Test_defaultHubCheckFlags(t *testing.T) {
	for _, c := range []struct {
		setValue string
		show     bool
		expErr   bool
	}{
		{setValue: "gcr.io", show: false, expErr: false},
		{setValue: "", show: true, expErr: false},
		{setValue: "gcr.io", show: true, expErr: true},
		{setValue: "", show: false, expErr: true},
	} {
		actual := defaultHubCheckFlags(c.setValue, c.show)
		if c.expErr {
			require.Error(t, actual)
		} else {
			require.NoError(t, actual)
		}
	}
}

func Test_defaultHubHandleSet(t *testing.T) {
	getistio.GlobalConfigMux.Lock()
	defer getistio.GlobalConfigMux.Unlock()
	home, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(home)

	w := new(bytes.Buffer)
	logger.Lock()
	defer logger.Unlock()
	logger.SetWriter(w)
	value := "myhub.com"
	require.NoError(t, defaultHubHandleSet(home, value))
	require.Equal(t, value, getistio.GetActiveConfig().DefaultHub)
	require.Contains(t, w.String(), "The default hub is now set to myhub.com")
}

func Test_defaultHubHandleShow(t *testing.T) {
	t.Run("not set", func(t *testing.T) {
		w := new(bytes.Buffer)
		logger.Lock()
		defer logger.Unlock()
		logger.SetWriter(w)
		defaultHubHandleShow("")
		require.Contains(t, w.String(), "The default hub is not set yet. Istioctl's default value is used for \"getistio istioctl install\" command\n")
	})

	t.Run("set", func(t *testing.T) {
		w := new(bytes.Buffer)
		logger.Lock()
		defer logger.Unlock()
		logger.SetWriter(w)
		value := "myhub.com"
		defaultHubHandleShow(value)
		require.Contains(t, w.String(), "The current default hub is set to myhub.com")
	})
}
