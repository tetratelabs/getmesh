---
title: "getmesh istioctl"
url: /getmesh-cli/reference/getmesh_istioctl/
---

Execute istioctl with given arguments where the version of istioctl is set by "getsitio fetch or switch"

```
getmesh istioctl <args...> [flags]
```

#### Examples

```
# install Istio with the default profile
getmesh istioctl install --set profile=default

# check versions of Istio data plane, control plane, and istioctl
getmesh istioctl version
```

#### Options

```
  -h, --help   help for istioctl
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getmesh](/getmesh-cli/reference/getmesh/)	 - getmesh is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

