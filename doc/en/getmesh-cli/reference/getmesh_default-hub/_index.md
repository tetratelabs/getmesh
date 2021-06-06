---
title: "getmesh default-hub"
url: /getmesh-cli/reference/getmesh_default-hub/
---

Set or Show the default hub (root for Istio docker image paths) passed to "getmesh istioctl install" via "--set hub="  e.g. docker.io/istio

```
getmesh default-hub [flags]
```

#### Examples

```
# Set the default hub to docker.io/istio
$ getmesh default-hub --set docker.io/istio

# Show the configured default hub
$ getmesh default-hub --show

# Remove the configured default hub
$ getmesh default-hub --remove

```

#### Options

```
  -h, --help         help for default-hub
      --remove       remove the configured default hub
      --set string   pass the location of hub, e.g. --set gcr.io/istio-testing
      --show         set to show the current default hub value
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getmesh](/getmesh-cli/reference/getmesh/)	 - getmesh is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

