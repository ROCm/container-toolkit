#!/bin/bash
# Copyright (c) Advanced Micro Devices, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the \"License\");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an \"AS IS\" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail
export sudo DEBIAN_FRONTEND=noninteractive

sudo grep -qxF "nameserver 8.8.8.8" /etc/resolv.conf || sudo sh -c 'echo "nameserver 8.8.8.8" >> /etc/resolv.conf'
mkdir -p enroot
cd enroot

ARCH=$(dpkg --print-architecture)
ENROOT_VERSION="${1:-latest}"
BASE_URL="https://github.com/NVIDIA/enroot/releases"

if [[ "$ENROOT_VERSION" == "latest" ]]; then
    echo "Fetching latest Enroot release tag..."

    ENROOT_VERSION=$(curl -fsSL \
        https://api.github.com/repos/NVIDIA/enroot/releases/latest \
        | grep -Po '"tag_name":\s*"v\K[0-9.]+' )

    if [[ -z "$ENROOT_VERSION" ]]; then
        echo "Failed to determine latest Enroot version"
        exit 1
    fi
fi

echo "Resolved Enroot version: $ENROOT_VERSION"
ENROOT_DEB="enroot_${ENROOT_VERSION}-1_${ARCH}.deb"
CAPS_DEB="enroot+caps_${ENROOT_VERSION}-1_${ARCH}.deb"

ENROOT_URL="${BASE_URL}/download/v${ENROOT_VERSION}/${ENROOT_DEB}"
CAPS_URL="${BASE_URL}/download/v${ENROOT_VERSION}/${CAPS_DEB}"

# -----------------------------
# Download
# -----------------------------
echo "Downloading Enroot packages..."
wget -c "$ENROOT_URL" || {
	    echo "Download failed: $DOWNLOAD_URL"
        echo "Please check the enroot version"
    exit 1
}

wget -c "$CAPS_URL" || {
	    echo "Download failed: $DOWNLOAD_URL"
    exit 1
}

echo "Enroot and Enroot+caps (.deb) downloaded successfully"

sudo apt install -y ./*.deb
yes "Y" | sudo apt --fix-broken install
enroot version 
which enroot
hash -r
