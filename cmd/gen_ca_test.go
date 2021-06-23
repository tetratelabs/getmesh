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
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/tetratelabs/getmesh/src/cacerts/providers/config"
)

func TestPreFlightChecks(t *testing.T) {
	t.Run("Secret File Path", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.SetDefaultValues()

		d, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(d)

		cfg.CertParameters.SecretFilePath = filepath.Join(d, "non-exist")

		require.NoError(t, genCAPreFlightChecks(cfg, nil))
		file, err := ioutil.TempFile(d, "")
		require.NoError(t, err)
		defer file.Close()

		cfg.CertParameters.SecretFilePath = path.Join(d, file.Name())
		require.Error(t, genCAPreFlightChecks(cfg, nil))
	})

	t.Run("Disable Secret Creation", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		cfg := &config.Config{}
		cfg.SetDefaultValues()
		cfg.DisableSecretCreation = false
		err := genCAPreFlightChecks(cfg, cs)
		require.Contains(t, err, "namespaces \"istio-system\" not found")

		ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "istio-system"}}
		_, err = cs.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		require.NoError(t, err)

		err = genCAPreFlightChecks(cfg, cs)
		require.NoError(t, err)
	})

	t.Run("genCAValidateSecretFilePath", func(t *testing.T) {
		d, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(d)

		cs := fake.NewSimpleClientset()
		cfg := &config.Config{}
		cfg.DisableSecretCreation = true
		require.Equal(t, genCAPreFlightChecks(cfg, cs), nil)

		// readonly error
		ro := filepath.Join(d, "readonly")
		require.NoError(t, os.Mkdir(ro, 0400))
		cfg.CertParameters.SecretFilePath = filepath.Join(ro, "test.yaml")
		err = genCAPreFlightChecks(cfg, cs)
		require.Contains(t, err, "unable to write on secret file path:")

		// ok
		f, err := ioutil.TempFile(d, "")
		require.NoError(t, err)
		cfg.CertParameters.SecretFilePath = f.Name()
		err = genCAPreFlightChecks(cfg, cs)
		require.Contains(t, err, f.Name()+"` already exist, please change the file path before proceeding")
	})
}

func TestFetchParametersError(t *testing.T) {
	cs := []struct {
		// stuff to be passed to the binary in cli
		arguments []string
		// the string that is expected in the error
		expected string
		label    string
	}{{
		arguments: []string{"--provider", "azure"},
		expected:  "`azure` provider yet to be implement",
		label:     "Missing Provider"},
	}

	flags := pflag.NewFlagSet("Gen CA flags", pflag.ContinueOnError)
	genCAProviderParameters(flags)
	genCAx509CertRequestParameters(flags)

	for _, c := range cs {
		t.Run(c.label, func(t *testing.T) {
			require.NoError(t, flags.Parse(c.arguments))
			_, err := genCAFetchParameters(flags)
			require.Contains(t, err, c.expected)
		})
	}
}

func TestFetchParametersSucess(t *testing.T) {
	cs := []struct {
		// stuff we would passing to the binary in cli
		arguments []string
		// a function to verify the contents of the config
		checker func(c *config.Config) bool
		label   string
	}{{
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy"},
		checker: func(c *config.Config) bool {
			return c.ProviderName == "aws" && c.ProviderConfig.AWSConfig.SigningCA == "dummy"
		},
		label: "AWS provider check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--template-arn", "dummy-arn"},
		checker: func(c *config.Config) bool {
			return c.ProviderConfig.AWSConfig.TemplateARN == "dummy-arn"
		},
		label: "AWS provider template urn check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--signing-algorithm", "dummy-algorithm"},
		checker: func(c *config.Config) bool {
			return c.ProviderConfig.AWSConfig.SigningAlgorithm == "dummy-algorithm"
		},
		label: "AWS provider signing algorithm check",
	}, {
		arguments: []string{"--provider", "gcp", "--cas-ca-name", "dummy"},
		checker: func(c *config.Config) bool {
			return c.ProviderName == "gcp" && c.ProviderConfig.GCPConfig.CASCAName == "dummy"
		},
		label: "GCP provider check",
	}, {
		arguments: []string{"--provider", "gcp", "--cas-ca-name", "dummy", "--max-issuer-path-len", "93"},
		checker: func(c *config.Config) bool {
			return c.ProviderConfig.GCPConfig.MaxIssuerPathLen == 93
		},
		label: "GCP max issuer path len check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--disable-secret-creation"},
		checker: func(c *config.Config) bool {
			return c.DisableSecretCreation
		},
		label: "Disable secret creation flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--validity-days", "100"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.ValidityDays == 100
		},
		label: "Validity days flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--key-length", "20"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.KeyLength == 20
		},
		label: "Key length flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--validity-days", "100"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.ValidityDays == 100
		},
		label: "Validity days flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--common-name", "dummy-name"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.CertRequest.Subject.CommonName == "dummy-name"
		},
		label: "Common Name flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--country", "us"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.CertRequest.Subject.Country[0] == "us"
		},
		label: "Country flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--province", "dummy-province"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.CertRequest.Subject.Province[0] == "dummy-province"
		},
		label: "Province flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--organization", "dummy-org"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.CertRequest.Subject.Organization[0] == "dummy-org"
		},
		label: "Validity days flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--organizational-unit", "unit"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.CertRequest.Subject.OrganizationalUnit[0] == "unit"
		},
		label: "Organizational Unit flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--email", "some@something.oo"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.CertRequest.EmailAddresses[0] == "some@something.oo"
		},
		label: "Email flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--secret-file-path", "dummy/path"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.SecretFilePath == "dummy/path"
		},
		label: "Secret file path flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--override-existing-ca-cert-secret"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.OverrideExistingCACertSecret
		},
		label: "OverrideExistingCACertSecret flag check",
	}, {
		arguments: []string{"--provider", "aws", "--signing-ca", "dummy", "--istio-ca-namespace", "dummy-namespace"},
		checker: func(c *config.Config) bool {
			return c.CertParameters.IstioNamespace == "dummy-namespace"
		},
		label: "Istio CA Namespaces flag check",
	}}

	flags := pflag.NewFlagSet("Gen CA flags", pflag.ContinueOnError)
	genCAProviderParameters(flags)
	genCAx509CertRequestParameters(flags)

	for _, c := range cs {
		t.Run(c.label, func(t *testing.T) {
			require.NoError(t, flags.Parse(c.arguments))

			// ideally there shouldn't be any error
			// if there is one, something is wrong with the tests
			// so explicit error checks aren't required
			cfg, err := genCAFetchParameters(flags)
			require.NoError(t, err)
			require.Equal(t, true, c.checker(cfg))
		})
	}
}
