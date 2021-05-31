---
title: "getmesh fetch"
url: /getmesh-cli/reference/getmesh_fetch/
---

Fetch istioctl of the specified version, flavor and flavor-version available in "getmesh list" command

```
getmesh fetch [flags]
```

#### Examples

```
# Fetch the latest "tetrate flavored" istioctl of version=1.8
$ getmesh fetch --version 1.8

# Fetch the latest istioctl in version=1.7 and flavor=tetratefips
$ getmesh fetch --version 1.7 --flavor tetratefips

# Fetch the latest istioctl of version=1.7, flavor=tetrate and flavor-version=0
$ getmesh fetch --version 1.7 --flavor tetrate --flavor-version 0

# Fetch the istioctl of version=1.7.4 flavor=tetrate flavor-version=0
$ getmesh fetch --version 1.7.4 --flavor tetrate --flavor-version 0

# Fetch the istioctl of version=1.7.4 flavor=tetrate flavor-version=0 using name
$ getmesh fetch --name 1.7.4-tetrate-v0

# Fetch the latest istioctl of version=1.7.4 and flavor=tetratefips
$ getmesh fetch --version 1.7.4 --flavor tetratefips

# Fetch the latest "tetrate flavored" istioctl of version=1.7.4
$ getmesh fetch --version 1.7.4

# Fetch the istioctl of version=1.8.3 flavor=istio flavor-version=0
$ getmesh fetch --version 1.8.3 --flavor istio



# Fetch the latest "tetrate flavored" istioctl
$ getmesh fetch

As you can see the above examples:
- If --flavor-versions is not given, it defaults to the latest flavor version in the list
	If the value does not have patch version, "1.7" or "1.8" for example, then we fallback to the latest patch version in that minor version. 
- If --flavor is not given, it defaults to "tetrate" flavor.
- If --versions is not given, it defaults to the latest version of "tetrate" flavor.


For more information, please refer to "getmesh list --help" command.

```

#### Options

```
      --name string          Name of distribution, e.g. 1.9.0-istio-v0
      --version string       Version of istioctl e.g. "--version 1.7.4". When --name flag is set, this will not be used.
      --flavor string        Flavor of istioctl, e.g. "--flavor tetrate" or --flavor tetratefips" or --flavor istio". When --name flag is set, this will not be used.
      --flavor-version int   Version of the flavor, e.g. "--version 1". When --name flag is set, this will not be used. (default -1)
  -h, --help                 help for fetch
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getmesh](/getmesh-cli/reference/getmesh/)	 - getmesh is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

