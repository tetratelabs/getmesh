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
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	ops "github.com/tetratelabs/getistio/src/cacerts/providers"
	"github.com/tetratelabs/getistio/src/cacerts/providers/models"
)

var (
	errNilGCPConfig         = errors.New("found nil GCP Config")
	errEmptyGCPCAName       = errors.New("found empty GCP CA Name")
	errNilAWSConfig         = errors.New("no AWS Information provided")
	errEmptyAWSSigningCAArn = errors.New("found empty AWS Signing CA ARN")
	errParsePath            = errors.New("unable to parse config file path")
	errOpenPath             = errors.New("unable to open config file path")
	errReadPath             = errors.New("unable to read config file contents")
	errUnmarshalConfig      = errors.New("unable to unmarshal config file contents")
)

// Config represent a structure used to accept request for `getistio`
// from a config file.
// Sample Config for AWS:
// providerName: "aws"
// disableSecretCreation: "true"
// providerConfig:
//   aws:
//     creds: "creds"
//     rootCAArn: "ROOT_CA_ARN"
//     region: "us-west-2"
//     templateArn: "TEMPLATE_ARN"
// certificateParameters:
//   secretOptions:
//     istioCANamespace: "istio-system"
//     secretFilePath: "/tmp/getistio/secret.yaml"
//   caOptions:
//     validityDays: 3650
//     keyLength: 2048
//     certSigningRequestParams:
//       subject:
//         commonname: "Istio CA"
//         country:
//           - "US"
//         locality:
//           - "Sunnyvale"
//         organization:
//           - "Istio"
type Config struct {
	// ProviderName is the name of the provider to be used.
	ProviderName          string `yaml:"providerName"`
	DisableSecretCreation bool   `yaml:"disableSecretCreation"`
	// ProviderConfig encapsulates all the configuration
	// needed to connect to the Provider.
	ProviderConfig struct {
		// AWSConfig contains all the AWS related configuration
		AWSConfig *ops.ProviderAWS `yaml:"aws,omitempty"`
		GCPConfig *ops.ProviderGCP `yaml:"gcp,omitempty"`
	} `yaml:"providerConfig"`

	// CertParameters contains all the Certificate related information.
	CertParameters models.IssueCAOptions `yaml:"certificateParameters,omitempty"`
}

var ExampleAWSInstance = Config{
	ProviderName:          "aws",
	DisableSecretCreation: false,
	ProviderConfig: struct {
		AWSConfig *ops.ProviderAWS `yaml:"aws,omitempty"`
		GCPConfig *ops.ProviderGCP `yaml:"gcp,omitempty"`
	}{AWSConfig: &ops.ProviderAWS{
		SigningCA:        "<your ACM PCA CA ARN>",
		TemplateARN:      "arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen0/V1",
		SigningAlgorithm: "SHA256WITHRSA",
	}},
	CertParameters: defaultCertParams,
}

var ExampleGCPInstance = Config{
	ProviderName:          "gcp",
	DisableSecretCreation: false,
	ProviderConfig: struct {
		AWSConfig *ops.ProviderAWS `yaml:"aws,omitempty"`
		GCPConfig *ops.ProviderGCP `yaml:"gcp,omitempty"`
	}{GCPConfig: &ops.ProviderGCP{
		CASCAName:        "projects/{project-id}/locations/{location}/certificateAuthorities/{YourCA}",
		MaxIssuerPathLen: 0,
	}},
	CertParameters: defaultCertParams,
}

// NewConfig returns new parsed config
// from the absolute path provided.
func NewConfig(path string) (*Config, error) {
	c := &Config{}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errParsePath, err)
	}

	configFile, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errOpenPath, err)
	}
	defer configFile.Close()

	b, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errReadPath, err)
	}

	if len(b) != 0 {
		if err := yaml.Unmarshal(b, c); err != nil {
			return nil, fmt.Errorf("%v: %w", errUnmarshalConfig, err)
		}
	}

	return c, nil
}

func (c *Config) ToYaml() (string, error) {
	str, err := yaml.Marshal(c)
	return string(str), err
}

