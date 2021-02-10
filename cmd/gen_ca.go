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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sys/unix"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tetratelabs/getistio/src/cacerts/k8s"
	"github.com/tetratelabs/getistio/src/cacerts/providers"
	"github.com/tetratelabs/getistio/src/cacerts/providers/config"
	"github.com/tetratelabs/getistio/src/cacerts/providers/models"
	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/util"
	"github.com/tetratelabs/getistio/src/util/logger"
)

func newGenCACmd() *cobra.Command {
	genCACmd := &cobra.Command{
		Use:   "gen-ca",
		Short: "Generate intermediate CA",
		Long:  `Generates intermediate CA from different managed services such as AWS ACMPCA, GCP CAS`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if getistio.GetActiveConfig().IstioDistribution == nil {
				return errors.New("please fetch Istioctl by `getistio fetch` beforehand")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := genCAFetchParameters(cmd.Flags())
			if err != nil {
				return err
			}

			var kubeCli *kubernetes.Clientset
			if !cfg.DisableSecretCreation {
				kubeCli, err = util.GetK8sClient()
				if err != nil {
					return fmt.Errorf("unable to create kube cli: %w", err)
				}
			}

			if err := genCAPreFlightChecks(cfg, kubeCli); err != nil {
				return fmt.Errorf("validation checks failed: %w", err)
			}

			var istioSecretDetails *k8s.IstioSecretDetails
			switch {

			case cfg.ProviderConfig.AWSConfig != nil:
				//TODO: (@rahulchheda) Init Context
				istioSecretDetails, err = cfg.ProviderConfig.AWSConfig.IssueCA(context.Background(), cfg.CertParameters)
				if err != nil {
					return fmt.Errorf("unable to issue CA, due to error: %v", err)
				}

			case cfg.ProviderConfig.GCPConfig != nil:
				istioSecretDetails, err = cfg.ProviderConfig.GCPConfig.IssueCA(context.Background(), cfg.CertParameters)
				if err != nil {
					return fmt.Errorf("unable to issue CA, due to error: %v", err)
				}

			default:
				return fmt.Errorf("%s provider yet to be implement", cfg.ProviderName)
			}

			secretFilePath, err := istioSecretDetails.CreateSecretFile()
			if err != nil {
				return fmt.Errorf("unable to create secret file: %v", err)
			}

			if cfg.DisableSecretCreation {
				logger.Infof("Please fire `kubectl apply -f %s`, to create a secret in Kubernetes Cluster.\n", secretFilePath)
				return nil
			}

			err = istioSecretDetails.CreateSecret(&cfg.CertParameters.SecretOptions)
			if err != nil {
				logger.Infof("As a fallback procedure, we have created a secret YAML at above location,"+
					" please fire `kubectl apply -f %s`, to create a secret in Kubernetes Cluster.\n", secretFilePath)
				return err
			}
			return nil
		},
	}

	genCAProviderParameters(genCACmd.Flags())
	genCAx509CertRequestParameters(genCACmd.Flags())

	return genCACmd
}

func genCAProviderParameters(flags *pflag.FlagSet) {
	flags.StringP("config-file", "", "", "path to config file")
	flags.BoolP("disable-secret-creation", "", false, "file only, doesn't create secret")

	// common to all providers
	flags.StringP("provider", "p", "", "name of the provider to be used, i.e aws, gcp")

	// specific to AWS
	flags.StringP("signing-ca", "", "", "signing CA ARN string")
	flags.StringP("template-arn", "", "", "Template ARN used to be used for issuing Cert using CSR")
	flags.StringP("signing-algorithm", "", "", "Signing Algorithm to be used for issuing Cert using CSR for AWS")

	//specific to GCP
	flags.StringP("cas-ca-name", "", "", "CAS CA Name string")
	flags.Int32P("max-issuer-path-len", "", 0, "CAS CA Max Issuer Path Length")

}

