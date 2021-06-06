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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/src/util/logger"
)

func TestLatestVersion(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`
aaa=aa
1+1
echo "hello world"
GETMESH_LATEST_VERSION="1.1.1"
`))
		}))
		defer ts.Close()

		require.NoError(t, os.Setenv(downloadShellTestURLEnvKey, ts.URL))

		actual, err := LatestVersion()
		require.NoError(t, err)
		assert.Equal(t, "1.1.1", actual)
	})

	t.Run("error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`
aaa=aa
1+1
`))
		}))
		defer ts.Close()
		require.NoError(t, os.Setenv(downloadShellTestURLEnvKey, ts.URL))
		_, err := LatestVersion()
		require.Error(t, err)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("up-to-date with getistio prefix", func(t *testing.T) {
		v := "1.1.1"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(fmt.Sprintf(`GETISTIO_LATEST_VERSION="%s"`, v)))
		}))
		defer ts.Close()
		require.NoError(t, os.Setenv(downloadShellTestURLEnvKey, ts.URL))

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, Update(v))
		})

		actual := buf.String()
		assert.Contains(t, actual, fmt.Sprintf("Your getmesh version is up-to-date: %s", v))
		t.Log(actual)
	})

	t.Run("up-to-date with getmesh prefix", func(t *testing.T) {
		v := "1.1.1"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(fmt.Sprintf(`GETMESH_LATEST_VERSION="%s"`, v)))
		}))
		defer ts.Close()
		require.NoError(t, os.Setenv(downloadShellTestURLEnvKey, ts.URL))

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, Update(v))
		})

		actual := buf.String()
		assert.Contains(t, actual, fmt.Sprintf("Your getmesh version is up-to-date: %s", v))
		t.Log(actual)
	})

	t.Run("download with getistio prefix", func(t *testing.T) {
		msg := "download script executed"
		current, latest := "0.0.0", "0.0.1"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(fmt.Sprintf(`GETISTIO_LATEST_VERSION="%s"
echo "%s"`, latest, msg)))
		}))
		defer ts.Close()
		require.NoError(t, os.Setenv(downloadShellTestURLEnvKey, ts.URL))

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, Update(current))
		})

		actual := buf.String()
		assert.Contains(t, actual, msg)
		assert.Contains(t, actual, fmt.Sprintf("getmesh successfully updated from %s to %s!", current, latest))
		t.Log(actual)
	})

	t.Run("download with getmesh prefix", func(t *testing.T) {
		msg := "download script executed"
		current, latest := "0.0.0", "0.0.1"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(fmt.Sprintf(`GETMESH_LATEST_VERSION="%s"
echo "%s"`, latest, msg)))
		}))
		defer ts.Close()
		require.NoError(t, os.Setenv(downloadShellTestURLEnvKey, ts.URL))

		buf := logger.ExecuteWithLock(func() {
			require.NoError(t, Update(current))
		})

		actual := buf.String()
		assert.Contains(t, actual, msg)
		assert.Contains(t, actual, fmt.Sprintf("getmesh successfully updated from %s to %s!", current, latest))
		t.Log(actual)
	})
}
