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
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/acmpca/acmpcaiface"

	"github.com/tetratelabs/getmesh/internal/cacerts/certutils"
	"github.com/tetratelabs/getmesh/internal/cacerts/k8s"
	"github.com/tetratelabs/getmesh/internal/cacerts/providers/models"
)

// ProviderAWS contains credentials for establishing connection to the AWS,
type ProviderAWS struct {
	SigningCA   string `yaml:"signingCAArn"`
	TemplateARN string `yaml:"templateArn"`
	// SigningAlgorithm is the signing algorithm to be used by AWS Issue Certificate.
	SigningAlgorithm      string `yaml:"signingAlgorithm"`
	acmpcaiface.ACMPCAAPI `yaml:"-"`
}

var _ ProviderInterface = &ProviderAWS{}

// IssueCA issues new Intermediate CA, and provides a secret.
func (p *ProviderAWS) IssueCA(ctx context.Context, opts models.IssueCAOptions) (*k8s.IstioSecretDetails, error) {
	if err := p.createAWSClient(); err != nil {
		return nil, fmt.Errorf("unable to initialize AWS Client: %w", err)
	}

	csrBytesEncoded, subordinateCaKey, err := certutils.CreateCSR(opts.CertRequest, opts.KeyLength)
	if err != nil {
		return nil, fmt.Errorf("unable to create certificate signing request: %w", err)
	}

	rootCert, err := p.ACMPCAAPI.GetCertificateAuthorityCertificate(
		&acmpca.GetCertificateAuthorityCertificateInput{CertificateAuthorityArn: &p.SigningCA})
	if err != nil {
		return nil, fmt.Errorf("unable to get CACertificateRequest: %w", err)
	}

	subordinateCertArn, err := p.issueCert(ctx, models.IssueCertOptions{
		CertRequest: opts.CertRequest,
		CSR:         csrBytesEncoded,
		// TODO(@rahulchheda): allow to provide year, and month too.
		ValidityType:  "DAYS",
		ValidityValue: opts.ValidityDays,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to issue certificate: %w", err)
	}

	subordinateCert, err := p.getCert(ctx, models.GetCertOptions{
		CertNameIdentifier: *subordinateCertArn,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get certificate: %w", err)
	}

	certChain := *subordinateCert + "\n" + *rootCert.Certificate + "\n"
	var rootCertPem string
	if rootCert.CertificateChain != nil {
		certs := strings.SplitAfter(*rootCert.CertificateChain, "-----END CERTIFICATE-----")
		if len(certs) >= 2 {
			rootCertPem = certs[len(certs)-2]
			for i := 0; i < len(certs); i++ {
				certChain = certChain + certs[i] + "\n"
			}
		}
	} else {
		rootCertPem = *rootCert.Certificate
	}

	return k8s.NewBuilder().
		AddCACert(*subordinateCert).
		AddCAKey(subordinateCaKey).
		AddCertChain(certChain).
		AddRootCertPem(rootCertPem).
		Build()
}

// GetCert retrives a ACM Certificate from AWS.
func (p *ProviderAWS) getCert(ctx context.Context, opts models.GetCertOptions) (*string, error) {
	in := acmpca.GetCertificateInput{CertificateArn: &opts.CertNameIdentifier, CertificateAuthorityArn: &p.SigningCA}
	err := p.ACMPCAAPI.WaitUntilCertificateIssued(&in)
	if err != nil {
		return nil, fmt.Errorf("unable to get certificate: %w", err)
	}
	out, err := p.ACMPCAAPI.GetCertificate(&in)
	if err != nil {
		return nil, fmt.Errorf("unable to get certificate: %w", err)
	}
	return out.Certificate, nil
}

// IssueCert issues an ACM Certificate.
func (p *ProviderAWS) issueCert(ctx context.Context, opts models.IssueCertOptions) (*string, error) {
	t := fmt.Sprint(rand.Int())
	ici := acmpca.IssueCertificateInput{
		CertificateAuthorityArn: &p.SigningCA,
		SigningAlgorithm:        &p.SigningAlgorithm,
		IdempotencyToken:        &t,
		Validity:                &acmpca.Validity{Type: &opts.ValidityType, Value: &opts.ValidityValue},
		Csr:                     opts.CSR,
		TemplateArn:             &p.TemplateARN,
	}

	o, err := p.ACMPCAAPI.IssueCertificate(&ici)
	if err != nil {
		return nil, fmt.Errorf("unable to issue certificate: %w", err)
	}
	return o.CertificateArn, nil
}

func (p *ProviderAWS) createAWSClient() error {
	if p.ACMPCAAPI != nil {
		return nil
	}

	mySession := session.Must(session.NewSession())

	region, err := p.getRegion()
	if err != nil {
		return fmt.Errorf("unable to get AWS Region: %w", err)
	}

	p.ACMPCAAPI = acmpca.New(mySession, aws.NewConfig().WithRegion(region))
	return nil
}

func (p *ProviderAWS) getRegion() (string, error) {
	splitString := strings.SplitN(p.SigningCA, ":", 5)
	if len(splitString) < 5 {
		return "", errors.New("unable to find region for AWS Signing CA ARN")
	}

	return splitString[len(splitString)-2], nil
}
