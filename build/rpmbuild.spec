Name: amd-container-toolkit
Version: 1.3.0
Release: 1%{?dist}
Summary: Package containing AMD container tookkit binaries

License: Apache-2.0
URL: https://github.com/ROCm/container-toolkit/

BuildArch: x86_64
Requires: jq

%description
This package contains pre-built binaries for AMD containter toolkit

%preun
/bin/bash /usr/share/amd-container-toolkit/cleanup.sh

%install
base_dir=${CONTAINER_WORKDIR}
install -D -m 0755 ${base_dir}/bin/rpmbuild/amd-ctk  %{buildroot}/usr/bin/amd-ctk
install -D -m 0755 ${base_dir}/bin/rpmbuild/amd-container-runtime  %{buildroot}/usr/bin/amd-container-runtime
install -D -m 0644 ${base_dir}/README.md %{buildroot}%{_pkgdocdir}/README.md
install -D -m 0644 ${base_dir}/LICENSE %{buildroot}%{_licensedir}/%{name}/LICENSE
install -D -m 0755 ${base_dir}/build/cleanup.sh %{buildroot}/usr/share/amd-container-toolkit/cleanup.sh

%files
/usr/bin/amd-ctk
/usr/bin/amd-container-runtime
%doc %{_pkgdocdir}
%license %{_licensedir}/%{name}
/usr/share/amd-container-toolkit/cleanup.sh

%changelog
* Fri Apr 11 2025 1.0-1
- Initial package creation.
