---
title: "getistio default-hub"
url: /getistio-cli/reference/getistio_default-hub/
---

Set or Show the default hub (root for Istio docker image paths) passed to "getistio istioctl install" via "--set hub="  e.g. docker.io/istio

```
getistio default-hub [flags]
```

#### Examples

```
# Set the default hub to docker.io/istio
$ getistio default-hub --set docker.io/istio

# Show the current default hub
$ getistio default-hub --show

```

#### Options

```
  -h, --help         help for default-hub
      --set string   pass the URL of hub, e.g. --set gcr.io/istio-testing
      --show         set to show the current default hub value
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getistio](/getistio-cli/reference/getistio/)	 - GetIstio is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

