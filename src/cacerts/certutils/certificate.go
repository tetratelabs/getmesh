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

package certutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	privateca2 "google.golang.org/genproto/googleapis/cloud/security/privateca/v1beta1"
)

// CreateCSR creates a Cert Signing Request, and returns
// CSR, and sub-ordinate CA Key.
func CreateCSR(x509CertRequest x509.CertificateRequest, keyLenBits int) ([]byte, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keyLenBits)
	if err != nil {
		return nil, "", fmt.Errorf("unable to create private key: %w", err)
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("unable to marshal private key: %w", err)
	}

	keyBytesEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	subordinateCaKey := string(keyBytesEncoded)

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &x509CertRequest, privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("unable to create certificate request: %w", err)
	}

	csrBytesEncoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	return csrBytesEncoded, subordinateCaKey, nil
}

// CreateKeyPair creates private and public key
func CreateKeyPair(keyLenBits int) ([]byte, []byte, error) {
	privateKeyBytes, err := rsa.GenerateKey(rand.Reader, keyLenBits)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create private key: %w", err)
	}
	privateKey, err := x509.MarshalPKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal private key: %w", err)
	}

	privateKeyBytesEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})

	publicKey, err := x509.MarshalPKIXPublicKey(&privateKeyBytes.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal public key: %w", err)
	}
	publicKeyBytesEncoded := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKey})

	return privateKeyBytesEncoded, publicKeyBytesEncoded, nil
}

// CreateGCPRootCertificateChain creates GCP Root Certificate Chain.
func CreateGCPRootCertificateChain(certificate *privateca2.Certificate) (rootCertChain string) {
	rootCertChain = fmt.Sprintf("%s\n", certificate.PemCertificate)
	for _, s := range certificate.PemCertificateChain {
		rootCertChain += s + "\n"
	}
	return rootCertChain
}
