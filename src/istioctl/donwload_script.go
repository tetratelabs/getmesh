// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package istioctl

// copied from https://github.com/istio/istio/blob/984704533cc356908f7f56323e50f7f840c4cdc3/release/downloadIstioCandidate.sh
// with changing the download url to Tetrate's cloudsmith

const downloadScript = `
set -e

# Determines the operating system.
OS="$(uname)"
if [ "x${OS}" = "xDarwin" ] ; then
  OSEXT="osx"
else
  OSEXT="linux"
fi

LOCAL_ARCH=$(uname -m)
if [ "${TARGET_ARCH}" ]; then
    LOCAL_ARCH=${TARGET_ARCH}
fi

case "${LOCAL_ARCH}" in
  x86_64)
    ISTIO_ARCH=amd64
    ;;
  armv8*)
    ISTIO_ARCH=arm64
    ;;
  aarch64*)
    ISTIO_ARCH=arm64
    ;;
  armv*)
    ISTIO_ARCH=armv7
    ;;
  amd64|arm64)
    ISTIO_ARCH=${LOCAL_ARCH}
    ;;
  *)
    echo "This system's architecture, ${LOCAL_ARCH}, isn't supported"
    exit 1
    ;;
esac

URL="https://dl.cloudsmith.io/public/tetrate/getistio/raw/files/istio-${DISTRIBUTION_IDENTIFIER}-${OSEXT}.tar.gz"
ARCH_URL="https://dl.cloudsmith.io/public/tetrate/getistio/raw/files/istio-${DISTRIBUTION_IDENTIFIER}-${OSEXT}-${ISTIO_ARCH}.tar.gz"

with_arch() {
  printf "\nDownloading %s from %s ...\n" "$DISTRIBUTION_IDENTIFIER" "$ARCH_URL"
  if ! curl -o /dev/null -sIf "$ARCH_URL"; then
    printf "\n%s is not found, please specify a valid ISTIO_VERSION and TARGET_ARCH\n" "$ARCH_URL"
    exit 1
  fi
  curl -fsLO "$ARCH_URL"
  filename="istio-${DISTRIBUTION_IDENTIFIER}-${OSEXT}-${ISTIO_ARCH}.tar.gz"
  tar -xzf "${filename}" --strip 1
  rm "${filename}"
}

without_arch() {
  printf "\nDownloading %s from %s ..." "$DISTRIBUTION_IDENTIFIER" "$URL"
  if ! curl -o /dev/null -sIf "$URL"; then
    printf "\n%s is not found, please specify a valid ISTIO_VERSION\n" "$URL"
    exit 1
  fi
  curl -fsLO "$URL"
  filename="istio-${DISTRIBUTION_IDENTIFIER}-${OSEXT}.tar.gz"
  tar -xzf "${filename}" --strip 1
  rm "${filename}"
}

# Istio 1.6 and above support arch
ARCH_SUPPORTED=$(echo "$ISTIO_VERSION" | awk  '{ ARCH_SUPPORTED=substr($0, 1, 3); print ARCH_SUPPORTED; }' )
# Istio 1.5 and below do not have arch support
ARCH_UNSUPPORTED="1.5"

if [ "${OS}" = "Linux" ] ; then
  # This checks if 1.6 <= 1.5 or 1.4 <= 1.5
  if [ "$(expr "${ARCH_SUPPORTED}" \<= "${ARCH_UNSUPPORTED}")" -eq 1 ]; then
    without_arch
  else
    with_arch
  fi
elif [ "x${OS}" = "xDarwin" ] ; then
  without_arch
else
  printf "\n\n"
  printf "Unable to download Istio %s at this moment!\n" "$ISTIO_VERSION"
  printf "Please verify the version you are trying to download.\n\n"
  exit 1
fi

printf ""
printf "\nIstio %s Download Complete!\n" "$ISTIO_VERSION"
printf "\n"
printf "Istio has been successfully downloaded into your system.\n" "$DISTRIBUTION_IDENTIFIER"
printf "\n"
`
