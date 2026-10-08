#!/bin/bash
#
# Builds the rpm or deb package with the Makefile packaging targets, installs
# it and checks the installed CLI. The binaries must already be in bin/rpmbuild
# (rpm) or bin/deb (deb). Needs make, git and jq, plus rpmbuild for rpm. It
# installs system packages, so run it as root in a disposable container.

set -euxo pipefail

cd "$(dirname "$0")/.."
if [ -d /sys/module/amdgpu/drivers/ ]; then
    echo "SKIP: the amdgpu driver is loaded, so this run does not check installation without it" >&2
fi
workdir=$(mktemp -d)

case "${1:?Specify rpm or deb}" in
    rpm)
        make rpm-pkg-only CONTAINER_WORKDIR="$PWD" RPM_TOPDIR="$workdir/rpm"
        rpm -ivh "$workdir"/rpm/RPMS/*/*.rpm
        rpm -q amd-container-toolkit
        ;;
    deb)
        make deb-pkg-only
        dpkg -i bin/amd-container-toolkit_*_amd64.deb
        dpkg-query -W amd-container-toolkit
        ;;
    *)
        echo "Unknown package format: $1" >&2
        exit 1
        ;;
esac

/usr/bin/amd-ctk version
/usr/bin/amd-ctk runtime configure --config-path="$workdir/daemon.json"
jq -e '.runtimes.amd.path == "amd-container-runtime"' "$workdir/daemon.json"
