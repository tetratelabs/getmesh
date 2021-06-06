---
title: "getmesh prune"
url: /getmesh-cli/reference/getmesh_prune/
---

Remove specific istioctl installed, or all, except the active one

```
getmesh prune [flags]
```

#### Examples

```
# remove all the installed
$ getmesh prune

# remove the specific distribution
$ getmesh prune --version 1.7.4 --flavor tetrate --flavor-version 0

```

#### Options

```
      --version string       Version of istioctl e.g. 1.7.4
      --flavor string        Flavor of istioctl, e.g. "tetrate" or "tetratefips" or "istio"
      --flavor-version int   Version of the flavor, e.g. 1 (default -1)
  -h, --help                 help for prune
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getmesh](/getmesh-cli/reference/getmesh/)	 - getmesh is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

