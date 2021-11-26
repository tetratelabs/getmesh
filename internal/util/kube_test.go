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

package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
)

func TestGetKubeConfigLocation(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "")

		config := GetKubeConfigLocation()
		require.Equal(t, config, clientcmd.RecommendedHomeFile)
	})

	t.Run("KUBECONFIG", func(t *testing.T) {
		actual := "./testconfig"
		t.Setenv("KUBECONFIG", actual)

		config := GetKubeConfigLocation()
		require.Equal(t, config, actual)
	})
	t.Run("local", func(t *testing.T) {
		actual := "./testconfig"
		t.Setenv("KUBECONFIG", "")

		KubeConfig = actual
		config := GetKubeConfigLocation()
		require.Equal(t, config, actual)
		KubeConfig = "" //cleanup
	})
}
