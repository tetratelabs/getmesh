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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/internal/getmesh"
	"github.com/tetratelabs/getmesh/internal/util/logger"
)

func Test_defaultHubCheckFlags(t *testing.T) {
	for _, c := range []struct {
		remove   bool
		setValue string
		show     bool
		expErr   bool
	}{
		{remove: false, setValue: "gcr.io", show: false, expErr: false},
		{remove: false, setValue: "", show: true, expErr: false},
		{remove: false, setValue: "gcr.io", show: true, expErr: true},
		{remove: false, setValue: "", show: false, expErr: true},
		{remove: true, setValue: "gcr.io", show: false, expErr: true},
		{remove: true, setValue: "", show: true, expErr: true},
		{remove: true, setValue: "gcr.io", show: true, expErr: true},
		{remove: true, setValue: "", show: false, expErr: false},
	} {
		actual := defaultHubCheckFlags(c.remove, c.setValue, c.show)
		if c.expErr {
			require.Error(t, actual)
		} else {
			require.NoError(t, actual)
		}
	}
}

func Test_defaultHubHandleSet(t *testing.T) {
	getmesh.GlobalConfigMux.Lock()
	defer getmesh.GlobalConfigMux.Unlock()
	home := t.TempDir()

	value := "myhub.com"
	buf := logger.ExecuteWithLock(func() {
		require.NoError(t, defaultHubHandleSet(home, value))
	})
	require.Equal(t, value, getmesh.GetActiveConfig().DefaultHub)
	require.Contains(t, buf.String(), "The default hub is now set to myhub.com")
}

func Test_defaultHubHandleShow(t *testing.T) {
	t.Run("not set", func(t *testing.T) {
		buf := logger.ExecuteWithLock(func() {
			defaultHubHandleShow("")
		})
		require.Contains(t, buf.String(), "The default hub is not set yet. Istioctl's default value is used for \"getmesh istioctl install\" command\n")
	})

	t.Run("set", func(t *testing.T) {
		value := "myhub.com"
		buf := logger.ExecuteWithLock(func() {
			defaultHubHandleShow(value)
		})
		require.Contains(t, buf.String(), "The current default hub is set to myhub.com")
	})
}

func Test_defaultHubHandleRemove(t *testing.T) {
	getmesh.GlobalConfigMux.Lock()
	defer getmesh.GlobalConfigMux.Unlock()
	home := t.TempDir()

	value := "myhub.com"
	require.NoError(t, defaultHubHandleSet(home, value))

	buf := logger.ExecuteWithLock(func() {
		require.NoError(t, defaultHubHandleRemove(home))
	})
	require.Contains(t, buf.String(), "The default hub is removed. Now Istioctl's default value is used for \"getmesh istioctl install\"")
}
