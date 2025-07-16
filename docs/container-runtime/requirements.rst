System Requirements
====================

Before installing or using the AMD Container Toolkit, ensure your environment meets the following prerequisites:

Supported Operating Systems
---------------------------
- Ubuntu 22.04 LTS (Jammy Jellyfish)
- Ubuntu 24.04 LTS (Noble Numbat)
- RHEL 9 / CentOS Stream 9

.. note::
   - RHEL and CentOS 9 support is new in v1.1.0 and requires the RPM-based installation flow.
   - Follow the installation instructions in the documentation for these platforms.

Docker Compatibility
--------------------
- Docker 25.0+ is required for all features.
- Docker 28.3.0+ is required to use the standardized ``--gpus`` flag for AMD GPU selection.

.. note::
   - Docker version **28.3.0** or later is recommended to take advantage of the `--gpus` flag integration for AMD GPUs.
   - Ensure ROCm version **6.4.1** or higher is used in container images to support GPU partitioning capabilities.

ROCm and Driver Compatibility
-----------------------------
- ROCm 6.4.1 or newer is required to view and verify partitioned GPUs inside containers.

Note
----
A mismatch between ROCm and driver versions may lead to runtime failures.

System Prerequisites
---------------------

The following packages and configurations are required on the host system:

- **Kernel Headers** and **Extra Kernel Modules** for your running kernel.
- **Docker** (preferably installed via the `docker.io` package or Docker’s official repositories).
- **User Permissions**:
  - The user running containers must belong to the `render` and `video` groups.
  - Example:

    .. code-block:: bash

       sudo usermod -aG render,video $USER
       newgrp render && newgrp video
- For RHEL/CentOS 9, ensure the AMD Container Toolkit YUM repository is configured as described in the installation guide.

GPU Partitioning Requirements
-----------------------------
- To use GPU partitioning, ensure your ROCm version is 6.4.1 or newer.
- After any partitioning change, you must regenerate the CDI spec to reflect the new GPU topology:
- The `amd-smi` tool can be used to inspect partition status and details from within containers.

Important Notes
----------------

- ROCm must be installed on the host system and must match the expected version compatibility with your container images.
- Using mismatched ROCm driver and runtime versions may result in runtime errors or undefined behavior.
- Ensure CDI specs are kept up to date in environments where GPU topology can change frequently (e.g., partitioned systems or multi-GPU deployments).
- Always refer to the latest documentation for platform-specific installation and configuration steps.

Failure to meet these requirements may result in incomplete functionality or runtime errors. For troubleshooting and advanced configuration, consult the relevant sections of the documentation.
