---
title: "getistio prune"
url: /getistio-cli/reference/getistio_prune/
---

Remove specific istioctl installed, or all, except the active one

```
getistio prune [flags]
```

#### Examples

```
# remove all the installed
$ getistio prune

# remove the specific distribution
$ getistio prune --version 1.7.4 --flavor tetrate --flavor-version 0

```

#### Options

```
      --flavor string        Flavor of istioctl, e.g. "tetrate" or "tetratefips"
      --flavor-version int   Version of the flavor, e.g. 1 (default -1)
  -h, --help                 help for prune
      --version string       Version of istioctl e.g. 1.7.4
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getistio](/getistio-cli/reference/getistio/)	 - GetIstio is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

