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

package cmd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/internal/manifest"
	"github.com/tetratelabs/getmesh/internal/test"
	"github.com/tetratelabs/getmesh/internal/util"
	"github.com/tetratelabs/getmesh/internal/util/logger"
)

func TestIstioctl_istioctlArgChecks(t *testing.T) {
	manifest.GlobalManifestURLMux.Lock()
	defer manifest.GlobalManifestURLMux.Unlock()

	m := &manifest.Manifest{
		IstioDistributions: []*manifest.IstioDistribution{
			{
				Version:       "1.7.6",
				Flavor:        manifest.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
			{
				Version:       "1.7.5",
				Flavor:        manifest.IstioDistributionFlavorTetrate,
				FlavorVersion: 0,
			},
		},
	}

	raw, err := json.Marshal(m)
	require.NoError(t, err)

	f := test.TempFile(t, "", "")

	_, err = f.Write(raw)
	require.NoError(t, err)

	t.Setenv("GETMESH_TEST_MANIFEST_PATH", f.Name())

	t.Run("ok", func(t *testing.T) {
		out, err := istioctlArgChecks([]string{"analyze"}, nil, "")
		require.NoError(t, err)
		require.Equal(t, []string{"analyze"}, out)

		// Default hub is given but should not affect commands other than "install".
		out, err = istioctlArgChecks([]string{"analyze"}, nil, "gcr.io/istio")
		require.NoError(t, err)
		require.Equal(t, []string{"analyze"}, out)

		out, err = istioctlArgChecks([]string{"install"}, m.IstioDistributions[0], "")
		require.NoError(t, err)
		require.Equal(t, []string{"install"}, out)

		// Default hub is given and should be set to output args.
		out, err = istioctlArgChecks([]string{"install"}, m.IstioDistributions[0], "gcr.io/istio")
		require.NoError(t, err)
		require.Equal(t, []string{"install", "--set", "hub=gcr.io/istio"}, out)

		// Default hub is given but it should not affect the explicitly given hub arg
		out, err = istioctlArgChecks([]string{"install", "--set=hub=my-space.com/istio"}, m.IstioDistributions[0], "gcr.io/istio")
		require.NoError(t, err)
		require.Equal(t, []string{"install", "--set", "hub=my-space.com/istio"}, out)
	})

	t.Run("warning", func(t *testing.T) {
		buf := logger.ExecuteWithLock(func() {
			// confirmation failed so error must be returned
			_, err := istioctlArgChecks([]string{"install"}, &manifest.IstioDistribution{
				Version:       "1.7.4",
				Flavor:        manifest.IstioDistributionFlavorTetrateFIPS,
				FlavorVersion: 0,
			}, "")
			require.Error(t, err)
		})

		require.Contains(t, buf.String(), "Your active istioctl of version 1.7.4-tetratefips-v0 is deprecated.")
		t.Log(buf.String())
	})
}

func TestIstioctl_istioctlParsePreCheckArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		exp  []string
	}{
		{
			name: "default",
			args: []string{"install"},
			exp:  []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation()},
		},
		{
			name: "istioOperator files",
			args: []string{"install", "-f", "a", "--filename", "b"},
			exp:  []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation(), "-f", "a", "--filename", "b"},
		},
		{
			name: "revision",
			args: []string{"install", "--revision", "canary"},
			exp:  []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation(), "--revision", "canary"},
		},
		{
			name: "istio namesapce",
			args: []string{"install", "--set", "values.global.istioNamespace=default"},
			exp:  []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation(), "--istioNamespace", "default"},
		},
		{
			name: "istioOperator files",
			args: []string{"install", "-f", "a", "--filename", "b"},
			exp:  []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation(), "-f", "a", "--filename", "b"},
		},
		{
			name: "help",
			args: []string{"install", "--help"},
			exp:  nil,
		},
		{
			name: "-h",
			args: []string{"install", "-h"},
			exp:  nil,
		},
		{
			name: "eq",
			args: []string{"install", "--set=values.global.istioNamespace=default"},
			exp:  []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation(), "--istioNamespace", "default"},
		},
		{
			name: "full",
			args: []string{"install", "-s=values.global.istioNamespace=default",
				"--revision", "canary", "-f", "a", "--filename=b"},
			exp: []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation(),
				"--istioNamespace", "default", "--revision", "canary", "-f", "a", "--filename", "b"},
		},
		{
			name: "full 2",
			args: []string{"-s=values.global.istioNamespace=default",
				"--revision", "canary", "-f", "a", "--filename=b", "install"},
			exp: []string{"x", "precheck", "--kubeconfig", util.GetKubeConfigLocation(),
				"--istioNamespace", "default", "--revision", "canary", "-f", "a", "--filename", "b"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := istioctlParsePreCheckArgs(c.args)
			require.Equal(t, c.exp, actual)
		})
	}
}

