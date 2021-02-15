# GetIstio [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

![Simplify Istio with GetIsito!](getIstio.png)

# Overview

An integration, and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio. The enterprises require ability to control Istio versioning, support multiple versions of istio, ability to easily move between the versions, integration with cloud providers certification systems and centralized config management and validation. 

The GetIsio CLI tool supports these enterprise level requirements via:

- enforcement of fetching certified versions of Istio and enables only compatible versions of Istio installation
- allows seamlessly switching between multiple istioctl versions
- includes FIPS compliant flavor
- delivers Istio configuration validations platform based by integrating validation libraries from multiple sources
- uses number of cloud provider certificate management systems to create Istio CA certs that are used for signing Service-Mesh managed workloads 
- also provides multiple additional integration points with cloud providers

Istio release schedule can be very aggressive for the enterprise life-cycle and change management practices - GetIstio addresses this concern by testing all Istio versions against different kubernetes distributions for functional integrity. The getistio supported versions of Istio are actively supported for security patches and other bug updates and have much longer support life than provided by upstream Istio.

Considering that some of Service-Mesh customers needs to support elevated security requirements - GetIstio addresses the compliance restriction by offering two flavors of Istio distribution:

- _tetrate_ tracks the upstream Istio and may have additional patches applied
- _tetratefips_ a FIPS compliant version of tetrate flavor

The above functionality is achieved via elegant transparent approach, where the existing setup and tools are fully leveraged to provide additional functionality and enterprise desired feature sets and controls:

- GetIstio connects to the kubernetes cluster pointed to by the default kubernetes config file. If KUBECONFIG environment variable is set, then takes precedence.
- Config validation is done against two targets:
cluster current config that might include multiple Istio configuration constructs
in addition GetIstio validates the manifest yaml files (that are not applied yet to the cluster)
- Creation of CA cert for Istio assumes the provider set up to issue intermediary CA cert is already done. This is optional and the default is self signed cert by Istio for workload certificates

# Get Started

GetIstio can be obtained by issuing the following command:

```sh
curl -sL https://tetrate.bintray.com/getistio/download.sh | bash
```

This, by default, downloads the latest version of GetIstio and certified Istio. To check if the download was successful, run the [version command](/doc/en/getistio-cli/reference/getistio_version/_index.md):

```sh
getistio version
```

or

```sh
getistio version --remote=false #only the client version details
```

An output of the form below suggests that GetIstio was installed successfully.
<pre>getistio version: 0.6.0
active istioctl: 1.8.2-tetrate-v0
</pre>

<br />

To see the list of commands available with GetIstio and its supported features, run the [help command](/doc/en/getistio-cli/reference/getistio_help/_index.md):

```sh
getistio --help
```

<br />

Once GetIstio is downloaded, it can be self-updated to the latest version by running the [update command](/doc/en/getistio-cli/reference/getistio_update/_index.md):

```sh
getistio update
```

<br />
While we recommend always using the latest GetIstio, if the user wants to download a different version of GetIstio for testing or other reasons, they could do so with the following command:

```sh
export GETISTIO_VERSION=<your_version> ##say 0.5.0 for example
curl -sL https://tetrate.bintray.com/getistio/download.sh | bash
```
This would overwrite the existing version of GetIstio to the one just downloaded.

For more information on GetIstio, please visit [getistio.io](https://getistio.io).

# Contributing

For developers interested in contributing to GetIstio, please follow the instruction in [CONTRIBUTING.md](CONTRIBUTING.md).
