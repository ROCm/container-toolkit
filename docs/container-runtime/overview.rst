Overview
========

- The AMD Container Toolkit provides a robust and flexible framework to streamline the use of AMD Instinct GPUs with containerized applications.
- It simplifies GPU access within Docker environments, enhances device discovery, and enables better integration with modern container technologies.
- The toolkit is designed to work seamlessly with the ROCm software stack, allowing developers to leverage the full power of AMD GPUs for high-performance computing, machine learning, and other GPU-accelerated workloads.
- The AMD Container Toolkit architecture integrates directly with the Docker daemon to manage GPU resources seamlessly.


The toolkit consists of two primary components:

- **amd-container-runtime**: A custom container runtime (wrapper around ``runc``) for injecting AMD GPUs into container specifications.
- **amd-ctk (Container Toolkit CLI)**: A command-line utility for managing GPU configurations, runtime settings, and container orchestration integrations.

Key Benefits:
-------------

- Seamless GPU access in containers with minimal configuration.
- Simplified device management and discovery through CDI and environment variables.
- Smooth integration with popular containerized machine learning, HPC, and data science frameworks.
- Enables efficient image building and development workflows for AMD GPUs.

Use Cases:
-----------

- Building machine learning applications with GPU acceleration.
- Running ROCm-compatible containerized workloads.
- Rapid experimentation in development environments.

Core Concepts
=============

Architecture Overview
----------------------
The AMD Container Toolkit sits between Docker and the Linux container runtime, enabling GPU access by modifying the OCI specification before handing off to ``runc``.

Visual Flow:

.. code-block:: text

   +------------+                       +---------------+             +-----------------------+
   |            |                       |               |             |                       |
   | Docker CLI |  AMD_VISIBLE_DEVICES  | Docker Daemon |  OCI SPEC   | AMD Container Runtime |
   |            |          -->          |               |     -->     |                       |
   +------------+                       +---------------+             +-----------------------+
                                                                                  |
                                                                            UPDATED OCI SPEC
                                                                          INJECTED GPU DEVICES
                                                                                  |
                                                                                  v
                                                                       +----------------------+
                                                                       |         RUNC         +
                                                                       +----------------------+
                                                                                  |
                                                                                  v
                                                                       +----------------------+
                                                                       |   CONTAINER PROCESS  +
                                                                       +----------------------+
                                                                                  |
                                                                                  v
                                                                       +----------------------+
                                                                       |      GPU DRIVER      +
                                                                       +----------------------+

Docker integration with amd-container-runtime
--------------------------------------------

For installation, configuration, and running GPU workloads with Docker using the amd-container-runtime, see the :doc:`Quick Start Guide <quick-start-guide>`.

Using CDI for GPU injection
---------------------------

The toolkit also supports the Container Device Interface (CDI) for GPU injection. To set up and use CDI to run workloads, see the :doc:`CDI guide <cdi-guide>` and :doc:`Running Workloads <running-workloads>`.
