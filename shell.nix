# This is a shell.nix file used to describe the environment that getmesh needs
# for development.
#
# For more information about this and why this file is useful, see here:
# https://nixos.org/guides/nix-pills/developing-with-nix-shell.html
#
# Also look into direnv: https://direnv.net/, this can make it so that you can
# automatically get your environment set up when you change folders into the
# project.
{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/94db887ac729c16e14df87ad16db89221e840fb1.tar.gz") {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    golangci-lint
    kubectl
  ];
  shellHook = ''
    export PATH=$PATH:$(go env GOPATH)/bin
    go install github.com/google/addlicense@v1.0.0

    # This is here for covenience. Hence we can run the e2e test against a prepared cluster.
    export KUBECONFIG=/tmp/output/kubeconfig
    export K3S_VERSION=v1.21.2-k3s1
    export K3S_CONTAINER_NAME=getmesh-k3s-$VERSION
    docker ps --all | grep $K3S_CONTAINER_NAME | cut -d ' ' -f1 | xargs docker container rm -f || true
    docker run -d --privileged --name=$K3S_CONTAINER_NAME -e K3S_KUBECONFIG_OUTPUT=$KUBECONFIG -e K3S_KUBECONFIG_MODE=666 -v /tmp/output:/tmp/output -p 6443:6443 rancher/k3s:$K3S_VERSION server
    sleep 10
    kubectl wait --for=condition=Ready $(kubectl get nodes --no-headers -oname)
  '';
}
