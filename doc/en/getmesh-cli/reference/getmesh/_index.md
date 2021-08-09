---
title: "getmesh"
url: /getmesh-cli/reference/getmesh/
---

getmesh is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

#### Options

```
  -h, --help                help for getmesh
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getmesh check-upgrade](/getmesh-cli/reference/getmesh_check-upgrade/)	 - Check if there are patches available in the current minor version
* [getmesh config-validate](/getmesh-cli/reference/getmesh_config-validate/)	 - Validate the current Istio configurations in your cluster
* [getmesh default-hub](/getmesh-cli/reference/getmesh_default-hub/)	 - Set or Show the default hub passed to "getmesh istioctl install" via "--set hub=" e.g. docker.io/istio
* [getmesh fetch](/getmesh-cli/reference/getmesh_fetch/)	 - Fetch istioctl of the specified version, flavor and flavor-version available in "getmesh list" command
* [getmesh gen-ca](/getmesh-cli/reference/getmesh_gen-ca/)	 - Generate intermediate CA
* [getmesh istioctl](/getmesh-cli/reference/getmesh_istioctl/)	 - Execute istioctl with given arguments
* [getmesh list](/getmesh-cli/reference/getmesh_list/)	 - List available Istio distributions built by Tetrate
* [getmesh prune](/getmesh-cli/reference/getmesh_prune/)	 - Remove specific istioctl installed, or all, except the active one
* [getmesh show](/getmesh-cli/reference/getmesh_show/)	 - Show fetched Istio versions
* [getmesh switch](/getmesh-cli/reference/getmesh_switch/)	 - Switch the active istioctl to a specified version
* [getmesh version](/getmesh-cli/reference/getmesh_version/)	 - Show the versions of getmesh cli, running Istiod, Envoy, and the active istioctl

