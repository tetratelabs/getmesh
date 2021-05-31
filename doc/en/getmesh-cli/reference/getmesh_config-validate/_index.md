---
title: "getmesh config-validate"
url: /getmesh-cli/reference/getmesh_config-validate/
---

Validate the current Istio configurations in your cluster just like 'istioctl analyze'. Inspect all namespaces by default.
If the <file/directory> is specified, we analyze the effect of applying these yaml files against the current cluster.

```
getmesh config-validate <file/directory>... [flags]
```

#### Examples

```
# validating a local manifest against the current cluster
$ getmesh config-validate my-app.yaml another-app.yaml

# validating local manifests in a directory against the current cluster in a specific namespace
$ getmesh config-validate -n bookinfo my-manifest-dir/

NAME                        	RESOURCE TYPE 	ERROR CODE	SEVERITY	MESSAGE
httpbin                     	Service       	IST0108   	Warning 	[my-manifest-dir/service.yaml:1] Unknown annotation: networking.istio.io/non-exist

# for all namespaces
$ getmesh config-validate

NAMESPACE               NAME                    RESOURCE TYPE           ERROR CODE      SEVERITY        MESSAGE
default                 bookinfo-gateway        Gateway                 IST0101         Error           Referenced selector not found: "app=nonexisting"
bookinfo                default                 Peerauthentication      KIA0505         Error           Destination Rule disabling namespace-wide mTLS is missing
bookinfo                bookinfo-gateway        Gateway                 KIA0302         Warning         No matching workload found for gateway selector in this namespace

# for a specific namespace
$ getmesh config-validate -n bookinfo

NAME                    RESOURCE TYPE           ERROR CODE      SEVERITY        MESSAGE
bookinfo-gateway        Gateway                 IST0101         Error           Referenced selector not found: "app=nonexisting"
bookinfo-gateway        Gateway                 KIA0302         Warning         No matching workload found for gateway selector in this namespace

# for a specific namespace with Error as threshold for validation
$ getmesh config-validate -n bookinfo --output-threshold Error

NAME                    RESOURCE TYPE           ERROR CODE      SEVERITY        MESSAGE
bookinfo-gateway        Gateway                 IST0101         Error           Referenced selector not found: "app=nonexisting"

The following is the explanation of each column:
[NAMESPACE]
namespace of the resource

[NAME]
name of the resource

[RESOURCE TYPE]
resource type, i.e. kind, of the resource

[ERROR CODE]
The error code of the found issue which is prefixed by 'IST' or 'KIA'. Please refer to
- https://istio.io/latest/docs/reference/config/analysis/ for 'IST' error codes
- https://kiali.io/documentation/latest/validations/ for 'KIA' error codes

[SEVERITY] the severity of the found issue

[MESSAGE] the detailed message of the found issue
```

#### Options

```
  -n, --namespace string          namespace for config validation
      --output-threshold string   severity level of analysis at which to display messages. Valid values: [Error Warning Info] (default "Info")
  -h, --help                      help for config-validate
```

#### Options inherited from parent commands

```
  -c, --kubeconfig string   Kubernetes configuration file
```

#### SEE ALSO

* [getmesh](/getmesh-cli/reference/getmesh/)	 - getmesh is an integration and lifecycle management CLI tool that ensures the use of supported and trusted versions of Istio.

