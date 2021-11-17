package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TempFile is a helper function which creates a temporary file
// and automatically closes after test is completed
func TempFile(t *testing.T, dir, pattern string) *os.File {
	t.Helper()

	tempFile, err := ioutil.TempFile(dir, pattern)
	require.NoError(t, err)
	require.NotNil(t, tempFile)

	t.Cleanup(func() { require.NoError(t, os.Remove(tempFile.Name())) })
	return tempFile
}

// TempDir is a helper function which creates a temporary directory
// and automatically closes after test is completed
func TempDir(t *testing.T, dir, pattern string) string {
	t.Helper()

	tempDir, err := ioutil.TempDir(dir, pattern)
	require.NoError(t, err)
	require.NotNil(t, tempDir)

	t.Cleanup(func() { require.NoError(t, os.RemoveAll(tempDir)) })
	return tempDir
}
