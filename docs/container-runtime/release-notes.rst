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
   * - 1.2.0
     - 25.0+ 
     - Ubuntu 22.04, Ubuntu 24.04, RHEL 9.5
   * - 1.1.0
     - 25.0+ (``--gpus`` flag in 28.3.0+)
     - Ubuntu 22.04, Ubuntu 24.04, RHEL 9.5
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
   * - v1.2.0
     - November 2025
     - GPU Tracker feature support, Docker Swarm Support
   * - v1.1.0
     - July 2025
     - GPU Partitioning Support , Docker --gpus Support, RHEL 9.5 Support
   * - v1.0.0
     - June 2025
     - Initial Release with `amd-ctk`, CDI, Docker Integration, Ubuntu 22.04/24.04 Support

-------------------
v1.2.0 (November 2025)
-------------------

Overview
--------

Version 1.2.0 of the AMD Container Toolkit introduces two major enhancements aimed at improving GPU visibility and orchestration flexibility in containerized environments:

- **GPU Tracker:** A new monitoring utility for real-time tracking of GPU usage across containers.
- **Docker Swarm Support:** Native integration with Docker Swarm for orchestrating AMD GPU workloads at scale.

New Features
~~~~~~~~~~~~

- **GPU Tracker**

   - GPU Tracker is an extremely lightweight feature of AMD Container Toolkit that allows you to track access of GPUs in containers.
   - GPU Tracker provides CLIs that can be used to control the accessibility of GPUs in containers. The accessibility of GPUs can be set to either `shared` or `exclusive`.

- **Docker Swarm Support**

   - Allows users to deploy and manage GPU-accelerated containers across a cluster instead of being limited to a single host.
   - Uses GPU UUIDs for accurate resource mapping and scheduling, ensuring workloads run on specific GPUs.

Known Issues
~~~~~~~~~~~~

- None reported for this release.

Upgrade Notes
-------------

-  GPU Tracker feature is currently supported only if containers are started using the `docker run` command and GPUs are made accessible in containers using the `AMD_VISIBLE_DEVICES` environment variable. If containers are started and granted access to GPUs in any other manner, GPU Tracker feature is not supported.

-------------------
v1.1.0 (July 2025)
-------------------

Overview
--------

Version 1.1.0 of the AMD Container Toolkit delivers a significant leap in flexibility, usability, and platform reach for GPU-accelerated container workloads. This release introduces three impactful features: **GPU Partitioning**, **support for RHEL 9.5**, and **integration with Docker’s standardized ``--gpus`` flag**. These enhancements empower users to maximize GPU utilization, streamline deployment across diverse environments, and adopt industry-standard container interfaces for specifying GPU resources.

New Features
~~~~~~~~~~~~

- **GPU Partitioning Support**

  - ROCm-based GPUs can now be partitioned into multiple logical devices that are independently accessible from within containers.
  - Fine-grained control via `AMD_VISIBLE_DEVICES` environment variable.
  - Support for range-based device specification (e.g., `0-3,8,17-20`).
  - Compatibility with container runtimes using updated CDI specs.

- **Full Support for RHEL 9.5**

  - This release introduces native RPM packaging and support for RHEL 9.5 systems.

- **Support for `--gpus` Flag in Docker 28.x+**

  - Starting from Docker **28.3.0**, containerized GPU workloads can now utilize the standardized `--gpus` flag to request AMD GPUs directly in `docker run` commands.
  - Declarative selection of GPU resources without manual device path specification.
  - Improved integration with Docker's native GPU management features.

Improvements
------------

- **Range of GPU Device Selection:**  
  The `AMD_VISIBLE_DEVICES` environment variable allows users to specify range of GPUs, making it easier to select multiple GPUs or partitions in a concise manner.
- **Documentation Updates:**  
  All documentation related to GPU partitioning, RHEL installation, and Docker ``--gpus`` flag usage has been updated to reflect these new capabilities.

Known Issues
~~~~~~~~~~~~

- None reported for this release.

Upgrade Notes
-------------

- After any GPU partitioning changes, always regenerate and validate the CDI spec to ensure containers have access to the correct devices.
- For partitioned GPU visibility inside containers, ensure you are using ROCm version 6.4.1 or newer.
- For RHEL 9.5, follow the new installation instructions in the documentation.
- To use the ``--gpus`` flag, upgrade Docker to version 28.3.0 or newer.

Next Steps
----------

1. Review the updated requirements and quick start guide in the documentation.
2. For GPU partitioning, users can provide range of GPUs and remember to regenerate CDI specs after changes.
3. For RHEL 9.5, follow the new RPM-based installation workflow.
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
- Ensure the amdgpu driver version matches the compatibility matrix listed in `requirements.rst`.
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
