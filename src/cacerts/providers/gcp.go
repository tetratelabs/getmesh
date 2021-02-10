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
	"fmt"
	"strings"

	privateca "cloud.google.com/go/security/privateca/apiv1beta1"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/google/uuid"
	privateca2 "google.golang.org/genproto/googleapis/cloud/security/privateca/v1beta1"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/tetratelabs/getistio/src/cacerts/certutils"
	"github.com/tetratelabs/getistio/src/cacerts/k8s"
	"github.com/tetratelabs/getistio/src/cacerts/providers/models"
)

// ProviderGCP contains credentials for establishing connection to the GCP,
type ProviderGCP struct {
	CASCAName        string `yaml:"casCAName"`
	MaxIssuerPathLen int32  `yaml:"maxIssuerPathLen"`
	*privateca.CertificateAuthorityClient
}

var _ ProviderInterface = &ProviderGCP{}

// IssueCA issues new Intermediate Certificate Authority, and provides a secret.
func (p *ProviderGCP) IssueCA(ctx context.Context, opts models.IssueCAOptions) (*k8s.IstioSecretDetails, error) {
	if err := p.createGCPClient(); err != nil {
		return nil, fmt.Errorf("unable to create GCP client :%w", err)
	}

	privateKey, publicKey, err := certutils.CreateKeyPair(opts.KeyLength)
	if err != nil {
		return nil, fmt.Errorf("unable to create key pair: %w", err)
	}

	certificateID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create certificate ID: %w", err)
	}

	requestID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create request ID: %w", err)
	}

	req := &privateca2.CreateCertificateRequest{
		CertificateId: certificateID.String(),
		RequestId:     requestID.String(),
		Parent:        p.CASCAName,
		Certificate: &privateca2.Certificate{
			Lifetime: &duration.Duration{
				Seconds: 60 * 60 * 24 * 365.25 * (opts.ValidityDays),
			},
			// configurable parameters
			CertificateConfig: &privateca2.Certificate_Config{
				Config: &privateca2.CertificateConfig{
					SubjectConfig: getSubjectConfig(opts.CAOptions),
					// non-configurable parameters
					ReusableConfig: p.getReusableConfig(),
					PublicKey: &privateca2.PublicKey{
						Type: privateca2.PublicKey_PEM_RSA_KEY,
						Key:  publicKey,
					},
				},
			},
		},
	}
	certificate, err := p.CertificateAuthorityClient.CreateCertificate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create GCP certificate: %w", err)
	}

	//Get the current CA's certificate
	request := privateca2.GetCertificateAuthorityRequest{Name: p.CASCAName}
	certificateAuthority, err := p.CertificateAuthorityClient.GetCertificateAuthority(ctx, &request)
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP Root CA certificate: %w", err)
	}

	casCaCertificate := certificateAuthority.PemCaCertificates[0]

	rootCertChain := certutils.CreateGCPRootCertificateChain(certificate)

	return k8s.NewBuilder().
		AddCACert(certificate.PemCertificate).
		AddCAKey(string(privateKey)).
		AddCertChain(rootCertChain).
		AddRootCertPem(casCaCertificate).
		Build()

}

func (p *ProviderGCP) createGCPClient() error {
	ctx := context.Background()

	client, err := privateca.NewCertificateAuthorityClient(ctx)
	if err != nil {
		return fmt.Errorf("unable to create GCP CAS client: %w", err)
	}
	p.CertificateAuthorityClient = client
	return nil
}

func (p *ProviderGCP) getReusableConfig() *privateca2.ReusableConfigWrapper {
	return &privateca2.ReusableConfigWrapper{
		ConfigValues: &privateca2.ReusableConfigWrapper_ReusableConfigValues{
			ReusableConfigValues: &privateca2.ReusableConfigValues{
				KeyUsage: &privateca2.KeyUsage{
					BaseKeyUsage: &privateca2.KeyUsage_KeyUsageOptions{
						CertSign:         true,
						DigitalSignature: true,
						CrlSign:          true,
						KeyEncipherment:  true,
						KeyAgreement:     true,
					},
					ExtendedKeyUsage: &privateca2.KeyUsage_ExtendedKeyUsageOptions{
						ServerAuth:      true,
						ClientAuth:      true,
						CodeSigning:     true,
						EmailProtection: true,
					},
				},
				CaOptions: &privateca2.ReusableConfigValues_CaOptions{
					IsCa: &wrapperspb.BoolValue{
						Value: true,
					},
					MaxIssuerPathLength: &wrapperspb.Int32Value{
						Value: p.MaxIssuerPathLen,
					},
				},
			},
		},
	}
}

func getSubjectConfig(caOptions models.CAOptions) *privateca2.CertificateConfig_SubjectConfig {
	return &privateca2.CertificateConfig_SubjectConfig{
		Subject: &privateca2.Subject{
			CountryCode:        strings.Join(caOptions.CertRequest.Subject.Country, ","),
			Organization:       strings.Join(caOptions.CertRequest.Subject.Organization, ","),
			OrganizationalUnit: strings.Join(caOptions.CertRequest.Subject.OrganizationalUnit, ","),
			Locality:           strings.Join(caOptions.CertRequest.Subject.Locality, ","),
			Province:           strings.Join(caOptions.CertRequest.Subject.Province, ","),
			StreetAddress:      strings.Join(caOptions.CertRequest.Subject.StreetAddress, ","),
			PostalCode:         strings.Join(caOptions.CertRequest.Subject.PostalCode, ","),
		},
		CommonName: caOptions.CertRequest.Subject.CommonName,
		SubjectAltName: &privateca2.SubjectAltNames{
			DnsNames:       caOptions.CertRequest.DNSNames,
			EmailAddresses: caOptions.CertRequest.EmailAddresses,
		},
	}
}
