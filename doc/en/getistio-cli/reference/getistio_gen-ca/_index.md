---
title: "getistio gen-ca"
url: /getistio-cli/reference/getistio_gen-ca/
---

Generates intermediate CA from different managed services such as AWS ACMPCA, GCP CAS

```
getistio gen-ca [flags]
```

#### Options

```
      --cas-ca-name string                CAS CA Name string
      --common-name string                Common name for x509 Cert request
      --config-file string                path to config file
      --country stringArray               Country names for x509 Cert request
      --disable-secret-creation           file only, doesn't create secret
      --email stringArray                 Emails for x509 Cert request
      --force                             force flag just deletes the existing secret and creates a new one
  -h, --help                              help for gen-ca
      --istio-ca-namespace cacerts        Namespace refered for creating the cacerts secrets in
      --key-length int                    length of generated key in bits for CA
      --locality stringArray              Locality names for x509 Cert request
      --max-issuer-path-len int32         CAS CA Max Issuer Path Length
      --organization stringArray          Organization names for x509 Cert request
      --organizational-unit stringArray   OrganizationalUnit names for x509 Cert request
  -p, --provider string                   name of the provider to be used, i.e aws, gcp
      --province stringArray              Province names for x509 Cert request
      --secret-file-path string           secret-file-path flag creates the secret YAML file
      --signing-algorithm string          Signing Algorithm to be used for issuing Cert using CSR for AWS
      --signing-ca string                 signing CA ARN string
      --template-arn string               Template ARN used to be used for issuing Cert using CSR
      --validity-days int                 valid dates for subordinate CA
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getistio](/getistio-cli/reference/getistio/)	 - GetIstio is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

