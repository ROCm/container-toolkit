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

#SBATCH --job-name=pytorch-nccl-multinode
#SBATCH --nodes=2
#SBATCH --ntasks-per-node=8
#SBATCH --cpus-per-task=8
#SBATCH --time=01:00:00
#SBATCH --output=pytorch_logs/pytorch-rccl-%j.out
#SBATCH --error=pytorch_logs/pytorch-rccl-%j.err

set -e

# Create logs directory
mkdir -p pytorch_logs

# USER CONFIGURABLE
export IMAGE_NAME=ubuntu22_pytorch.sqsh
export DOCKER_IMAGE=docker://rocm/pytorch:rocm7.0.2_ubuntu22.04_py3.10_pytorch_release_2.7.1
export ENROOT_DATA_PATH=/tmp/enroot/data
export ENROOT_CACHE_PATH=/tmp/enroot/cache
export ENROOT_RUNTIME_PATH=/tmp/enroot/runtime 

# Stage image ONCE per node
# Create temporary directories on all nodes
echo "Creating temporary directories on all nodes..."
srun --ntasks-per-node=1 bash -lc 'mkdir -p /tmp/xdg-$SLURM_NODEID/enroot /tmp/xdg-cache-$SLURM_NODEID /tmp/enroot/cache /tmp/enroot/data /tmp/enroot/runtime'

# Create temporary directories on all nodes
echo "Creating enroot image on all nodes..."
srun --nodes=2 --ntasks=2 \
     --ntasks-per-node=1 \
     bash -c '
    mkdir -p "$ENROOT_DATA_PATH" "$ENROOT_CACHE_PATH"

        if [ ! -f "$ENROOT_DATA_PATH/$IMAGE_NAME" ]; then
            echo "[$(hostname)] Importing enroot image..."
            enroot import -o "$ENROOT_DATA_PATH/$IMAGE_NAME" "$DOCKER_IMAGE"
        else
            echo "[$(hostname)] Image already present"
        fi
     '


# Barrier (all nodes must have image)
srun --nodes=2 --ntasks=2 \
     --ntasks-per-node=1 \
     bash -c 'echo "[\$(hostname)] Image ready"'

retry_command() {
    local max_attempts=$1
    local command=$2
    local result=""

    for attempt in $(seq 1 $max_attempts); do
        echo "  Attempt $attempt/$max_attempts..." >&2
        result=$(eval "$command" 2>/dev/null | head -n 1)

        if [ -n "$result" ]; then
            echo "$result"
            return 0
        fi

        if [ $attempt -lt $max_attempts ]; then
            sleep 1
        fi
    done
    return 1
}

echo "================================================"
echo "Job Configuration:"
echo "  Job ID: $SLURM_JOB_ID"
echo "  Nodes: ${SLURM_JOB_NODELIST}"
echo "  Total Tasks: $SLURM_NTASKS"
echo "  Tasks/Node: $SLURM_NTASKS_PER_NODE"
echo "  Number of Nodes: $SLURM_JOB_NUM_NODES"
echo "================================================"

# Get master node IP using srun
echo "Detecting master node IP..."
MASTER_ADDR=$(srun --nodes=1 --ntasks=1 bash -c 'getent hosts $(hostname -s) | awk "{print \$1}"' | head -n 1)
MASTER_PORT=29500

# Validate we got an IP
if [ -z "$MASTER_ADDR" ]; then
    echo "ERROR: Could not determine master IP address!"
    exit 1
fi
echo "Master: $MASTER_ADDR:$MASTER_PORT"

# Detect network interface for the master IP
echo "Detecting network interface..."
SOCKET_IFNAME=$(retry_command 3 "srun --nodes=1 --ntasks=1 bash -c 'ip -o addr show | grep \"inet $MASTER_ADDR\" | awk \"{print \\\$2}\"'")
echo "NCCL_SOCKET_IFNAME: $SOCKET_IFNAME"

# Validate SOCKET_IFNAME has been determined 
if [ -z "$SOCKET_IFNAME" ]; then
    echo "ERROR: Could not detect interface"
    exit 1 
fi

echo "================================================"
echo "Network Configuration:"
echo "  Master Address: $MASTER_ADDR"
echo "  Master Port: $MASTER_PORT"
echo "  Socket Interface: $SOCKET_IFNAME"
echo "================================================"

# Export environmental variables 
export MASTER_ADDR=$MASTER_ADDR 
export MASTER_PORT=$MASTER_PORT

# To enable NCCL/RCCL Debug logs 
export NCCL_DEBUG=INFO
export NCCL_DEBUG_SUBSYS=INIT,NET

# Set Socket IF, if there are multiple interfaces for TCP communication 
export NCCL_SOCKET_IFNAME=$SOCKET_IFNAME

####################################
# srun export : export XDG_DATA_HOME, XDG_CACHE_HOME in the srun if multiple nodes are sharing the same working directory
# srun container-mount : 
#          /test_pytorch     : Contains distributed pytorch script
#          /etc/libibverbs.d : Contains libibverbs provider configuration files.
#          /usr/lib          : Contains RDMA userspace provider shared libraries 
####################################

# Run the distributed training
srun --unbuffered \
     --export=ALL,XDG_DATA_HOME=/tmp/xdg-\$SLURM_NODEID,XDG_CACHE_HOME=/tmp/xdg-cache-\$SLURM_NODEID \
     --container-image="$ENROOT_DATA_PATH/$IMAGE_NAME" \
     --container-mounts=$(pwd)/test_pytorch:/workspace,/etc/libibverbs.d:/etc/libibverbs.d,/usr/lib:/usr/lib \
     --container-workdir=/workspace \
     bash -c '
     export WORLD_SIZE=$SLURM_NTASKS
     export RANK=$SLURM_PROCID
     export LOCAL_RANK=$SLURM_LOCALID

     echo "[Rank $RANK on $SLURMD_NODENAME] Using $MASTER_ADDR:$MASTER_PORT on interface $NCCL_SOCKET_IFNAME"
     python3 -u distributed_pytorch.py
     '

echo "================================================"
echo "Job completed: $(date)"
echo "================================================"