// ValidationsForConfig validates the config before proceeding.
func (c *Config) ValidationsForConfig() error {
	if models.ProviderName(c.ProviderName) != models.AWS &&
		models.ProviderName(c.ProviderName) != models.GCP {
		return fmt.Errorf("unable to identity provider name: `%s`. Please try with lower case letters", c.ProviderName)
	}

	switch models.ProviderName(c.ProviderName) {
	case models.AWS:
		return validationForAWS(c.ProviderConfig.AWSConfig)

	case models.GCP:
		return validationForGCP(c.ProviderConfig.GCPConfig)

	default:
		return fmt.Errorf("provider %s yet to be implemented", c.ProviderName)
	}
}

var defaultCertParams = models.IssueCAOptions{
	SecretOptions: models.SecretOptions{
		IstioNamespace:               "istio-system",
		SecretFilePath:               "~/.getistio/secret/",
		OverrideExistingCACertSecret: false,
	},
	CAOptions: models.CAOptions{
		CertRequest: x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:   "Istio CA",
				Organization: []string{"Istio"},
				Country:      []string{"US"},
				Province:     []string{"California"},
				Locality:     []string{"Sunnyvale"},
			},
			DNSNames: []string{"ca.istio.io"},
		},
		ValidityDays: 3650,
		KeyLength:    2048,
	},
}

// SetDefaultValues sets the defaults in config struct as follows:
// certificateParameters:
//   secretOptions:
//     istioCANamespace: "istio-system"
//   caOptions:
//     validityDays: 3650
//     keyLength: 2048
//       subject:
//         commonname: "Istio CA"
//         country:
//           - "US"
//         locality:
//           - "Sunnyvale"
//         organization:
//           - "Istio"
func (c *Config) SetDefaultValues() {
	// Default Cert Parameters

	if c.CertParameters.ValidityDays == 0 {
		c.CertParameters.ValidityDays = defaultCertParams.ValidityDays
	}

	if c.CertParameters.KeyLength == 0 {
		c.CertParameters.KeyLength = defaultCertParams.KeyLength
	}

	if c.CertParameters.IstioNamespace == "" {
		c.CertParameters.IstioNamespace = defaultCertParams.IstioNamespace
	}

	// Default Cert Request
	if c.CertParameters.CertRequest.Subject.CommonName == "" {
		c.CertParameters.CertRequest.Subject.CommonName = defaultCertParams.CertRequest.Subject.CommonName
	}

	if c.CertParameters.CertRequest.Subject.Province == nil {
		c.CertParameters.CertRequest.Subject.Province = defaultCertParams.CertRequest.Subject.Province
	}

	if c.CertParameters.CertRequest.Subject.Locality == nil {
		c.CertParameters.CertRequest.Subject.Locality = defaultCertParams.CertRequest.Subject.Locality
	}

	if c.CertParameters.CertRequest.Subject.Organization == nil {
		c.CertParameters.CertRequest.Subject.Organization = defaultCertParams.CertRequest.Subject.Organization
	}

	if c.CertParameters.CertRequest.Subject.Country == nil {
		c.CertParameters.CertRequest.Subject.Country = defaultCertParams.CertRequest.Subject.Country
	}

	if c.CertParameters.CertRequest.DNSNames == nil {
		c.CertParameters.CertRequest.DNSNames = defaultCertParams.CertRequest.DNSNames
	}

}

func validationForAWS(awsConfig *ops.ProviderAWS) error {
	if awsConfig == nil {
		return errNilAWSConfig
	}

	if awsConfig.SigningCA == "" {
		return errEmptyAWSSigningCAArn
	}

	if awsConfig.TemplateARN == "" {
		awsConfig.TemplateARN = "arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen0/V1"
	}

	if awsConfig.SigningAlgorithm == "" {
		awsConfig.SigningAlgorithm = "SHA256WITHRSA"
	}

	return nil
}

func validationForGCP(gcpConfig *ops.ProviderGCP) error {
	// TODO(@rahulchheda): Fix these validations to seperate one's
	if gcpConfig == nil {
		return errNilGCPConfig
	}

	if gcpConfig.CASCAName == "" {
		return errEmptyGCPCAName
	}

	if gcpConfig.MaxIssuerPathLen == 0 {
		gcpConfig.MaxIssuerPathLen = 0
	}

	return nil
}
