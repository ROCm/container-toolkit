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
        # Capture rpm output before grep -q: grep exits on the first match,
        # and under pipefail a SIGPIPE on rpm's next write fails the test.
        files=$(rpm -ql amd-container-toolkit)
        grep -qx 'Apache-2.0' <<< "$(rpm -q --qf '%{LICENSE}\n' amd-container-toolkit)"
        grep -qx '/usr/share/doc/amd-container-toolkit' <<< "$files"
        grep -qx '/usr/share/licenses/amd-container-toolkit' <<< "$files"
        grep -qx '/usr/share/doc/amd-container-toolkit/README.md' <<< "$(rpm -qd amd-container-toolkit)"
        grep -qx '/usr/share/licenses/amd-container-toolkit/LICENSE' <<< "$(rpm -qL amd-container-toolkit)"
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

# Reinstalling the same version runs the old package's prerm/%preun as an upgrade.
install_pkg() {
    case "$pkg" in
        rpm) rpm -Uvh --replacepkgs "$workdir"/rpm/RPMS/*/*.rpm ;;
        deb) dpkg -i bin/amd-container-toolkit_*_amd64.deb ;;
    esac
}

remove_pkg() {
    case "$pkg" in
        rpm) rpm -e amd-container-toolkit ;;
        deb) dpkg -r amd-container-toolkit ;;
    esac
}

plant_specs() {
    mkdir -p /etc/cdi
    echo '{"cdiVersion": "0.5.0", "kind": "vendor.example/device", "devices": []}' > /etc/cdi/acme.json
    echo '{"cdiVersion": "0.5.0", "kind": "amd.com/gpu", "devices": []}' > /etc/cdi/amd.json
}

# An upgrade must not touch CDI specs, the Docker config or GPU tracker state.
plant_specs
mkdir -p /etc/docker
echo '{"runtimes": {"other": {"path": "other-runtime"}}}' > /etc/docker/daemon.json
echo '{}' > /var/log/gpu-tracker.json
/usr/bin/amd-ctk runtime configure --set-as-default
# configure replaces the whole features map, so add the extra feature afterwards.
jq '.features["containerd-snapshotter"] = true' /etc/docker/daemon.json > "$workdir/daemon.json.new"
mv "$workdir/daemon.json.new" /etc/docker/daemon.json

install_pkg
test -f /etc/cdi/acme.json
test -f /etc/cdi/amd.json
test -f /var/log/gpu-tracker.json
jq -e '.["default-runtime"] == "amd" and .runtimes.amd' /etc/docker/daemon.json

case "$pkg" in
    rpm) deps=$(rpm -qR amd-container-toolkit) ;;
    deb) deps=$(dpkg-query -W -f='${Depends}' amd-container-toolkit) ;;
esac
if grep -qw jq <<< "$deps"; then
    echo "The package must not depend on jq: $deps" >&2
    exit 1
fi

# Removal takes away only what the toolkit owns.
remove_pkg
test -f /etc/cdi/acme.json
test ! -e /etc/cdi/amd.json
jq -e '. == {"runtimes": {"other": {"path": "other-runtime"}}, "features": {"containerd-snapshotter": true}}' /etc/docker/daemon.json
case "$pkg" in
    # RPM has no purge, so erase drops tracker state; deb keeps it until purge.
    rpm) test ! -e /var/log/gpu-tracker.json ;;
    deb) test -f /var/log/gpu-tracker.json ;;
esac

# A daemon.json that cannot be parsed must neither block removal nor keep the
# AMD CDI spec behind.
install_pkg
plant_specs
echo '{bad' > /etc/docker/daemon.json
remove_pkg
test -f /etc/cdi/acme.json
test ! -e /etc/cdi/amd.json
test "$(cat /etc/docker/daemon.json)" = '{bad'
