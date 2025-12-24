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

#SBATCH --job-name=test-single-node-pytorch
#SBATCH --nodes=1
#SBATCH --cpus-per-task=8
#SBATCH --time=01:00:00
#SBATCH --output=pytorch_logs/pytorch-util-%j.out
#SBATCH --error=pytorch_logs/pytorch-util-%j.err
set -e
mkdir -p logs

# USER CONFIGURABLE
IMAGE_NAME=pytorch.sqsh
IMAGE_PATH=$PWD/$IMAGE_NAME
DOCKER_IMAGE=rocm/pytorch:latest

# Create image only once
if [[ ! -f "$IMAGE_PATH" ]]; then
    echo "Creating Enroot image..."
    enroot import -o "$IMAGE_PATH" docker://$DOCKER_IMAGE
else
    echo "Using existing Enroot image"
fi

# Create workspace directory on each node and write the Python script
srun -n1 --container-image="$IMAGE_PATH" \
         --container-mounts=/tmp/test_pytorch:/ws/test_slurm \
         --container-workdir=/ws/test_slurm \
         bash -lc '
python3 - << EOP
import torch, time
print("CUDA available:", torch.cuda.is_available())
print("Number of GPUs:", torch.cuda.device_count())
EOP
rocm-smi
'
srun      --container-image="$IMAGE_PATH" \
          --container-mounts=/tmp/test_pytorch:/ws/test_slurm \
          --container-workdir=/ws/test_slurm \
     bash -lc '
     export HOME=/ws/test_slurm
     mkdir -p /ws/test_slurm/.config/miopen
     export MIOPEN_USER_CACHE_DIR=/ws/test_slurm/.config/miopen
     export MIOPEN_CUSTOM_CACHE_DIR=/ws/test_slurm/.config/miopen
     mkdir -p /ws/test_slurm/miopen_cache
     chmod +x /ws/test_slurm/gpu_stress_10s.py
     torchrun --standalone --nproc_per_node=4 /ws/test_slurm/gpu_stress_10s.py '