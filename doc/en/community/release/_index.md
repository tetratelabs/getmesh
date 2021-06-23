---
title: "Release process"
url: /community/release
---

In order to cut a new release tag and release the new version from the main branch, you should create a PR 
where `GETMESH_LATEST_VERSION` in [install.sh](https://github.com/tetratelabs/getmesh/blob/main/site/install.sh) is updated to be the new release tag. 
In this way, the `install.sh` would behave so that it would download the new version by default.
