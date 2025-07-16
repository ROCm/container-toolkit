Release Notes
=============

This document provides an overview of the AMD Container Toolkit's release history, including new features, improvements, and compatibility information. The toolkit enables seamless integration of AMD GPUs into containerized environments, enhancing GPU utilization and management.

Compatibility Matrix
--------------------

.. list-table:: Compatibility Matrix
   :header-rows: 1

   * - AMD Container Toolkit
     - Docker Version
     - Supported OS
   * - 1.1.0
     - 25.0+ (``--gpus`` flag in 28.3.0+)
     - Ubuntu 22.04, Ubuntu 24.04, RHEL/CentOS 9
   * - 1.0.0
     - 25.0+
     - Ubuntu 22.04, Ubuntu 24.04

Versioning Information
----------------------

.. list-table:: Toolkit Versions
   :header-rows: 1

   * - Version
     - Release Date
     - Highlights
   * - v1.1.0
     - July 2025
     - GPU Partitioning Support , Docker --gpus Support, RHEL/CentOS 9 Support
   * - v1.0.0
     - June 2025
     - Initial Release with `amd-ctk`, CDI, Docker Integration, Ubuntu 22.04/24.04 Support
     
-------------------
v1.1.0 (July 2025)
-------------------

Overview
--------

Version 1.1.0 of the AMD Container Toolkit delivers a significant leap in flexibility, usability, and platform reach for GPU-accelerated container workloads. This release introduces three impactful features: **GPU Partitioning**, **support for RHEL/CentOS 9**, and **integration with Docker’s standardized ``--gpus`` flag**. These enhancements empower users to maximize GPU utilization, streamline deployment across diverse environments, and adopt industry-standard container interfaces for specifying GPU resources.

New Features
~~~~~~~~~~~~

- **GPU Partitioning Support**

  - ROCm-based GPUs can now be partitioned into multiple logical devices that are independently accessible from within containers.
  - Fine-grained control via `AMD_VISIBLE_DEVICES` environment variable.
  - Support for range-based device specification (e.g., `0-3,8,17-20`).
  - Compatibility with container runtimes using updated CDI specs.

- **Full Support for RHEL and CentOS 9**

  - This release introduces native RPM packaging and support for RHEL/CentOS 9 systems.

- **Support for `--gpus` Flag in Docker 28.x+**

  - Starting from Docker **28.3.0**, containerized GPU workloads can now utilize the standardized `--gpus` flag to request AMD GPUs directly in `docker run` commands.
  - Declarative selection of GPU resources without manual device path specification.
  - Improved integration with Docker's native GPU management features.

Improvements
------------

- **Range Operator for Device Selection:**  
  The `AMD_VISIBLE_DEVICES` environment variable now supports range expressions, making it easier to select multiple GPUs or partitions in a concise manner.
- **Documentation Updates:**  
  All documentation related to GPU partitioning, RHEL/CentOS installation, and Docker ``--gpus`` flag usage has been updated to reflect these new capabilities.

Known Issues
~~~~~~~~~~~~

- None reported for this release.

Upgrade Notes
-------------

- After any GPU partitioning changes, always regenerate and validate the CDI spec to ensure containers have access to the correct devices.
- For partitioned GPU visibility inside containers, ensure you are using ROCm version 6.4.1 or newer.
- For RHEL/CentOS 9, follow the new installation instructions in the documentation.
- To use the ``--gpus`` flag, upgrade Docker to version 28.3.0 or newer.

Next Steps
----------

1. Review the updated installation and partitioning guides in the documentation.
2. For GPU partitioning, use the new range operator and remember to regenerate CDI specs after changes.
3. For RHEL/CentOS 9, follow the new RPM-based installation workflow.
4. To use the ``--gpus`` flag, ensure you are running Docker 28.3.0 or newer.
5. Deploy your first container using:

   .. code-block:: bash

      docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi

For more information, refer to the full documentation.

-------------------
v1.0.0 (June 2025)
-------------------

Overview
--------

The initial release of the AMD Container Toolkit provided seamless GPU acceleration for containerized applications using AMD Instinct GPUs. Version 1.0.0 introduced robust Docker integration, CDI support, and developer tooling, making it easier than ever to deploy GPU-accelerated workloads in a containerized environment.

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

Known Issues
~~~~~~~~~~~~

- Partitioned GPUs were not supported.
- RPM builds were considered experimental.

Upgrade Notes
-------------

- Docker must be upgraded to version **25.0 or higher**.
- Ensure the ROCm driver version matches the compatibility matrix listed in `requirements.rst`.
- If migrating from NVIDIA, follow the steps outlined in `migration-guide.rst`.

Next Steps
----------

To get started:

1. Follow the installation steps in the `quick-start-guide.rst`.
2. Configure Docker using `amd-ctk configure runtime`.
3. Deploy your first container:

   .. code-block:: bash

      docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi

Additional details on GPU partitioning, runtime configuration, and CDI validation can be found in the Developer Guide.
