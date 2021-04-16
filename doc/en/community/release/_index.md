---
title: "Release process"
url: /community/release
---

On any new release tag on this repository, our release workflow defined in `.github/workflows/release.yaml` 
will be triggered to push the built binaries, `manifest.json`, and `download.sh` to
[GetIstio's repository on Cloudsmith](https://dl.getistio.io/public/raw/). Please note that `manifest.json` and `download.sh`
are not tagged at Cloudsmith level, so they are overwritten by the new revision. The reason for that is because it is convenient to have "static" URLs for these two resources.

In order to cut a new release tag and release the new version from the main branch, you should create a PR 
where `GETISTIO_LATEST_VERSION` in [download.sh](https://github.com/tetratelabs/getistio/blob/13c222fc020e35bd73ce8041c93294278971a226/download.sh#L5) is updated to be the new release tag. 
In this way, the `download.sh` would behave so that it would download the new version by default.
