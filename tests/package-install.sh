#!/bin/bash
#
# Builds the rpm or deb package with the Makefile packaging targets, installs
# it and checks the installed CLI. The binaries must already be in bin/rpmbuild
# (rpm) or bin/deb (deb). Needs make, git and jq, plus rpmbuild for rpm. It
# installs system packages, so run it as root in a disposable container.

set -euxo pipefail

cd "$(dirname "$0")/.."

pkg="${1:?Specify rpm or deb}"

# EXPECT_DRIVER controls which install path this run exercises:
#   none       the amdgpu driver must not be present (the default)
#   installed  the amdgpu driver package must have been installed first
# When "installed", fail loudly if the driver package is missing, so a silently
# skipped driver install never passes as the driver-present case.
case "${EXPECT_DRIVER:-none}" in
    installed)
        case "$pkg" in
            rpm) rpm -q amdgpu-dkms ;;
            deb) dpkg-query -W amdgpu-dkms ;;
        esac
        ;;
    *)
        if [ -d /sys/module/amdgpu/drivers/ ]; then
            echo "SKIP: the amdgpu driver is loaded, so this run does not check installation without it" >&2
        fi
        ;;
esac

workdir=$(mktemp -d)

case "$pkg" in
    rpm)
        make rpm-pkg-only CONTAINER_WORKDIR="$PWD" RPM_TOPDIR="$workdir/rpm"
        rpm -ivh "$workdir"/rpm/RPMS/*/*.rpm
        rpm -q amd-container-toolkit
        ;;
    deb)
        make deb-pkg-only UBUNTU_VERSION="${UBUNTU_VERSION:-jammy}"
        dpkg -i bin/amd-container-toolkit_*_amd64.deb
        dpkg-query -W amd-container-toolkit
        ;;
    *)
        echo "Unknown package format: $pkg" >&2
        exit 1
        ;;
esac

/usr/bin/amd-ctk version
/usr/bin/amd-ctk runtime configure --config-path="$workdir/daemon.json"
jq -e '.runtimes.amd.path == "amd-container-runtime"' "$workdir/daemon.json"
