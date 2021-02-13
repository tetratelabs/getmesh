---
title: "getistio fetch"
url: /getistio-cli/reference/getistio_fetch/
---

Fetch istioctl of the specified version, flavor and flavor-version available in "getistio list" command

```
getistio fetch [flags]
```

#### Examples

```
# Fetch the latest "tetrate flavored" istioctl of version=1.8
$ getistio fetch --version 1.8

# Fetch the latest istioctl in version=1.7 and flavor=tetratefips
$ getistio fetch --version 1.7 --flavor tetratefips

# Fetch the latest istioctl of version=1.7, flavor=tetrate and flavor-version=0
$ getistio fetch --version 1.7 --flavor tetrate --flavor-version 0

# Fetch the istioctl of version=1.7.4 flavor=tetrate flavor-version=0
$ getistio fetch --version 1.7.4 --flavor tetrate --flavor-version 0

# Fetch the latest istioctl of version=1.7.4 and flavor=tetratefips
$ getistio fetch --version 1.7.4 --flavor tetratefips

# Fetch the latest "tetrate flavored" istioctl of version=1.7.4
$ getistio fetch --version 1.7.4



# Fetch the latest "tetrate flavored" istioctl
$ getistio fetch

As you can see the above examples:
- If --flavor-versions is not given, it defaults to the latest flavor version in the list
	If the value does not have patch version, "1.7" or "1.8" for example, then we fallback to the latest patch version in that minor version. 
- If --flavor is not given, it defaults to "tetrate" flavor.
- If --versions is not given, it defaults to the latest version of "tetrate" flavor.


For more information, please refer to "getistio list --help" command.

```

#### Options

```
      --version string       Version of istioctl e.g. "--version 1.7.4"
      --flavor string        Flavor of istioctl, e.g. "--flavor tetrate" or --flavor tetratefips"
      --flavor-version int   Version of the flavor, e.g. "--version 1" (default -1)
  -h, --help                 help for fetch
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getistio](/getistio-cli/reference/getistio/)	 - GetIstio is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

