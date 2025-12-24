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
yes "Y" | sudo  DEBIAN_FRONTEND=noninteractive apt --fix-broken install
yes "Y" | sudo  DEBIAN_FRONTEND=noninteractive apt install mariadb-server
sudo systemctl enable --now mariadb
sudo mysql << EOF
CREATE DATABASE slurm_acct_db;
DROP USER IF EXISTS 'slurm'@'localhost';
GRANT ALL ON slurm_acct_db.* TO 'slurm'@'localhost' IDENTIFIED BY 'STRONG_PASSWORD';
FLUSH PRIVILEGES;
EXIT;
EOF
sudo chown slurm:slurm /etc/slurm/slurmdbd.conf
sudo chmod 600 /etc/slurm/slurmdbd.conf
sudo mkdir -p /var/log/slurm
sudo chown slurm:slurm /var/log/slurm
sudo systemctl restart mariadb
sudo systemctl enable slurmdbd
sudo systemctl start slurmdbd
sudo systemctl status slurmdbd
sudo systemctl restart slurmctld
sudo systemctl restart slurmd