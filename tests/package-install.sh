#!/bin/bash
#
# Builds the rpm or deb package with the Makefile packaging targets, installs
# it and checks the installed CLI. The binaries must already be in bin/rpmbuild
# (rpm) or bin/deb (deb). Needs make and git, plus rpmbuild for rpm. jq is
# only used by the checks below; the package itself must not depend on it. It
# installs, upgrades and removes system packages and writes to /etc, so run it
# as root in a disposable container.

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

# An upgrade must not touch CDI specs, the Docker config or GPU tracker state.
mkdir -p /etc/cdi /etc/docker
echo '{"cdiVersion": "0.5.0", "kind": "vendor.example/device", "devices": []}' > /etc/cdi/acme.json
echo '{"cdiVersion": "0.5.0", "kind": "amd.com/gpu", "devices": []}' > /etc/cdi/amd.json
echo '{}' > /var/log/gpu-tracker.json
/usr/bin/amd-ctk runtime configure

# Reinstalling the same version runs the old package's prerm/%preun as an upgrade.
case "$pkg" in
    rpm) rpm -Uvh --replacepkgs "$workdir"/rpm/RPMS/*/*.rpm ;;
    deb) dpkg -i bin/amd-container-toolkit_*_amd64.deb ;;
esac
test -f /etc/cdi/acme.json
test -f /etc/cdi/amd.json
test -f /var/log/gpu-tracker.json
jq -e '.runtimes.amd' /etc/docker/daemon.json

case "$pkg" in
    rpm) deps=$(rpm -qR amd-container-toolkit) ;;
    deb) deps=$(dpkg-query -W -f='${Depends}' amd-container-toolkit) ;;
esac
if grep -qw jq <<< "$deps"; then
    echo "The package must not depend on jq: $deps" >&2
    exit 1
fi

# Removal takes away only what the toolkit owns.
case "$pkg" in
    rpm) rpm -e amd-container-toolkit ;;
    deb) dpkg -r amd-container-toolkit ;;
esac
test -f /etc/cdi/acme.json
test ! -e /etc/cdi/amd.json
jq -e '.runtimes.amd == null and .features.cdi == null' /etc/docker/daemon.json
case "$pkg" in
    # RPM has no purge, so erase drops tracker state; deb keeps it until purge.
    rpm) test ! -e /var/log/gpu-tracker.json ;;
    deb) test -f /var/log/gpu-tracker.json ;;
esac
