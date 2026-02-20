#!/bin/bash
#SBATCH --nodes=2
#SBATCH --ntasks=16
#SBATCH --ntasks-per-node=8
#SBATCH --gpus-per-task=1
#SBATCH --gpu-bind=closest
#SBATCH --cpus-per-task=24
#SBATCH --distribution=block:block
#SBATCH --job-name=rccl_test
#SBATCH --output=logs/rccl_test_%j.out
#SBATCH --error=logs/rccl_test_%j.err

set -e

mkdir -p logs

# Configurable number of iterations with default value
NUM_ITERATIONS=${NUM_ITERATIONS:-10}

# Customizable container image with default value
DOCKER_IMAGE_VERSION=${DOCKER_IMAGE_VERSION:-"ubuntu24_rocm-7.0.2_rccl-7.0.2_anp-v1.2.0_ainic-1.117.5-a-56"}

CONTAINER_IMAGE="enroot_rccl-$DOCKER_IMAGE_VERSION.sqsh"

echo "Pulling container image for version: $DOCKER_IMAGE_VERSION and saving to $CONTAINER_IMAGE"

# Pull the image on every allocated node
srun --ntasks=2 --ntasks-per-node=1 bash -c "
   if [ ! -f \"$PWD/$CONTAINER_IMAGE\" ]; then
       echo \"Node \$(hostname): Pulling container image and saving to $CONTAINER_IMAGE\"
       if ! enroot import -o \"$PWD/$CONTAINER_IMAGE\" \"docker://rocm/roce-workload:$DOCKER_IMAGE_VERSION\"; then
           echo \"Node \$(hostname): Failed to pull container image\" >&2
           exit 1
       fi
   else
       echo \"Node \$(hostname): Container image already exists, skipping pull\"
   fi
"
# Export environment variables
export HSA_NO_SCRATCH_RECLAIM=1
export NCCL_SOCKET_IFNAME=ens50f0
export NCCL_IB_DISABLE=0
export PMIX_MCA_gds=hash
export OMPI_MCA_btl_tcp_if_exclude=lo,docker0
export IONIC_LOCKFREE=all
export NCCL_GDRCOPY_ENABLE=0
export NCCL_GDR_FLUSH_DISABLE=1
export NCCL_GRAPH_DUMP_FILE=/tmp/graph_all.txt
export NCCL_IB_GID_INDEX=1
export NCCL_PXN_DISABLE=0
export NCCL_IB_QPS_PER_CONNECTION=1
export NCCL_IB_TC=96
export NCCL_IB_FIFO_TC=184
export NCCL_IGNORE_CPU_AFFINITY=1
export NCCL_IB_USE_INLINE=1
export NCCL_NET_OPTIONAL_RECV_COMPLETION=0
export NCCL_TOPO_DUMP_FILE=/tmp/topo_all.txt
export RCCL_GDR_FLUSH_GPU_MEM_NO_RELAXED_ORDERING=0
export OMPI_MCA_btl=^openib
export UCX_UNIFIED_MODE=y
export RCCL_AINIC_ROCE=1
export RCCL_LL128_FORCE_ENABLE=1
export NCCL_IB_PCI_RELAXED_ORDERING=1
export NCCL_NET_PLUGIN=/root/amd-anp/build/librccl-anp.so
export NCCL_DEBUG=VERSION

# Run the test multiple times
echo "Running all_reduce_perf test $NUM_ITERATIONS times"
for i in $(seq 1 $NUM_ITERATIONS); do
    echo "=========================================="
    echo "Starting iteration $i of $NUM_ITERATIONS"
    echo "=========================================="

    srun --mpi=pmix \
        --container-image="$PWD/$CONTAINER_IMAGE" \
        /root/rccl-tests/build/all_reduce_perf -b 1K -e 16G -f 2 -g 1 -n 20 -w 5 -c 1

    echo "Completed iteration $i of $NUM_ITERATIONS"
    echo ""
done

echo "All $NUM_ITERATIONS iterations completed successfully"