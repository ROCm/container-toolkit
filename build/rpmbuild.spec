Name: amd-container-toolkit
Version: 1.2.0
Release: 1%{?dist}
Summary: Package containing AMD container tookkit binaries

License: GPLv3
URL: https://github.com/ROCm/container-toolkit/

BuildArch: x86_64
Requires: jq

%description
This package contains pre-built binaries for AMD containter toolkit

%pre
# Check if AMD GPU driver is installed before installation
echo "Checking for AMD GPU driver..."
if [ ! -d "/sys/module/amdgpu/drivers/" ]; then
    echo "Error: AMD GPU driver (amdgpu) is not installed or loaded."
    echo "The AMD Container Toolkit requires the amdgpu driver to be installed and loaded."
    echo "Please install the AMD GPU driver before installing the container toolkit."
    exit 1
fi
echo "AMD GPU driver found."

%post
# Initialize GPU tracker after install
/usr/bin/amd-ctk gpu-tracker init || true

%preun
/bin/bash /usr/share/amd-container-toolkit/cleanup.sh

%install
base_dir=${CONTAINER_WORKDIR}
install -D -m 0755 ${base_dir}/bin/rpmbuild/amd-ctk  %{buildroot}/usr/bin/amd-ctk
install -D -m 0755 ${base_dir}/bin/rpmbuild/amd-container-runtime  %{buildroot}/usr/bin/amd-container-runtime
install -D -m 0644 ${base_dir}/README.md %{buildroot}/usr/share/doc/my-binary-package/README.md
install -D -m 0755 ${base_dir}/build/cleanup.sh %{buildroot}/usr/share/amd-container-toolkit/cleanup.sh

%files
/usr/bin/amd-ctk
/usr/bin/amd-container-runtime
/usr/share/doc/my-binary-package/README.md
/usr/share/amd-container-toolkit/cleanup.sh

%changelog
* Fri Apr 11 2025 1.0-1
- Initial package creation.
