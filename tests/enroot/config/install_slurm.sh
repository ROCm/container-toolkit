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

export DEBIAN_FRONTEND=noninteractive
sudo grep -qxF "nameserver 8.8.8.8" /etc/resolv.conf || sudo sh -c 'echo "nameserver 8.8.8.8" >> /etc/resolv.conf'
export SLURMUSER=1003
sudo groupadd -g $SLURMUSER slurm
sudo useradd -m -c "SLURM workload manager" -d /var/lib/slurm -u $SLURMUSER -g slurm -s /bin/bash slurm
export MUNGEUSER=1004
sudo groupadd -g $MUNGEUSER munge
sudo useradd -m -c "MUNGE Uid 'N' Gid Emporium" -d /var/lib/munge -u $MUNGEUSER -g munge -s /sbin/nologin munge
id slurm
id munge
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt --fix-broken install
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt install munge libmunge2 libmunge-
sudo mkdir -p  /etc/munge/ /var/log/munge/ /var/lib/munge/ /run/mun
sudo chown -R munge: /etc/munge/ /var/log/munge/ /var/lib/munge/ /run/munge/
sudo chmod 0700 /etc/munge/ /var/log/munge/ /var/lib/munge/
sudo chmod 0755 /run/munge/
sudo  systemctl enable munge
sudo systemctl restart munge
munge -n | unmunge | grep STATUS
DEBIAN_FRONTEND=noninteractive sudo apt-get update
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt --fix-broken install
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt-get install build-essential fakeroot devscripts libmunge-dev libmunge2 munge
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt upgrade
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt install libpmix-dev libpmix2
pwd
mkdir /tmp/slurm
cd /tmp/slurm
pwd
# Code to fetch the latest stable slurm version 
SLURM_VERSION="${1:-latest}"
BASE_URL="https://download.schedmd.com/slurm"

echo "Requested Slurm version: $SLURM_VERSION"
if [[ "$SLURM_VERSION" == "latest" ]]; then
    echo "Fetching latest stable Slurm version..."

    SLURM_TARBALL=$(curl -s "$BASE_URL/" \
        | grep -oE 'slurm-[0-9]+\.[0-9]+\.[0-9]+\.tar\.bz2' \
        | sort -V \
        | tail -1)

    if [[ -z "$SLURM_TARBALL" ]]; then
        echo "Failed to determine latest Slurm version"
        exit 1
    fi
else
    SLURM_TARBALL="slurm-${SLURM_VERSION}.tar.bz2"
fi

DOWNLOAD_URL="${BASE_URL}/${SLURM_TARBALL}"
echo "Downloading: $DOWNLOAD_URL"
wget -c "$DOWNLOAD_URL" || {
	    echo "Download failed: $DOWNLOAD_URL"
        echo "Please check the slurm version"
    exit 1
}
echo "Extracting $SLURM_TARBALL"
tar -xjf "$SLURM_TARBALL"
if [[ ! -s "$SLURM_TARBALL" ]]; then
	    echo "ERROR: Downloaded file is missing or empty: $SLURM_TARBALL"
	        exit 1
fi
EXTRACTED_DIR=$(tar -tf "$SLURM_TARBALL" | head -1 | cut -d/ -f1)

yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt install equivs
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt install libswitch-perl equivs mk-build-deps
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt install devscripts equivs
sudo DEBIAN_FRONTEND=noninteractive mk-build-deps -i -t 'apt-get -y --no-install-recommends' debian/control
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt install libmunge-dev libgtk2.0-dev libpam0g-dev libperl-dev liblua5.3-dev libhwloc-dev dh-exec 
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt install libdbus-1-dev librdkafka
sudo groupadd slurm
sudo useradd -m -r -s /bin/false -d /tmp/slurm -g slurm slurm
cd "$EXTRACTED_DIR"
pwd
sudo DEBIAN_FRONTEND=noninteractive debuild -b -uc -us
sudo mkdir -p /etc/slurm
sudo mkdir -p /var/spool/slurm/savestate
sudo mkdir -p /var/spool/slurmd
sudo touch /var/log/slurmctld.log
sudo chown -R slurm:slurm /var/spool/slurm/savestate
sudo chown -R slurm:slurm /var/spool/slurmd
sudo chown -R slurm:slurm /var/log/slurmctld.log
cd ../ && DEBIAN_FRONTEND=noninteractive sudo dpkg -i slurm-*.deb
systemctl restart slurmctld
systemctl restart slurmd
sinfo

