#!/bin/bash

set -euxo pipefail

test ! -d /sys/module/amdgpu/drivers/
workdir=$(mktemp -d)

case "${1:?Specify rpm or deb}" in
    rpm)
        CONTAINER_WORKDIR="$PWD" rpmbuild -bb \
            --define "_topdir $workdir/rpm" build/rpmbuild.spec
        rpm -ivh "$workdir"/rpm/RPMS/x86_64/*.rpm
        rpm -q amd-container-toolkit
        ctk=/usr/bin/amd-ctk
        ;;
    deb)
        cp -a build/debian "$workdir/package"
        sed -i 's/BUILD_VER_ENV/0/' "$workdir/package/DEBIAN/control"
        install -m 0755 build/cleanup.sh "$workdir/package/DEBIAN/prerm"
        for binary in amd-ctk amd-container-runtime; do
            install -D -m 0755 "bin/rpmbuild/$binary" \
                "$workdir/package/usr/local/bin/$binary"
        done
        dpkg-deb --build "$workdir/package" "$workdir/toolkit.deb"
        dpkg -i "$workdir/toolkit.deb"
        dpkg-query -W amd-container-toolkit
        ctk=/usr/local/bin/amd-ctk
        ;;
    *)
        echo "Unknown package format: $1" >&2
        exit 1
        ;;
esac

"$ctk" version
"$ctk" runtime configure --config-path="$workdir/daemon.json"
jq -e '.runtimes.amd.path == "amd-container-runtime"' "$workdir/daemon.json"

if "$ctk" gpu list >"$workdir/gpu-list.log" 2>&1; then
    cat "$workdir/gpu-list.log"
    echo "GPU discovery succeeded without the amdgpu driver" >&2
    exit 1
fi
cat "$workdir/gpu-list.log"
grep -F 'amdgpu driver unavailable' "$workdir/gpu-list.log"
