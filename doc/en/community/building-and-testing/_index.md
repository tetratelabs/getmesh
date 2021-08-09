---
title: "Building and Testing"
url: /community/building-and-testing
---

Before you proceed, please make sure that you have the following dependencies available in your machine:

- https://github.com/google/addlicense
- https://github.com/golangci/golangci-lint
- a Kubernetes cluster (e.g. https://kind.sigs.k8s.io/)


### Run linter

Here we use `golangci-lint` configured in `.golangci.yml` for static analysis, so please make sure that you have it installed.

To run linter, simply execute:

```
make lint
```

### Run unittests

Running unittests does not require any k8s cluster, and it can be done by

```
make test
```

### Build binary

```
make build
```

### Run e2e tests

Running end-to-end tests requires you to have a valid k8s context. Please note that **e2e will use your default kubeconfig and default context**.

In order to run e2e tests, execute:

```
make e2e-test
```

### Build auto-generated docs

```
make doc-gen
```

### Add license headers

We require every source code to have the specified license header. Adding the header can be done by
```
make license
```

