Running RCCL Tests on AMD GPUs with Pollara AINIC
==================================================

https://github.com/ROCm/container-toolkit/blob/main/tests/enroot/batch_scripts/rccl_tests_sbatch.sh

Overview
--------

This document describes how to run RCCL (ROCm Communication Collectives Library) performance tests on AMD GPUs using Pollara AINIC networking in a Slurm cluster environment with Pyxis/Enroot containerization.

Prerequisites
-------------

- Slurm cluster with Pyxis and Enroot plugins installed
- AMD GPUs with ROCm support
- Pollara AINIC network adapters
- Access to rocm/roce-workload Docker images

The host AINIC firmware version should match the AINIC version of the docker image.

Script Configuration
--------------------

Resource Allocation
~~~~~~~~~~~~~~~~~~

.. code-block:: bash

   #SBATCH --nodes=2                  # Number of compute nodes
   #SBATCH --ntasks=16                # Total number of MPI tasks
   #SBATCH --ntasks-per-node=8        # 8 tasks per node
   #SBATCH --gpus-per-task=1          # 1 GPU per task (16 GPUs total)
   #SBATCH --gpu-bind=closest         # Bind GPUs to closest CPU cores
   #SBATCH --cpus-per-task=24         # 24 CPU cores per task
   #SBATCH --distribution=block:block # Block distribution for tasks and CPUs

Container Image
~~~~~~~~~~~~~~

The script uses a containerized environment with:

- **Default Image:** ``ubuntu24_rocm-7.0.2_rccl-7.0.2_anp-v1.2.0_ainic-1.117.5-a-56``
- **Customization:** Override by setting ``DOCKER_IMAGE_VERSION`` environment variable

Test Iterations
~~~~~~~~~~~~~~

The batch script runs **multiple iterations** of the RCCL all_reduce_perf benchmark based on the ``NUM_ITERATIONS`` variable (default: **10**). Each iteration executes the benchmark and outputs performance metrics to the log file.

Performance Validation
~~~~~~~~~~~~~~~~~~~~~

The performance validation is performed by the test suite script (``test_enroot.py``), which:

1. Parses the output from all iterations in the log file
2. Extracts the bus bandwidth measurements from each iteration
3. Calculates the **average bus bandwidth** across all iterations
4. Compares the average bandwidth against a predefined threshold value of **130 GB/s**
5. **Fails the test** if the measured average bandwidth is more than **5% below the threshold** (i.e., < 123.5 GB/s)

This validation ensures consistent performance across multiple runs and helps identify performance degradation or configuration issues.

Environment Variables
---------------------

GPU Configuration
~~~~~~~~~~~~~~~~

``HSA_NO_SCRATCH_RECLAIM=1``
    Prevents the HSA runtime from reclaiming scratch memory between kernel launches, improving performance by avoiding memory reallocation overhead

Network Configuration
~~~~~~~~~~~~~~~~~~~~

``NCCL_SOCKET_IFNAME=ens50f0``
    Specifies which network interface RCCL should use for inter-node communication (replace with your actual interface name)

``NCCL_IB_GID_INDEX=1``
    Sets the Global Identifier (GID) index for RoCE communication; index 1 typically corresponds to RoCEv2 (IP-routable)

Performance Optimizations
~~~~~~~~~~~~~~~~~~~~~~~~

``NCCL_GDRCOPY_ENABLE=0``
    Disables GPUDirect RDMA Copy library; may be necessary if gdrcopy is unavailable or causing compatibility issues

``NCCL_GDR_FLUSH_DISABLE=1``
    Skips GPU memory flush operations after RDMA writes, improving latency

``NCCL_IB_USE_INLINE=1``
    Sends small messages inline within the InfiniBand work request, reducing latency for small transfers

``NCCL_IB_QPS_PER_CONNECTION=1``
    Sets the number of Queue Pairs per connection; using 1 reduces memory overhead while maintaining performance

