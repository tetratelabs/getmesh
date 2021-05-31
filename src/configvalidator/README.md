### How it works

1. execute Kiali's config validator as a library and get the results (in `validation_kiali.go`)
   - If the file paths are provided as command line arguments, then we pass them so that we can locally test these yamls against the live cluster before actually applying.
2. execute `istioctl analyze` directly, and parse the outputted results (in `validation_istio.go`)
   - If the file paths are provided as command line arguments, then we pass them so that we can locally test these yamls against the live cluster before actually applying.
   - Parsing the `istioctl analyze`'s stdout directly is a little hacky (See `parseIstioctlAnalyzeResult` function in `validation_istio.go`).
      After Kiali's upgrade of client-go, we should use istioctl's analysis as a library (See [#44](https://github.com/tetratelabs/getmesh/issues/44)).
3. print the results

