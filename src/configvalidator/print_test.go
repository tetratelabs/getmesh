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

	"github.com/tetratelabs/getmesh/src/util/logger"
)

func TestPrintConfigValidationResultsWithoutNamespace(t *testing.T) {
	res := configValidationResult{
		name:         "xxxxxxxxxx",
		namespace:    "yyyyyyyyyy",
		errorCode:    "zzzzzzzzzz",
		resourceType: "Aaaaaaaaaa",
		message:      "bbbbbbbbbb",
	}

	buf := logger.ExecuteWithLock(func() {
		printResultsWithoutNamespace([]configValidationResult{res})
	})
	actual := buf.String()

	exps := append(tableColumns[1:], res.name, res.errorCode, res.resourceType, res.message, res.severity.Name)
	for _, c := range exps {
		require.Contains(t, actual, c)
	}

	require.NotContains(t, actual, res.namespace)
	require.NotContains(t, actual, "NAMESPACE")
	t.Log(actual)

}

func TestPrintConfigValidationResults(t *testing.T) {
	res := configValidationResult{
		name:         "xxxxxxxxxx",
		namespace:    "yyyyyyyyyy",
		errorCode:    "zzzzzzzzzz",
		resourceType: "Aaaaaaaaaa",
		message:      "bbbbbbbbbb",
	}

	buf := logger.ExecuteWithLock(func() {
		printResults([]configValidationResult{res})
	})

	actual := buf.String()
	exps := append(tableColumns, res.namespace, res.name, res.errorCode, res.resourceType, res.message, res.severity.Name)
	for _, c := range exps {
		require.Contains(t, actual, c)
	}

	t.Log(actual)
}
