#!/bin/bash
#
# Installs the amdgpu driver packages from repo.radeon.com so the package
# install test can exercise the "driver present" path. Containers share the
# runner's kernel and have no kernel headers, so the DKMS module build cannot
# succeed. We only require the driver package to be installed, not the module to
# be built or loaded. Run as root in a disposable container.
#
# Env:
#   AMDGPU_REPO_VERSION  repo.radeon.com/amdgpu channel (default: latest)
#   UBUNTU_CODENAME      jammy|noble|resolute (deb only)
#   EL_VERSION           9|10 (rpm only)

set -euxo pipefail

AMDGPU_REPO_VERSION="${AMDGPU_REPO_VERSION:-latest}"
GPG_KEY_URL="https://repo.radeon.com/rocm/rocm.gpg.key"

# dpkg leaves a package whose postinst (the DKMS build) fails in a
# half-configured state. Treat installed, half-configured and unpacked as "the
# package is present"; anything else is a real failure.
deb_package_present() {
    local status
    status=$(dpkg-query -W -f='${Status}' "$1" 2>/dev/null) || return 1
    case "$status" in
        *" installed"|*" half-configured"|*" unpacked") return 0 ;;
        *) return 1 ;;
    esac
}

install_deb() {
    : "${UBUNTU_CODENAME:?Set UBUNTU_CODENAME for deb driver install}"

    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install -y --no-install-recommends ca-certificates curl gnupg

    install -d -m 0755 /etc/apt/keyrings
    curl -fsSL "$GPG_KEY_URL" | gpg --dearmor -o /etc/apt/keyrings/rocm.gpg
    chmod 0644 /etc/apt/keyrings/rocm.gpg

    echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amdgpu/${AMDGPU_REPO_VERSION}/ubuntu ${UBUNTU_CODENAME} main" \
        > /etc/apt/sources.list.d/amdgpu.list
    apt-get update

    # The DKMS postinst build fails without kernel headers, so the install may
    # exit non-zero even though the package unpacked. Tolerate only that case.
    if ! apt-get install -y amdgpu-dkms; then
        deb_package_present amdgpu-dkms
        echo "amdgpu-dkms unpacked; DKMS module build skipped (no kernel headers in container)" >&2
    fi
    deb_package_present amdgpu-dkms
}

install_rpm() {
    : "${EL_VERSION:?Set EL_VERSION for rpm driver install}"

    cat > /etc/yum.repos.d/amdgpu.repo <<EOF
[amdgpu]
name=amdgpu
baseurl=https://repo.radeon.com/amdgpu/${AMDGPU_REPO_VERSION}/rhel/${EL_VERSION}/main/x86_64/
enabled=1
priority=50
gpgcheck=1
gpgkey=${GPG_KEY_URL}
EOF

    # dkms lives in EPEL, not the UBI repos.
    dnf install -y "https://dl.fedoraproject.org/pub/epel/epel-release-latest-${EL_VERSION}.noarch.rpm"

    # The %post DKMS build fails without kernel headers. Retry without running
    # scriptlets so the package still lands on disk.
    if ! dnf install -y amdgpu-dkms; then
        dnf install -y --setopt=tsflags=noscripts amdgpu-dkms
        echo "amdgpu-dkms installed with scriptlets skipped (no kernel headers in container)" >&2
    fi
    rpm -q amdgpu-dkms
}

case "${1:?Specify rpm or deb}" in
    deb) install_deb ;;
    rpm) install_rpm ;;
    *) echo "Unknown package format: $1" >&2; exit 1 ;;
esac
