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

package models

import "crypto/x509"

// ProviderName to specify the Name of the Provider to be used.
type ProviderName string

const (
	// GCP is yet not implemented yet.
	GCP ProviderName = "gcp"
	// AWS is identifier for using ACMPCA.
	AWS ProviderName = "aws"
)

// CAOptions are the customizable options available
// for Issuing Intermediate CA.
type CAOptions struct {
	// CertRequest encapsulates all the configurable parameters for creating
	// Certificate Request
	CertRequest x509.CertificateRequest `yaml:"certSigningRequestParams"`
	// ValidityDays represents the number of validity days before the CA expires.
	// This should not exceed the validity days of Root CA.
	ValidityDays int64 `yaml:"validityDays"`
	// KeyLength is the length(bits) of Key to be created.
	KeyLength int `yaml:"keyLength"`
}

// SecretOptions are the option available for secret creation.
type SecretOptions struct {
	// IstioNamespace is the namespace in which the `cacerts` secret
	// is created after a successful workflow.
	IstioNamespace string `yaml:"istioCANamespace"`
	// SecretFilePath is the file path used to store the Kubernetes Secret
	// to be applied afterwards.
	SecretFilePath string `yaml:"secretFilePath"`
	// Force flag when enabled forcefully deletes the `cacerts` secret
	// in istioNamespace, and creates a new one.
	Force bool `yaml:"force"`
}

// IssueCAOptions are the options available for CA creation/
type IssueCAOptions struct {
	// SecretOptions encapsulates all the Secret related
	// options used for issuing CA.
	SecretOptions `yaml:"secretOptions"`
	// CAOptions encapsulates all the CA related options
	// used for issuing CA.
	CAOptions `yaml:"caOptions"`
}

// IssueCertOptions are the options used to Issue new Cert.
type IssueCertOptions struct {
	// CertRequest encapsulates all the configurable parameters for creating
	// Certificate Request
	CertRequest x509.CertificateRequest
	// CSR represents Certificate Signing Request
	CSR []byte
	// ValidityType is set to `DAYS`
	ValidityType string
	// ValidityValue is the integer to be set as
	// the CA ValidityDays.
	ValidityValue int64
}

// GetCertOptions are the options used to get Cert.
type GetCertOptions struct {
	// CertNameIdentifier is a unique Identifier
	// for fetching Certificate.
	CertNameIdentifier string
}
