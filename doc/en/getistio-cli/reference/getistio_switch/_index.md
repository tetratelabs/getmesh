---
title: "getistio switch"
url: /getistio-cli/reference/getistio_switch/
---

Switch the active istioctl to a specified version

```
getistio switch <istio version | flavor | flavor-version>|<istio version full name> [flags]
```

#### Examples

```
# switch the active istioctl version to version=1.7.4, flavor=tetrate and flavor-version=1
$ getistio switch --version 1.7.4 --flavor tetrate --flavor-version=1, 
$ getistio switch --name 1.7.4-tetrate-v1
switch also supports to change only one version|flavor|flavorVersion flag and follow the rest settings in active version
```

#### Options

```
      --name string          Name of istioctl, e.g. 1.9.0-istio-v0
      --version string       Version of istioctl, e.g. 1.7.4
      --flavor string        Flavor of istioctl, e.g. "tetrate" or "tetratefips" or "istio"
      --flavor-version int   Version of the flavor, e.g. 1 (default -1)
  -h, --help                 help for switch
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getistio](/getistio-cli/reference/getistio/)	 - GetIstio is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

