#!/usr/bin/env bash

# originally copied from https://wasmtime.dev/install.sh with some modifications

GETMESH_LATEST_VERSION="1.0.6"

if [[ -z "${GETMESH_VERSION}" ]]; then
  VERSION="${GETMESH_LATEST_VERSION}"
else
  VERSION="${GETMESH_VERSION}"
fi

if [[ -z "${FETCH_LATEST_ISTIOCTL}" ]]; then
  FETCH_LATEST_ISTIOCTL="true"
else
  FETCH_LATEST_ISTIOCTL="${FETCH_LATEST_ISTIOCTL}"
fi

OS="$(uname)"
LOCAL_ARCH="$(uname -m)"
GETMESH_HOME=${HOME}/.getmesh
GETMESH_BIN_DIR="${GETMESH_HOME}"/bin
mkdir -p "${GETMESH_BIN_DIR}"
mkdir -p "${GETMESH_HOME}"/istio
EXECUTABLE_OUT="${GETMESH_BIN_DIR}"/getmesh

error() {
  command printf '\033[1;31mError\033[0m: %s\n\n' "$1" 1>&2
}

# If file exists, echo it
echo_fexists() {
  [ -f "$1" ] && echo "$1"
}

eprintf() {
  command printf '%s\n' "$1" 1>&2
}


case ${OS} in
  Linux)
    OS=linux
    ;;
  Darwin)
    OS=darwin
    ;;
  *)
    echo "This system's OS, ${LOCAL_ARCH}, isn't supported"
    exit 1
    ;;
esac

case ${LOCAL_ARCH} in
  x86_64)
    LOCAL_ARCH=amd64
    ;;
  *)
    echo "This system's architecture, ${LOCAL_ARCH}, isn't supported"
    exit 1
    ;;
esac

URL="https://dl.getistio.io/public/raw/files/GETMESH_${OS}_${LOCAL_ARCH}_v${VERSION}.tar.gz"

if [[ -n "${GETMESH_TEST_BINRAY_URL}" ]]; then
    URL=${GETMESH_TEST_BINRAY_URL}
fi

printf "\nDownloading getmesh from %s ...\n" "$URL"
if ! curl -o /dev/null -sIf "$URL"; then
  printf "\n%s is not found\n" "$URL"
  exit 1
fi

curl -fsL "$URL" -o getmesh.tar.gz
tar -zxf getmesh.tar.gz -C"${GETMESH_BIN_DIR}"
chmod u+x "${EXECUTABLE_OUT}"
rm getmesh.tar.gz

printf "getmesh Download Complete!\n\n"

detect_profile() {
  local shellname="$1"
  local uname="$2"

  if [ -f "$PROFILE" ]; then
    echo "$PROFILE"
    return
  fi

  # try to detect the current shell
  case "$shellname" in
    bash)
      # based on Ubuntu 20.04 tests - the sequence of the profiles processing 
      # is the same for both Linux and Mac - .bash_profile first and then
      # bashrc, also confirmed here:
      # https://askubuntu.com/questions/161249/bashrc-not-executed-when-opening-new-terminal
      echo_fexists "$HOME/.bash_profile" || echo_fexists "$HOME/.bashrc"
      ;;
    zsh)
      echo "$HOME/.zshrc"
      ;;
    fish)
      echo "$HOME/.config/fish/config.fish"
      ;;
    *)
      # Fall back to checking for profile file existence. Once again, the order
      # differs between macOS and everything else.
      local profiles
      
      profiles=( .profile .bash_profile .bashrc .zshrc .config/fish/config.fish )
          ;;
        *)

      for profile in "${profiles[@]}"; do
        echo_fexists "$HOME/$profile" && break
      done
      ;;
  esac
}

# generate shell code to source the loading script and modify the path for the input profile
build_path_str() {
  local profile="$1"
  local profile_install_dir="$2"

  if [[ $profile =~ \.fish$ ]]; then
    # fish uses a little different syntax to modify the PATH
    cat <<END_FISH_SCRIPT

set -gx GETMESH_HOME "$profile_install_dir"
string match -r ".getmesh" "\$PATH" > /dev/null; or set -gx PATH "\$GETMESH_HOME/bin" \$PATH
END_FISH_SCRIPT
  else
    # bash and zsh
    cat <<END_BASH_SCRIPT

export GETMESH_HOME="$profile_install_dir"
export PATH="\$GETMESH_HOME/bin:\$PATH"
END_BASH_SCRIPT
  fi
}

update_profile() {
  local install_dir="$1"

  local profile_install_dir=$(echo "$install_dir" | sed "s:^$HOME:\$HOME:")
  local detected_profile="$(detect_profile $(basename "/$SHELL") $(uname -s) )"
  local path_str="$(build_path_str "$detected_profile" "$profile_install_dir")"

  if [ -z "${detected_profile-}" ] ; then
    error "No user profile found."
    eprintf "Tried \$PROFILE ($PROFILE), ~/.bashrc, ~/.bash_profile, ~/.zshrc, ~/.profile, and ~/.config/fish/config.fish."
    eprintf ''
    eprintf "You can either create one of these and try again or add this to the appropriate file:"
    eprintf "$path_str"
    return 1
  else
    if ! command grep -qc 'GETMESH_HOME' "$detected_profile"; then
      echo 'Updating' "user profile ($detected_profile)..."
      printf "The following two lines are added into your profile (%s):\n" $detected_profile
      printf "$path_str\n\n"
      command printf "$path_str" >> "$detected_profile"
    fi
  fi
}

update_profile "${GETMESH_HOME}"

if [ "${FETCH_LATEST_ISTIOCTL}" = "true" ]; then
  printf "Downloading latest istio ...\n"
  ${EXECUTABLE_OUT} fetch
fi

printf "Finished installation. Open a new terminal to start using getmesh!\n"