func genCAx509CertRequestParameters(flags *pflag.FlagSet) {
	flags.StringP("common-name", "", "", "Common name for x509 Cert request")
	flags.StringArrayP("country", "", []string{""}, "Country names for x509 Cert request")
	flags.StringArrayP("province", "", []string{""}, "Province names for x509 Cert request")
	flags.StringArrayP("locality", "", []string{""}, "Locality names for x509 Cert request")
	flags.StringArrayP("organization", "", []string{""}, "Organization names for x509 Cert request")
	flags.StringArrayP("organizational-unit", "", []string{""}, "OrganizationalUnit names for x509 Cert request")
	flags.StringArrayP("email", "", []string{""}, "Emails for x509 Cert request")
	flags.StringP("istio-ca-namespace", "", "", "Namespace refered for creating the `cacerts` secrets in")
	flags.StringP("secret-file-path", "", "", "secret-file-path flag creates the secret YAML file")
	flags.BoolP("force", "", false, "force flag just deletes the existing secret and creates a new one")
	flags.Int64P("validity-days", "", 0, "valid dates for subordinate CA")
	flags.IntP("key-length", "", 0, "length of generated key in bits for CA")

}

func genCAFetchParameters(flags *pflag.FlagSet) (*config.Config, error) {
	errorList := []error{fmt.Errorf("invalid parameters: ")}
	var configBuilder config.Config
	p, err := flags.GetString("config-file")
	if err == nil && p != "" {
		// this means config path path is available
		configFile, err := config.NewConfig(p)
		if err != nil {
			return nil, err
		}
		configBuilder = *configFile
	}

	// Override the values from ConfigFile, with the flags provided.

	// Required Flags
	provider, err := flags.GetString("provider")
	if err != nil {
		return nil, errors.New("unable to parse `--provider` flag")
	}
	if configBuilder.ProviderName == "" {
		configBuilder.ProviderName = provider
	}

	switch models.ProviderName(configBuilder.ProviderName) {
	case models.AWS:
		signingCA, err := flags.GetString("signing-ca")
		if err != nil {
			errorList = append(errorList, fmt.Errorf("unable to parse signing-ca flag: %w", err))
		}

		templateArn, err := flags.GetString("template-arn")
		if err != nil {
			errorList = append(errorList, fmt.Errorf("unable to parse template-arn flag: %w", err))
		}

		signingAlgorithm, err := flags.GetString("signing-algorithm")
		if err != nil {
			errorList = append(errorList, fmt.Errorf("unable to parse signing-algorithm flag: %w", err))
		}

		if configBuilder.ProviderConfig.AWSConfig == nil {
			configBuilder.ProviderConfig.AWSConfig = &providers.ProviderAWS{}
		}
		if signingCA != "" {
			configBuilder.ProviderConfig.AWSConfig.SigningCA = signingCA
		}

		if templateArn != "" {
			configBuilder.ProviderConfig.AWSConfig.TemplateARN = templateArn
		}

		if signingAlgorithm != "" {
			configBuilder.ProviderConfig.AWSConfig.SigningAlgorithm = signingAlgorithm
		}

	case models.GCP:

		casCAName, err := flags.GetString("cas-ca-name")
		if err != nil {
			errorList = append(errorList, fmt.Errorf("unable to parse `--cas-ca-name` flag: %w", err))
		}

		maxIssuerPathLen, err := flags.GetInt32("max-issuer-path-len")
		if err != nil {
			errorList = append(errorList, fmt.Errorf("unable to parse `--max-issuer-path-len` flag: %w", err))
		}

		if configBuilder.ProviderConfig.GCPConfig == nil {
			configBuilder.ProviderConfig.GCPConfig = &providers.ProviderGCP{}
		}

		if casCAName != "" {
			configBuilder.ProviderConfig.GCPConfig.CASCAName = casCAName
		}

		if maxIssuerPathLen != 0 {
			configBuilder.ProviderConfig.GCPConfig.MaxIssuerPathLen = maxIssuerPathLen
		}

	default:
		return nil, fmt.Errorf("`%s` provider yet to be implement", configBuilder.ProviderName)
	}

	// Optional Flags
	configBuilder.DisableSecretCreation, err = flags.GetBool("disable-secret-creation")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --disable-secret-creation: %w", err))
	}

	validityDays, err := flags.GetInt64("validity-days")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --validity-days: %w", err))
	}

	if validityDays != 0 {
		configBuilder.CertParameters.ValidityDays = validityDays
	}

	keyLen, err := flags.GetInt("key-length")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --key-length: %w", err))
	}

	if keyLen != 0 {
		configBuilder.CertParameters.KeyLength = keyLen
	}

	commonName, err := flags.GetString("common-name")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --common-name: %w", err))
	}

	if commonName != "" {
		configBuilder.CertParameters.CertRequest.Subject.CommonName = commonName
	}

	countries, err := flags.GetStringArray("country")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --country: %w", err))
	}

	if countries != nil {
		configBuilder.CertParameters.CertRequest.Subject.Country = countries
	}

	provinces, err := flags.GetStringArray("province")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --province: %w", err))
	}

	if provinces != nil {
		configBuilder.CertParameters.CertRequest.Subject.Province = provinces
	}

	localities, err := flags.GetStringArray("locality")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --locality: %w", err))
	}

	if localities != nil {
		configBuilder.CertParameters.CertRequest.Subject.Locality = localities
	}

	organizations, err := flags.GetStringArray("organization")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --organization: %w", err))
	}

	if organizations != nil {
		configBuilder.CertParameters.CertRequest.Subject.Organization = organizations
	}

	organizationalUnits, err := flags.GetStringArray("organizational-unit")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --organizational-unit: %w", err))
	}

	if organizationalUnits != nil {
		configBuilder.CertParameters.CertRequest.Subject.OrganizationalUnit = organizationalUnits
	}

	emails, err := flags.GetStringArray("email")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --email: %w", err))
	}

	if emails != nil {
		configBuilder.CertParameters.CertRequest.EmailAddresses = emails
	}

	secretFilePath, err := flags.GetString("secret-file-path")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --secret-file-path: %w", err))
	}

	if secretFilePath != "" {
		configBuilder.CertParameters.SecretFilePath = secretFilePath
	}

	force, err := flags.GetBool("force")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --force: %w", err))
	}

	if force {
		configBuilder.CertParameters.Force = force
	}

	istioNamespace, err := flags.GetString("istio-ca-namespace")
	if err != nil {
		errorList = append(errorList, fmt.Errorf("invalid --istio-ca-namespace: %w", err))
	}

	if istioNamespace != "" {
		configBuilder.CertParameters.IstioNamespace = istioNamespace
	}

	if len(errorList) > 1 {
		return nil, util.HandleMultipleErrors(errorList)
	}

	if err := configBuilder.ValidationsForConfig(); err != nil {
		return nil, err
	}

	configBuilder.SetDefaultValues()
	return &configBuilder, nil
}

