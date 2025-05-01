Name: amd-container-toolkit
Version: 1.0
Release: 1%{?dist}
Summary: Package containing AMD container tookkit binaries

License: GPLv3
URL: https://github.com/ROCm/container-toolkit/

BuildArch: x86_64

%description
This package contains pre-built binaries for AMD containter toolkit

%install
base_dir=${CONTAINER_WORKDIR}
install -D -m 0755 ${base_dir}/bin/rpmbuild/amd-ctk  %{buildroot}/usr/bin/amd-ctk
install -D -m 0755 ${base_dir}/bin/rpmbuild/amd-container-runtime  %{buildroot}/usr/bin/amd-container-runtime
install -D -m 0644 ${base_dir}/README.md %{buildroot}/usr/share/doc/my-binary-package/README.md

%files
/usr/bin/amd-ctk
/usr/bin/amd-container-runtime
/usr/share/doc/my-binary-package/README.md

%changelog
* Fri Apr 11 2025 1.0-1
- Initial package creation.