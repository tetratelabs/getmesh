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

package config

import (
	"testing"

	"github.com/tetratelabs/getmesh/internal/test"

	"github.com/stretchr/testify/require"
)

func TestNewConfigDefault(t *testing.T) {
	file := test.TempFile(t, "", "config")

	setConfig := `providerName: "aws"
providerConfig:
  aws:
    signingCAArn: "DUMMYARN"
    templateArn: "DUMMY-TEMPLATE_ARN"
certificateParameters:
  secretOptions:
    istioCANamespace: "dummy-istio-ns"`

	_, err := file.WriteString(setConfig)
	require.NoError(t, err)

	getConfig, err := NewConfig(file.Name())
	require.NoError(t, err)
	err = getConfig.ValidationsForConfig()
	require.NoError(t, err)
	getConfig.SetDefaultValues()
	require.NoError(t, err)

	// check modified values
	require.Equal(t, "DUMMYARN", getConfig.ProviderConfig.AWSConfig.SigningCA)
	require.Equal(t, "DUMMY-TEMPLATE_ARN", getConfig.ProviderConfig.AWSConfig.TemplateARN)
	require.Equal(t, "SHA256WITHRSA", getConfig.ProviderConfig.AWSConfig.SigningAlgorithm)
	require.Equal(t, "dummy-istio-ns", getConfig.CertParameters.SecretOptions.IstioNamespace)

	// check default values
	require.Equal(t, int64(3650), getConfig.CertParameters.ValidityDays)
	require.Equal(t, 2048, getConfig.CertParameters.KeyLength)
	require.Equal(t, "Istio CA", getConfig.CertParameters.CertRequest.Subject.CommonName)
	require.Equal(t, "California", getConfig.CertParameters.CertRequest.Subject.Province[0])
	require.Equal(t, "Sunnyvale", getConfig.CertParameters.CertRequest.Subject.Locality[0])
	require.Equal(t, "Istio", getConfig.CertParameters.CertRequest.Subject.Organization[0])
	require.Equal(t, "US", getConfig.CertParameters.CertRequest.Subject.Country[0])
	require.Equal(t, "ca.istio.io", getConfig.CertParameters.CertRequest.DNSNames[0])

}

func TestNewConfigWithoutProvider(t *testing.T) {
	file := test.TempFile(t, "", "config")

	setConfig := `providerName: "aws"
certificateParameters:
  secretOptions:
    istioCANamespace: "dummy-istio-ns"`

	_, err := file.WriteString(setConfig)
	require.NoError(t, err)

	getConfig, err := NewConfig(file.Name())
	require.NoError(t, err)
	getConfig.SetDefaultValues()

	err = getConfig.ValidationsForConfig()
	require.Error(t, err)
}

func TestNewConfigWithoutProviderRegion(t *testing.T) {
	file := test.TempFile(t, "", "config")

	setConfig := `providerName: "aws"
providerConfig:
  aws:
    signingCAArn: "DUMMYARN"
    templateArn: "DUMMY-TEMPLATE_ARN"
certificateParameters:
  secretOptions:
    istioCANamespace: "dummy-istio-ns"`

	_, err := file.WriteString(setConfig)
	require.NoError(t, err)

	getConfig, err := NewConfig(file.Name())
	require.NoError(t, err)
	getConfig.SetDefaultValues()

	err = getConfig.ValidationsForConfig()
	require.NoError(t, err)
}

func TestNewConfigWithoutProviderTemplateARN(t *testing.T) {
	file := test.TempFile(t, "", "config")

	setConfig := `providerName: "aws"
providerConfig:
  aws:
    signingCAArn: "DUMMYARN"
certificateParameters:
  secretOptions:
    istioCANamespace: "dumy-istio-ns"`

	_, err := file.WriteString(setConfig)
	require.NoError(t, err)

	getConfig, err := NewConfig(file.Name())
	require.NoError(t, err)
	getConfig.SetDefaultValues()

	err = getConfig.ValidationsForConfig()
	require.NoError(t, err)
}

func TestNewConfigWithoutProviderSigningCA(t *testing.T) {
	file := test.TempFile(t, "", "config")

	setConfig := `providerName: "aws"
providerConfig:
  aws:
    templateArn: "DUMMY-TEMPLATE_ARN"
certificateParameters:
  secretOptions:
    istioCANamespace: "dumy-istio-ns"`

	_, err := file.WriteString(setConfig)
	require.NoError(t, err)

	getConfig, err := NewConfig(file.Name())
	require.NoError(t, err)
	getConfig.SetDefaultValues()

	err = getConfig.ValidationsForConfig()
	require.Error(t, err)
}

func TestNewConfigNoPath(t *testing.T) {
	_, err := NewConfig("")
	require.Error(t, err)
}