func genCAPreFlightChecks(cfg *config.Config, kubeCli kubernetes.Interface) error {
	opts := cfg.CertParameters
	if opts.SecretFilePath != "" {
		return genCAValidateSecretFilePath(opts.SecretFilePath)
	}

	if !cfg.DisableSecretCreation {
		_, err := kubeCli.CoreV1().Namespaces().Get(context.Background(), opts.IstioNamespace, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to check if namespace %s exists: %w", opts.IstioNamespace, err)
		}

		_, err = kubeCli.CoreV1().Secrets(opts.IstioNamespace).Get(context.Background(), "cacerts", v1.GetOptions{})
		if err == nil && !opts.Force {
			return fmt.Errorf("`cacerts` secret already exist in %s namespace", opts.IstioNamespace)
		}

		if k8serrors.IsNotFound(err) {
			// validation successful because the secert doesn't exist.
			return nil
		}
		// some other error while getting the secret has come up.
		return err
	}
	return nil
}

func genCAValidateSecretFilePath(secretFilePath string) error {
	absPath, err := filepath.Abs(secretFilePath)
	if err != nil {
		return fmt.Errorf("unable to parse secret file path: %w", err)
	}
	if unix.Access(filepath.Dir(absPath), unix.W_OK) != nil {
		return fmt.Errorf("unable to write on secret file path: %w", err)
	}
	_, err = os.Stat(absPath)
	if err == nil {
		return fmt.Errorf("secret file path `%s` already exist, please change the file path before proceeding", absPath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}