``NCCL_IB_TC=96``
    Configures the InfiniBand traffic class for standard RDMA operations, enabling QoS and lossless operation

``NCCL_IB_FIFO_TC=184``
    Sets the traffic class for FIFO (control) messages, separating control and data plane traffic

``NCCL_IGNORE_CPU_AFFINITY=1``
    Allows NCCL to ignore CPU affinity settings, useful when running with custom task distributions

``NCCL_NET_OPTIONAL_RECV_COMPLETION=0``
    Disables optional receive completion optimization, ensuring all receive operations complete before proceeding

``RCCL_GDR_FLUSH_GPU_MEM_NO_RELAXED_ORDERING=0``
    Controls GPU memory flush behavior without relaxed ordering for GPUDirect RDMA operations

``UCX_UNIFIED_MODE=y``
    Enables unified mode for UCX (Unified Communication X) transport layer, optimizing memory handling

``RCCL_AINIC_ROCE=1``
    Enables AINIC-specific RoCE optimizations for RCCL communication

``RCCL_LL128_FORCE_ENABLE=1``
    Forces the use of LL128 (Low Latency 128-byte) protocol for improved small message performance

``NCCL_IB_PCI_RELAXED_ORDERING=1``
    Enables PCI relaxed ordering for InfiniBand operations, improving throughput on supported hardware

``NCCL_NET_PLUGIN=/root/amd-anp/build/librccl-anp.so``
    Specifies the path to the AMD Network Plugin library for RCCL

``NCCL_DEBUG=VERSION``
    Sets NCCL debug output to show version information

Topology and Graph Dumps
~~~~~~~~~~~~~~~~~~~~~~~~

``NCCL_TOPO_DUMP_FILE=/tmp/topo_all.txt``
    Specifies the file path where NCCL will dump the detected network topology information

MPI Configuration
~~~~~~~~~~~~~~~~

``PMIX_MCA_gds=hash``
    Configures PMIx Global Data Store to use hash-based storage, improving job launch performance in large-scale deployments

``OMPI_MCA_btl=^openib``
    Excludes the legacy OpenIB transport layer from OpenMPI; RCCL handles InfiniBand communication directly

``OMPI_MCA_btl_tcp_if_exclude=lo,docker0``
    Prevents OpenMPI from using loopback and Docker network interfaces, avoiding routing conflicts

Usage
-----

Basic Execution
~~~~~~~~~~~~~~

.. code-block:: bash

   sbatch rccl_tests_sbatch.sh

Custom Container Version
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

   DOCKER_IMAGE_VERSION="custom_version" sbatch rccl_tests_sbatch.sh

Custom Number of Iterations
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

   NUM_ITERATIONS=20 sbatch rccl_tests_sbatch.sh

Test Parameters
---------------

The script runs the ``all_reduce_perf`` benchmark with:

- ``-b 1K`` - Minimum message size: 1 KB
- ``-e 16G`` - Maximum message size: 16 GB
- ``-f 2`` - Size multiplication factor: 2x
- ``-g 1`` - Number of GPUs per task: 1
- ``-n 20`` - Number of iterations per message size
- ``-w 5`` - Number of warmup iterations
- ``-c 1`` - Check correctness of results

Output
------

- **Standard Output:** ``logs/rccl_test_<job_id>.out``
- **Standard Error:** ``logs/rccl_test_<job_id>.err``
- **Topology Dump:** ``/tmp/topo_all.txt`` (inside container)

Container Management
--------------------

The script automatically:

- Creates a ``logs/`` directory for output in the submission directory on the head node
- Checks if the container image exists locally
- Downloads and converts the Docker image to Enroot format (.sqsh) if needed
- Reuses existing container images to save time

Notes
-----

- Container images are cached in the working directory
- First run will download the container (~several GB)
- Subsequent runs reuse the cached container
- Ensure sufficient disk space in the working directory for container images
