---
title: "getmesh check-upgrade"
url: /getmesh-cli/reference/getmesh_check-upgrade/
---

Check if there are patches available in the current minor version, e.g. 1.7-tetrate: 1.7.4-tetrate-v1 -> 1.7.5-tetrate-v1

```
getmesh check-upgrade [flags]
```

#### Examples

```
# example output
$ getmesh check-upgrade
...
- Your data plane running in multiple minor versions: 1.7-tetrate, 1.8-tetrate
- Your control plane running in multiple minor versions: 1.6-tetrate, 1.8-tetrate
- The minor version 1.6-tetrate is not supported by Tetrate.io. We recommend you use the trusted minor versions in "getmesh list"
- There is the available patch for the minor version 1.7-tetrate. We recommend upgrading all 1.7-tetrate versions -> 1.7.4-tetrate-v1
- There is the available patch for the minor version 1.8-tetrate which includes **security upgrades**. We strongly recommend upgrading all 1.8-tetrate versions -> 1.8.1-tetrate-v1

In the above example, we call names in the form of x.y-${flavor} "minor version", where x.y is Istio's upstream minor and ${flavor} is the flavor of the distribution.
Please refer to 'getmesh fetch --help' or 'getmesh list --help' for more information.
```

#### Options

```
  -h, --help   help for check-upgrade
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getmesh](/getmesh-cli/reference/getmesh/)	 - getmesh is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

