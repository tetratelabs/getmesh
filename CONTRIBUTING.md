## Building & Testing

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
make unit-test
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

## Contributing

We welcome contributions from the community. Please read the following guidelines carefully to maximize the chances of your PR being merged.

### Code Reviews

* Indicate the priority of each comment, following this
  [feedback ladder](https://www.netlify.com/blog/2020/03/05/feedback-ladders-how-we-encode-code-reviews-at-netlify/).
  If none was indicated it will be treated as `[dust]`.
* A single approval is sufficient to merge, except when the change cuts
  across several components; then it should be approved by at least one owner
  of each component. If a reviewer asks for changes in a PR they should be
  addressed before the PR is merged, even if another reviewer has already
  approved the PR.
* During the review, address the comments and commit the changes _without_ squashing the commits.
  This facilitates incremental reviews since the reviewer does not go through all the code again to
  find out what has changed since the last review.

### DCO

All authors to the project retain copyright to their work.
However, to ensure that they are only submitting work that they have rights to,
we are requiring everyone to acknowledge this by signing their work.

The sign-off is a simple line at the end of the explanation for the
patch, which certifies that you wrote it or otherwise have the right to
pass it on as an open-source patch. The rules are pretty simple: if you
can certify the below (from
[developercertificate.org](https://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1
Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA
Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.
Developer's Certificate of Origin 1.1
By making a contribution to this project, I certify that:
(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or
(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or
(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.
(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe@gmail.com>

using your real name (sorry, no pseudonyms or anonymous contributions.)

You can add the sign off when creating the git commit via `git commit -s`. 

Or, you can sign off the whole PR via `git rebase --signoff main`.


## Release

On any new release tag on this repository, our release workflow defined in `.github/workflows/release.yaml` 
will be triggered to push the built binaries, `manifest.json`, and `download.sh` to
[getmesh's repository on Cloudsmith](https://dl.getmesh.io/public/raw/). Please note that `manifest.json` and `download.sh`
are not tagged at Cloudsmith level, so they are overwritten by the new revision. The reason for that is because it is convenient to have "static" URLs for these two resources.

In order to cut a new release tag and release the new version from the main branch, you should create a PR 
where `getmesh_LATEST_VERSION` in [download.sh](https://github.com/tetratelabs/getmesh/blob/13c222fc020e35bd73ce8041c93294278971a226/download.sh#L5) is updated to be the new release tag. 
In this way, the `download.sh` would behave so that it would download the new version by default.
