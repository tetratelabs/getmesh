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

package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TempFile is a helper function which creates a temporary file
// and automatically closes after test is completed
func TempFile(t *testing.T, dir, pattern string) *os.File {
	t.Helper()

	tempFile, err := os.CreateTemp(dir, pattern)
	require.NoError(t, err)
	require.NotNil(t, tempFile)

	t.Cleanup(func() { require.NoError(t, os.Remove(tempFile.Name())) })
	return tempFile
}
