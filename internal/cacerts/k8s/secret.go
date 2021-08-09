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

package k8s

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tetratelabs/getmesh/internal/cacerts/providers/models"
	"github.com/tetratelabs/getmesh/internal/util"
	"github.com/tetratelabs/getmesh/internal/util/logger"
)

// IstioSecretName is the name used to create the Secret in Istio CA Namespace.
const IstioSecretName = "cacerts"

// IstioSecretDetails is a structure used to contain
// all details about Kubernetes Secret Creation.
type IstioSecretDetails struct {
	secretDetails map[string][]byte
}

// Builder encapsulates IstioSecretDetails
// to add data into it.
type Builder struct {
	caSecretOptions *IstioSecretDetails
	errors          []error
}

// NewBuilder returns a new instance of Builder.
func NewBuilder() *Builder {
	return &Builder{
		caSecretOptions: &IstioSecretDetails{
			secretDetails: make(map[string][]byte),
		},
	}
}

// Build builds the IstioSecretDetails structure.
func (b *Builder) Build() (*IstioSecretDetails, error) {
	if len(b.errors) != 0 {
		return nil, util.HandleMultipleErrors(b.errors)
	}
	return b.buildIstioSecretDetails(), nil
}

func (b *Builder) buildIstioSecretDetails() *IstioSecretDetails {
	return &IstioSecretDetails{
		secretDetails: b.caSecretOptions.secretDetails,
	}
}

// AddCACert adds CACert Key in Secret's Data map.
func (b *Builder) AddCACert(caCert string) *Builder {
	if caCert == "" {
		b.errors = append(b.errors, errors.New("unable to add empty certificateauthority certificate key"))
		return b
	}
	b.caSecretOptions.secretDetails["ca-cert.pem"] = []byte(caCert)
	return b
}

// AddRootCertPem adds RootCertPem Key in Secret's Data map.
func (b *Builder) AddRootCertPem(rootCert string) *Builder {
	if rootCert == "" {
		b.errors = append(b.errors, errors.New("unable to add empty root certificate key"))
		return b
	}
	b.caSecretOptions.secretDetails["root-cert.pem"] = []byte(rootCert)
	return b
}

// AddCAKey adds CAKey Key in Secret's Data map.
func (b *Builder) AddCAKey(caKey string) *Builder {
	if caKey == "" {
		b.errors = append(b.errors, errors.New("unable to add empty ca key"))
		return b
	}
	b.caSecretOptions.secretDetails["ca-key.pem"] = []byte(caKey)
	return b
}

// AddCertChain adds CertChain Key in Secret's Data map.
func (b *Builder) AddCertChain(caKey string) *Builder {
	if caKey == "" {
		b.errors = append(b.errors, errors.New("unable to add empty cert chain key"))
		return b
	}
	b.caSecretOptions.secretDetails["cert-chain.pem"] = []byte(caKey)
	return b
}

// Create method creates the secret in the istio namespace provided.
func (secret *IstioSecretDetails) Create(namespace string, cli kubernetes.Interface) error {
	if _, err := cli.CoreV1().Secrets(namespace).Create(context.Background(), &v1.Secret{
		Data: secret.secretDetails,
		ObjectMeta: metav1.ObjectMeta{
			Name:      IstioSecretName,
			Namespace: namespace,
		},
	}, metav1.CreateOptions{}); err != nil {
		return err
	}

	logger.Infof("Kubernetes Secret created successfully with name: %s, in namespace: %s\n", IstioSecretName, namespace)
	return nil
}

// ForceCreate deletes the previous secret with similar name, and replace it with another.
func (secret *IstioSecretDetails) ForceCreate(namespace string, cli kubernetes.Interface) error {
	_ = cli.CoreV1().Secrets(namespace).Delete(context.Background(), IstioSecretName, metav1.DeleteOptions{})
	if _, err := cli.CoreV1().Secrets(namespace).Create(context.Background(), &v1.Secret{
		Data: secret.secretDetails,
		ObjectMeta: metav1.ObjectMeta{
			Name:      IstioSecretName,
			Namespace: namespace,
		},
	}, metav1.CreateOptions{}); err != nil {
		return err
	}

	logger.Infof("Kubernetes Secret created successfully with name: %s, in namespace: %s\n", IstioSecretName, namespace)
	return nil
}

// SaveToFile method prints out the secret to be verified and created later.
func (secret *IstioSecretDetails) SaveToFile(namespace string, secretFilePath string) error {
	toPrintSecret := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: cacerts
  namespace: %s
type: Opaque
data:
  ca-cert.pem: %s
  ca-key.pem: %s
  cert-chain.pem: %s
  root-cert.pem: %s`, namespace,
		base64.StdEncoding.EncodeToString(secret.secretDetails["ca-cert.pem"]),
		base64.StdEncoding.EncodeToString(secret.secretDetails["ca-key.pem"]),
		base64.StdEncoding.EncodeToString(secret.secretDetails["cert-chain.pem"]),
		base64.StdEncoding.EncodeToString(secret.secretDetails["root-cert.pem"]),
	)

	absPath, err := filepath.Abs(secretFilePath)
	if err != nil {
		return fmt.Errorf("unable to parse secret file path: %w", err)
	}

	file, err := os.Create(absPath)
	if err != nil {
		return fmt.Errorf("unable to create file in path: %s: %w", absPath, err)
	}

	if _, err := file.Write([]byte(toPrintSecret)); err != nil {
		return fmt.Errorf("unable to write file in path: %s: %w", absPath, err)
	}

	logger.Infof("Kubernetes Secret YAML created successfully in %s\n", secretFilePath)
	return nil
}

// CreateSecret creates secret at appropriate location according to the options provided.
func (secret *IstioSecretDetails) CreateSecret(secretOptions *models.SecretOptions) error {
	if secretOptions.SecretFilePath != "" {
		return secret.SaveToFile(secretOptions.IstioNamespace, secretOptions.SecretFilePath)
	}

	kubeCli, err := util.GetK8sClient()
	if err != nil {
		return fmt.Errorf("unable to create kube cli: %w", err)
	}

	if secretOptions.OverrideExistingCACertSecret {
		err := secret.ForceCreate(secretOptions.IstioNamespace, kubeCli)
		return err
	}

	return secret.Create(secretOptions.IstioNamespace, kubeCli)
}

// CreateSecretFile creates secret in getmesh Directory if we get any error while creating secret.
func (secret *IstioSecretDetails) CreateSecretFile() (string, error) {

	homeDir, err := util.GetmeshHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to find getmesh directory: %w", err)
	}

	if err = os.Mkdir(homeDir+"/secret", 0755); err != nil {
		if _, errDirExist := os.Stat(homeDir + "/secret"); os.IsNotExist(errDirExist) {
			return "", err
		}
	}

	tmpfile, err := ioutil.TempFile(homeDir+"/secret", "getmesh-*.yaml")
	if err != nil {
		return "", err
	}
	return tmpfile.Name(), secret.SaveToFile("istio-namespace", tmpfile.Name())
}
