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

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/tetratelabs/getmesh/internal/test"

	"github.com/stretchr/testify/require"
)

func TestGetIstioHomeDir(t *testing.T) {
	t.Run("not created", func(t *testing.T) {
		dir := test.TempDir(t, "", "")

		actual, err := getmeshHomeDir(dir)
		require.NoError(t, err)
		require.Equal(t, filepath.Join(dir, getmeshDirname), actual)
	})

	t.Run("created", func(t *testing.T) {
		dir := test.TempDir(t, "", "")

		// create .getmesh prior to calling getmeshHomeDir
		home := filepath.Join(dir, getmeshDirname)
		require.NoError(t, os.Mkdir(home, 0755))
		filePath := filepath.Join(home, "tmp.txt")
		f, err := os.Create(filePath)
		require.NoError(t, err)
		expBytes := []byte("this is my file")
		_, err = f.Write(expBytes)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		actual, err := getmeshHomeDir(dir)
		require.NoError(t, err)
		require.Equal(t, home, actual)

		// verify the existing directory left intact
		b, err := ioutil.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, expBytes, b)
	})
}
