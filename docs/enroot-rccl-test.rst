Running RCCL Tests on AMD GPUs with Pollara AINIC

https://github.com/ROCm/container-toolkit/blob/main/tests/enroot/batch_scripts/rccl_tests_sbatch.sh

Overview

This document describes how to run RCCL (ROCm Communication Collectives
Library) performance tests on AMD GPUs using Pollara AINIC networking in
a Slurm cluster environment with Pyxis/Enroot containerization.

Prerequisites

Slurm cluster with Pyxis and Enroot plugins installed
AMD GPUs with ROCm support
Pollara AINIC network adapters
Access to rocm/roce-workload Docker images

The host AINIC firmware version should match the AINIC version of the
docker image

Script Configuration
Resource Allocation
Bash
#SBATCH –nodes=2 # Number of compute nodes
#SBATCH –ntasks=2 # Total number of MPI tasks
#SBATCH –ntasks-per-node=1 # One task per node
#SBATCH –gpus-per-task=8 # 8 GPUs per task (16 GPUs total)

Container Image

The script uses a containerized environment with:
Default Image:
ubuntu24_rocm-7.0.2_rccl-7.0.2_anp-v1.2.0_ainic-1.117.5-a-56
Customization: Override by setting DOCKER_IMAGE_VERSION environment
variable

GPU Configuration

HSA_NO_SCRATCH_RECLAIM=1 - Prevents the HSA runtime from reclaiming
scratch memory between kernel launches, improving performance by
avoiding memory reallocation overhead

Network Configuration

NCCL_SOCKET_IFNAME=ens50f0 - Specifies which network interface RCCL
should use for inter-node communication (replace with your actual
interface name)

NCCL_IB_DISABLE=0 - Enables InfiniBand or RoCE (RDMA over Converged
Ethernet) support for high-performance network communication

NCCL_IB_GID_INDEX=1 - Sets the Global Identifier (GID) index for RoCE
communication; index 1 typically corresponds to RoCEv2 (IP-routable)

IONIC_LOCKFREE=all - Enables lock-free optimizations for Pollara AINIC
network adapters, reducing synchronization overhead

Performance Optimizations

NCCL_GDRCOPY_ENABLE=0 - Disables GPUDirect RDMA Copy library; may be
necessary if gdrcopy is unavailable or causing compatibility issues

NCCL_GDR_FLUSH_DISABLE=1 - Skips GPU memory flush operations after RDMA
writes, improving latency

NCCL_IB_USE_INLINE=1 - Sends small messages inline within the InfiniBand
work request, reducing latency for small transfers

NCCL_IB_QPS_PER_CONNECTION=1 - Sets the number of Queue Pairs per
connection; using 1 reduces memory overhead while maintaining
performance

NCCL_IB_TC=96 - Configures the InfiniBand traffic class for standard
RDMA operations, enabling QoS and lossless operation

NCCL_IB_FIFO_TC=184 - Sets the traffic class for FIFO (control)
messages, separating control and data plane traffic

MPI Configuration

PMIX_MCA_gds=hash - Configures PMIx Global Data Store to use hash-based
storage, improving job launch performance in large-scale deployments

OMPI_MCA_btl=^openib - Excludes the legacy OpenIB transport layer from
OpenMPI; RCCL handles InfiniBand communication directly

OMPI_MCA_btl_tcp_if_exclude=lo,docker0 - Prevents OpenMPI from using
loopback and Docker network interfaces, avoiding routing conflicts

Usage

Basic Execution
Bash
sbatch rccl_test.sbatch
Custom Container Version
Bash
DOCKER_IMAGE_VERSION=“custom_version” sbatch rccl_test.sbatch

Test Parameters

The script runs the all_reduce_perf benchmark with:
-b 16 - Minimum message size: 16 bytes
-e 8G - Maximum message size: 8 GB
-f 2 - Size multiplication factor: 2x
-g 8 - Number of GPUs per task: 8

Output

Standard Output: logs/rccl_test\_.out
Standard Error: logs/rccl_test\_.err
Topology Dump: /tmp/topo_all.txt (inside container)
Graph Dump: /tmp/graph_all.txt (inside container)

Container Management

The script automatically:
Creates a logs/ directory for output in the submission directory on the
head node
Checks if the container image exists locally
Downloads and converts the Docker image to Enroot format (.sqsh) if
needed
Reuses existing container images to save time

Notes

Container images are cached in the working directory
First run will download the container (~several GB)
Subsequent runs reuse the cached container
Ensure sufficient disk space in the working directory for container
images
