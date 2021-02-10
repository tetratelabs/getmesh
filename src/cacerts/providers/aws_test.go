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

package providers

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/acmpca/acmpcaiface"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/getistio/src/cacerts/providers/models"
	"gotest.tools/assert"
)

const (
	errMockGetCertificate             = "mockValidationErrorForGetCertificate"
	errMockWaitUntilCertificateIssued = "mockValidationErrorForWaitUntilCertificateIssued"
	errMockIssueCertificate           = "mockValidationErrorForIssueCertificate"
)

type MockACMPCACli struct {
	acmpcaiface.ACMPCAAPI
}

func (f MockACMPCACli) IssueCertificate(input *acmpca.IssueCertificateInput) (*acmpca.IssueCertificateOutput, error) {
	if *input.TemplateArn == "" {
		return nil, errors.New(errMockIssueCertificate)
	}
	mockCertificateArn := "mockCertificateArn"
	return &acmpca.IssueCertificateOutput{
		CertificateArn: &mockCertificateArn,
	}, nil
}

func (f MockACMPCACli) GetCertificate(input *acmpca.GetCertificateInput) (*acmpca.GetCertificateOutput, error) {
	if *input.CertificateArn == "" {
		return nil, errors.New(errMockGetCertificate)
	}
	mockCert := "mockCert"
	return &acmpca.GetCertificateOutput{
		Certificate: &mockCert,
	}, nil
}

func (f MockACMPCACli) WaitUntilCertificateIssued(input *acmpca.GetCertificateInput) error {
	if *input.CertificateAuthorityArn == "" {
		return errors.New(errMockWaitUntilCertificateIssued)
	}
	return nil
}

func (f MockACMPCACli) GetCertificateAuthorityCertificate(input *acmpca.GetCertificateAuthorityCertificateInput) (*acmpca.GetCertificateAuthorityCertificateOutput, error) {
	if *input.CertificateAuthorityArn == "" {
		return nil, errors.New(errMockGetCertificate)
	}
	mockCert := "mockCert"
	mockCertChain := "mockCertChain1-----END CERTIFICATE-----mockCertChain2-----END CERTIFICATE-----mockCertChain3-----END CERTIFICATE-----"
	return &acmpca.GetCertificateAuthorityCertificateOutput{
		Certificate:      &mockCert,
		CertificateChain: &mockCertChain,
	}, nil
}

func TestIssueCertificateAuthority(t *testing.T) {
	provider := ProviderAWS{
		ACMPCAAPI:        MockACMPCACli{},
		SigningCA:        "MockSigningCA",
		TemplateARN:      "MockTemplateARN",
		SigningAlgorithm: "SHA",
	}

	_, err := provider.IssueCA(context.TODO(), models.IssueCAOptions{
		CAOptions: models.CAOptions{
			CertRequest: x509.CertificateRequest{
				Subject: pkix.Name{
					CommonName:         "MK",
					Country:            []string{"MK"},
					Organization:       []string{"Mock"},
					OrganizationalUnit: []string{"Mock"},
				},
				EmailAddresses: []string{"Mock@Mockr.io"},
			},
			ValidityDays: 30,
			KeyLength:    2048,
		},
		SecretOptions: models.SecretOptions{
			IstioNamespace: "Mock",
		},
	})
	require.NoError(t, err)
}

func TestIssueCertificateAuthorityNegative(t *testing.T) {
	provider := ProviderAWS{
		ACMPCAAPI:        MockACMPCACli{},
		SigningCA:        "",
		TemplateARN:      "MockTemplateARN",
		SigningAlgorithm: "SHA",
	}

	_, err := provider.IssueCA(context.TODO(), models.IssueCAOptions{
		CAOptions: models.CAOptions{
			KeyLength: 2048,
			CertRequest: x509.CertificateRequest{
				Subject: pkix.Name{
					CommonName:         "MK",
					Country:            []string{"MK"},
					Organization:       []string{"Mock"},
					OrganizationalUnit: []string{"Mock"},
				},
				EmailAddresses: []string{"Mock@Mockr.io"},
			},
			ValidityDays: 30,
		},
		SecretOptions: models.SecretOptions{
			IstioNamespace: "Mock",
		},
	})
	assert.Equal(t, err.Error(), "unable to get CACertificateRequest: mockValidationErrorForGetCertificate")
}

func TestGetCertificate(t *testing.T) {
	provider := ProviderAWS{
		ACMPCAAPI:   MockACMPCACli{},
		SigningCA:   "MockSigningCA",
		TemplateARN: "MockTemplateARN",
	}

	_, err := provider.getCert(context.TODO(), models.GetCertOptions{
		CertNameIdentifier: "faceCAArn",
	})
	require.NoError(t, err)
}

func TestGetCertificateNegative(t *testing.T) {
	provider := ProviderAWS{
		ACMPCAAPI:   MockACMPCACli{},
		SigningCA:   "MockSigningCA",
		TemplateARN: "MockTemplateARN",
	}

	_, err := provider.getCert(context.TODO(), models.GetCertOptions{
		CertNameIdentifier: "",
	})
	assert.Equal(t, err.Error(), "unable to get certificate: "+errMockGetCertificate)
}

func TestIssueCertificate(t *testing.T) {
	provider := ProviderAWS{
		ACMPCAAPI:        MockACMPCACli{},
		SigningCA:        "MockSigningCA",
		TemplateARN:      "MockTemplateARN",
		SigningAlgorithm: "SHA",
	}

	_, err := provider.issueCert(context.TODO(), models.IssueCertOptions{
		CertRequest: x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:         "MK",
				Country:            []string{"MK"},
				Organization:       []string{"Mock"},
				OrganizationalUnit: []string{"Mock"},
			},
			EmailAddresses: []string{"Mock@Mockr.io"},
		},
		ValidityType:  "DAYS",
		ValidityValue: 30,
	})
	require.NoError(t, err)
}

func TestGetIssueCertificateNegative(t *testing.T) {
	provider := ProviderAWS{
		ACMPCAAPI:        MockACMPCACli{},
		SigningCA:        "MockSigningCA",
		SigningAlgorithm: "SHA",
	}

	_, err := provider.issueCert(context.TODO(), models.IssueCertOptions{
		CertRequest: x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:         "MK",
				Country:            []string{"MK"},
				Organization:       []string{"Mock"},
				OrganizationalUnit: []string{"Mock"},
			},
			EmailAddresses: []string{"Mock@Mockr.io"},
		},
		ValidityType:  "DAYS",
		ValidityValue: 30,
	})
	assert.Equal(t, err.Error(), "unable to issue certificate: "+errMockIssueCertificate)
}
