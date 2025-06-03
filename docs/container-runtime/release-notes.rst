Release Notes
=============

Overview
--------

The AMD Container Toolkit provides seamless GPU acceleration for containerized applications using AMD Instinct GPUs. This release introduces significant improvements in Docker integration, CDI support, and developer tooling, making it easier than ever to deploy GPU-accelerated workloads in a containerized environment.

Versioning Information
----------------------

Current Version: **1.0.0**

- Docker Compatibility: **25.0+**
- Supported Distributions: Ubuntu 22.04, Ubuntu 24.04

New Features
------------

- **CDI (Container Device Interface) Support:**
  - Full support for CDI-based GPU enumeration and allocation in Docker containers.
  - Simplified integration with Kubernetes CDI configurations.

- **Enhanced Docker Runtime Integration:**
  - The AMD Container Runtime (`amd-container-runtime`) is now fully aligned with Docker 25.0+, ensuring smooth device injection and runtime detection.

- **Command Line Utility (`amd-ctk`):** 
  Introduction of `amd-ctk`, a CLI tool for:

  - Listing GPU devices.
  - Generating CDI specs.
  - Configuring Docker runtime.

- **Simplified Installation Flow:**
  - Reduced installation steps with clear dependency management for Ubuntu systems.

Improvements
------------

- Optimized Docker Daemon Configuration
  - Improved detection of AMD GPUs in `/etc/docker/daemon.json`.
  - Easier integration with Docker Compose and multi-container setups.

- Better Log Management
  - All runtime logs are now centralized in `/var/log/amd-container-runtime.log` for easy access and troubleshooting.

- Enhanced GPU Discovery
  - Faster and more reliable device discovery through `amd-ctk cdi list`.

Upgrade Notes
-------------

- Docker must be upgraded to version **25.0 or higher**.
- Ensure the ROCm driver version matches the compatibility matrix listed in `requirements.rst`.
- If migrating from NVIDIA, follow the steps outlined in `migration-guide.rst`.

Compatibility Matrix
--------------------

.. list-table:: Compatibility Matrix
  :header-rows: 1

  * - AMD Container Toolkit
    - Docker Version
  * - 1.0.0
    - 25.0+

Next Steps
----------

To get started:

1. Follow the installation steps in the `quick-start-guide.rst`.
2. Configure Docker using `amd-ctk configure runtime`.
3. Deploy your first container using:

   .. code-block:: bash

      docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi

For more information, please refer to the documentation section.
