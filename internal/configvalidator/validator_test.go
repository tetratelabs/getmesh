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

package configvalidator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidator_filterResults(t *testing.T) {
	t.Run("without-threshold", func(t *testing.T) {
		t.Run("allnamespaces", func(t *testing.T) {
			in := []configValidationResult{
				{namespace: "default"}, {namespace: "kube-system"},
			}

			cv := &ConfigValidator{}
			require.Equal(t, in, cv.filterResults(in))
		})

		t.Run("single", func(t *testing.T) {
			in := []configValidationResult{
				{namespace: "default"}, {namespace: "kube-system"},
			}

			cv := &ConfigValidator{namespace: "kube-system"}
			require.Equal(t, in[1:], cv.filterResults(in))
		})
	})

	t.Run("with-threshold", func(t *testing.T) {
		t.Run("single", func(t *testing.T) {
			in := []configValidationResult{
				{namespace: "default", severity: SeverityLevelError},
				{namespace: "default", severity: SeverityLevelInfo},
				{namespace: "kube-system", severity: SeverityLevelError},
			}
			cv := &ConfigValidator{namespace: "default"}
			require.Equal(t, in[:1], cv.filterResults(in))
		})
		t.Run("allnamespaces", func(t *testing.T) {
			in := []configValidationResult{
				{namespace: "default", severity: SeverityLevelError},
				{namespace: "kube-system", severity: SeverityLevelError},
				{namespace: "default", severity: SeverityLevelInfo},
			}
			cv := &ConfigValidator{}
			require.Equal(t, in[:2], cv.filterResults(in))
		})
	})
}

func TestConfigValidator_allNamespaces(t *testing.T) {
	require.True(t, (&ConfigValidator{namespace: ""}).allNamespaces())
	require.False(t, (&ConfigValidator{namespace: "default"}).allNamespaces())
}
