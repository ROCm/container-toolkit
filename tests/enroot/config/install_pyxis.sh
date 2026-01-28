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
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive apt --fix-broken install
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive apt update
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive apt install -y devscripts
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive apt install -y debhelper
sudo rm -rf pyxis
mkdir -p pyxis_main
cd pyxis_main
git clone https://github.com/NVIDIA/pyxis
cd pyxis && pwd && make orig && make deb
sudo dpkg -i ../nvslurm-plugin-pyxis_*_amd64.deb
sudo mkdir /etc/slurm/plugstack.conf.d
sudo ln -s /usr/share/pyxis/pyxis.conf /etc/slurm/plugstack.conf.d/pyxis.conf
sudo touch /etc/slurm/plugstack.conf
echo "include /etc/slurm/plugstack.conf.d/*" | sudo tee -a /etc/slurm/plugstack.conf
sudo systemctl restart slurmctld slurmd
srun -h | grep container-image
cd ../../
sudo rm -rf pyxis_main