func TestIstioctl_istioctlParseVerifyInstallArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		exp  []string
	}{
		{
			name: "no install",
			args: []string{"analyze"},
			exp:  nil,
		},
		{
			name: "default",
			args: []string{"install"},
			exp:  []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation()},
		},
		{
			name: "istioOperator files",
			args: []string{"install", "-f", "a", "--filename", "b"},
			exp:  []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation(), "-f", "a", "--filename", "b"},
		},
		{
			name: "revision",
			args: []string{"install", "--revision", "canary"},
			exp:  []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation(), "--revision", "canary"},
		},
		{
			name: "istio namesapce",
			args: []string{"install", "--set", "values.global.istioNamespace=default"},
			exp:  []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation(), "--istioNamespace", "default"},
		},
		{
			name: "manifests",
			args: []string{"install", "--manifests", "manifests/"},
			exp:  []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation(), "--manifests", "manifests/"},
		},
		{
			name: "help",
			args: []string{"install", "--help"},
			exp:  nil,
		},
		{
			name: "-h",
			args: []string{"install", "-h"},
			exp:  nil,
		},
		{
			name: "eq",
			args: []string{"install", "--manifests=manifests/", "--set=values.global.istioNamespace=default", "-f=a", "--filename=b"},
			exp: []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation(), "--manifests", "manifests/",
				"--istioNamespace", "default", "-f", "a", "--filename", "b"},
		},
		{
			name: "full",
			args: []string{"install", "--set", "values.global.istioNamespace=default",
				"--revision", "canary", "-f", "a", "--filename", "b", "--manifests", "test/"},
			exp: []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation(),
				"--istioNamespace", "default", "--revision", "canary", "-f", "a", "--filename", "b", "--manifests", "test/"},
		},
		{
			name: "full 2",
			args: []string{"--set", "values.global.istioNamespace=default",
				"--revision", "canary", "-f", "a", "--filename", "b", "--manifests", "test/", "install"},
			exp: []string{"verify-install", "--kubeconfig", util.GetKubeConfigLocation(),
				"--istioNamespace", "default", "--revision", "canary", "-f", "a", "--filename", "b", "--manifests", "test/"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := istioctlParseVerifyInstallArgs(c.args)
			require.Equal(t, c.exp, actual)
		})
	}
}
func TestIstioctl_istioctPatchVersionCheck(t *testing.T) {
	t.Run("old-patch-version", func(t *testing.T) {
		m := &manifest.Manifest{
			IstioDistributions: []*manifest.IstioDistribution{
				{
					Version:       "1.7.6",
					Flavor:        manifest.IstioDistributionFlavorTetrate,
					FlavorVersion: 0,
				},
				{
					Version:       "1.7.5",
					Flavor:        manifest.IstioDistributionFlavorTetrate,
					FlavorVersion: 0,
				},
			},
		}
		current := &manifest.IstioDistribution{
			Version:       "1.7.5",
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}
		buf := logger.ExecuteWithLock(func() {
			// confirmation failed so error must be returned
			require.Error(t, istioctlPatchVersionCheck(current, m))
		})

		require.Contains(t, buf.String(), "your current patch version 1.7.5 is not the latest version 1.7.6")
		t.Log(buf.String())
	})

	t.Run("nil case", func(t *testing.T) {
		m := &manifest.Manifest{
			IstioDistributions: []*manifest.IstioDistribution{
				{
					Version:       "1.7.6",
					Flavor:        manifest.IstioDistributionFlavorTetrate,
					FlavorVersion: 0,
				},
				{
					Version:       "1.7.5",
					Flavor:        manifest.IstioDistributionFlavorTetrate,
					FlavorVersion: 0,
				},
			},
		}
		current := &manifest.IstioDistribution{
			Version:       "1.8.2",
			Flavor:        manifest.IstioDistributionFlavorTetrate,
			FlavorVersion: 0,
		}
		buf := logger.ExecuteWithLock(func() {
			// confirmation failed so error must be returned
			require.NoError(t, istioctlPatchVersionCheck(current, m))
		})
		t.Log(buf.String())
	})
}

func TestIstioctl_istioctlPreProcessArgs(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		wants []string
	}{
		{
			name:  "double dash",
			args:  []string{"--manifests=testfile"},
			wants: []string{"--manifests", "testfile"},
		},
		{
			name:  "single dash",
			args:  []string{"-d=testfile"},
			wants: []string{"-d", "testfile"},
		},
		{
			name:  "with --set",
			args:  []string{"--set", "profile=demo"},
			wants: []string{"--set", "profile=demo"},
		},
		{
			name:  "complex --set",
			args:  []string{"--set=profile=demo"},
			wants: []string{"--set", "profile=demo"},
		},
		{
			name:  "complex --set2",
			args:  []string{"--set=hub=docker-hub.io/istio"},
			wants: []string{"--set", "hub=docker-hub.io/istio"},
		},
		{
			name:  "with -s",
			args:  []string{"-s", "profile=demo"},
			wants: []string{"-s", "profile=demo"},
		},
		{
			name:  "complex -s",
			args:  []string{"-s=profile=demo"},
			wants: []string{"-s", "profile=demo"},
		},
		{
			name:  "with dot",
			args:  []string{"-s=values.option1=true"},
			wants: []string{"-s", "values.option1=true"},
		},
		{
			name:  "with directory",
			args:  []string{"-d=dir1/"},
			wants: []string{"-d", "dir1/"},
		},
		{
			name:  "integrate tests 1",
			args:  []string{"--set", "profile=demo", "--skip-confirmation", "--manifests=testfile"},
			wants: []string{"--set", "profile=demo", "--skip-confirmation", "--manifests", "testfile"},
		},
		{
			name:  "integrate tests 2",
			args:  []string{"--set", "profile=demo", "--skip-confirmation", "--manifests=testfile/"},
			wants: []string{"--set", "profile=demo", "--skip-confirmation", "--manifests", "testfile/"},
		},
		{
			name:  "integrate tests 3",
			args:  []string{"--set=profile=demo", "--skip-confirmation", "--manifests=testfile/", "--set=values.test.foo=bar"},
			wants: []string{"--set", "profile=demo", "--skip-confirmation", "--manifests", "testfile/", "--set", "values.test.foo=bar"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := istioctlPreProcessArgs(test.args)
			require.Equal(t, test.wants, actual)
		})
	}
}
