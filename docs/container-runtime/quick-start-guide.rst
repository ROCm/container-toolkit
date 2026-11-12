Quick Start
===========

This guide configures a container runtime for AMD GPUs and verifies GPU access in a container. Complete the :doc:`Installation <installation>` first.

The toolkit offers two ways to inject GPUs into containers:

- **CDI (recommended):** works with Docker, Podman, containerd, and other CDI-aware runtimes.
- **amd-container-runtime:** a Docker-only OCI runtime that selects GPUs with ``AMD_VISIBLE_DEVICES`` and supports the :doc:`GPU Tracker <gpu-tracker>`.

Option A: CDI
-------------

1. Generate and validate the CDI specification:

   .. code-block:: bash

      sudo amd-ctk cdi generate
      sudo amd-ctk cdi validate

2. On Docker 25.0 to 28.1, enable CDI in Docker. Docker 28.2 and newer enable it by default, so skip this step there:

   .. code-block:: bash

      sudo amd-ctk runtime configure
      sudo systemctl restart docker

3. Run a container with all AMD GPUs:

   .. code-block:: bash

      docker run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi monitor

To select specific GPUs or use Podman, nerdctl, or ctr, see :doc:`Running Workloads <running-workloads>`. For details on the specification, see the :doc:`CDI guide <cdi-guide>`.

Option B: AMD Container Runtime
-------------------------------

1. Register the AMD container runtime with Docker and restart the Docker daemon:

   .. code-block:: bash

      sudo amd-ctk runtime configure
      sudo systemctl restart docker

2. Run a container with access to all AMD GPUs:

   .. code-block:: bash

      docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/dev-ubuntu-24.04 amd-smi monitor

   The output lists every GPU on the host:

   .. code-block:: text

      GPU  POWER   GPU_T   MEM_T   GFX_CLK   GFX%   MEM%   ENC%   DEC%      VRAM_USAGE
        0  137 W   41 °C   36 °C   142 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        1  139 W   39 °C   33 °C   135 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        2  138 W   42 °C   34 °C   145 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        3  141 W   39 °C   33 °C   139 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        4  140 W   42 °C   36 °C   146 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        5  137 W   38 °C   33 °C   133 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        6  139 W   43 °C   36 °C   151 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        7  137 W   41 °C   34 °C   141 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB

3. Run a container with access to only the first GPU:

   .. code-block:: bash

      docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0 rocm/dev-ubuntu-24.04 amd-smi monitor

   The output lists only that GPU:

   .. code-block:: text

      GPU  POWER   GPU_T   MEM_T   GFX_CLK   GFX%   MEM%   ENC%   DEC%      VRAM_USAGE
        0  140 W   42 °C   36 °C   146 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB

To make the AMD container runtime the Docker default and drop ``--runtime=amd``, run ``sudo amd-ctk runtime configure --set-as-default`` and restart Docker. For GPU selection by range or UUID and the ``--gpus`` flag, see :doc:`Running Workloads <running-workloads>`.

Next Steps
----------

- :doc:`Running Workloads <running-workloads>` to select specific GPUs with Docker, Podman, nerdctl, or ctr.
- :doc:`Framework Integration <framework-integration>` to run PyTorch or TensorFlow.
- :doc:`GPU Partitioning <gpu-partitioning>` to run containers on GPU partitions.
- :doc:`Docker Compose <docker-compose>` and :doc:`Docker Swarm <docker-swarm>` for multi-container and multi-node deployments.
- :doc:`Troubleshooting <troubleshooting>` if a container does not see the expected GPUs.
