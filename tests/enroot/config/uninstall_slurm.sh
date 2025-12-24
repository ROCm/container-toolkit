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

export sudo DEBIAN_FRONTEND=noninteractive

sudo systemctl stop slurmctld
sudo systemctl stop slurmd
yes "Y" | DEBIAN_FRONTEND=noninteractive sudo apt-get remove --purge slurm-*
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive apt-get remove --purge slurmctld slurmd
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive apt autoremove
sudo rm -rf /etc/slurm
sudo rm -rf /var/spool/slurmctld
sudo rm -rf /var/spool/slurmd
sudo rm -rf /var/log/slurm
sudo rm -rf /var/lib/slurm
sudo rm -rf /run/slurm*
sudo rm -rf /usr/local/etc/slurm*
sudo userdel -r slurm
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive  sudo apt-get clean
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive  sudo apt-get autoremove
sudo systemctl stop slurmctld
sudo systemctl stop slurmd
sudo systemctl stop munge
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive  sudo apt-get purge munge munge-libs munge-doc
yes "Y" | sudo DEBIAN_FRONTEND=noninteractive sudo apt-get autoremove --purge
sudo rm -rf /etc/munge /run/munge /var/log/munge /var/lib/munge  
sudo userdel munge
sudo groupdel munge