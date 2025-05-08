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

Docker Runtime Integration
---------------------------

To use AMD GPUs with Docker:

1. Configure Docker runtime:

.. code-block:: bash

   sudo amd-ctk configure runtime

2. Restart Docker:

.. code-block:: bash

   sudo systemctl restart docker

3. Usage:

- Environment variable ``AMD_VISIBLE_DEVICES``:

  .. code-block:: bash

     sudo docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi

- CDI style:

  .. code-block:: bash

     amd-ctk cdi generate --output=/etc/cdi/amd.json
     sudo docker run --rm --device amd.com/gpu=all rocm/rocm-terminal rocm-smi

Device Discovery
----------------

Enumerate GPUs:
Outputs a list of available GPUs in CDI-compliant format.

.. code-block:: bash

   amd-ctk cdi list
