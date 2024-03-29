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
	"io/ioutil"
	"testing"

	"github.com/tetratelabs/getmesh/internal/test"

	"github.com/stretchr/testify/require"
)

func TestExtractYamlFilePaths(t *testing.T) {
	root := t.TempDir()

	var in, exp []string
	f := test.TempFile(t, root, "*.yaml")
	in = append(in, f.Name())
	exp = append(exp, f.Name())

	sub, err := ioutil.TempDir(root, "")
	require.NoError(t, err)
	in = append(in, sub)

	f = test.TempFile(t, sub, "*.yaml")
	exp = append(exp, f.Name())

	f = test.TempFile(t, sub, "*.go")
	excluded := f.Name()

	actual, err := extractYamlFilePaths([]string{root})
	require.NoError(t, err)

	for _, e := range exp {
		require.Contains(t, actual, e)
	}

	require.NotContains(t, actual, excluded)
	t.Logf("in: %v, out: %v", in, actual)
}
