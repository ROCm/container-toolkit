Migration Guide: NVIDIA to AMD
==============================

- Migrating from the NVIDIA Container Toolkit to the AMD Container Toolkit is a streamlined process that enables developers to leverage AMD Instinct GPUs in containerized environments with minimal modifications. This guide provides step-by-step instructions to make this transition smooth and efficient.
- Migrating container workflows from NVIDIA GPUs (using `nvidia-docker`) to AMD Instinct GPUs (using AMD Container Toolkit) involves the following key changes:

Step 1: Environment Variable Updates
-------------------------------------
NVIDIA GPUs are typically managed through the environment variable ``NVIDIA_VISIBLE_DEVICES``. In the AMD Container Toolkit, this is replaced with:

.. code-block:: bash

   export AMD_VISIBLE_DEVICES=all  # To enable all GPUs

   export AMD_VISIBLE_DEVICES=0,1  # To enable specific GPUs

The variable syntax remains consistent, but the prefix changes to ``AMD``. This environment variable is recognized by the AMD runtime to expose GPUs to your container workloads.

Step 2: Update Runtime Configuration
-------------------------------------
NVIDIA's container runtime is identified as ``nvidia`` in Docker commands. For AMD, the runtime flag needs to be updated to:

.. code-block:: bash

   sudo docker run --rm --runtime=amd <image-name>

To set AMD as the default runtime:

.. code-block:: bash

   sudo amd-ctk runtime configure --runtime=docker --set-as-default

If you previously used:

.. code-block:: bash

   sudo docker run --rm --runtime=nvidia <image-name>

You would now use the equivalent:

.. code-block:: bash

   sudo docker run --rm --runtime=amd <image-name>

Step 3: Command Line Utility Replacement
-----------------------------------------
NVIDIA provides ``nvidia-ctk`` for container configurations and runtime settings. In the AMD ecosystem, this is replaced with:

.. code-block:: bash

   amd-ctk

For example, listing all available GPUs using NVIDIA would look like:

.. code-block:: bash

   nvidia-ctk list

The AMD equivalent:

.. code-block:: bash

   amd-ctk list

You can also generate CDI specifications with:

.. code-block:: bash

   amd-ctk cdi generate --output=/etc/cdi/amd.json

Step 4: Container Images
------------------------
Containers built for NVIDIA GPUs often rely on CUDA-based images. AMD's container toolkit is designed to work seamlessly with ROCm-enabled images:

- TensorFlow: ``tensorflow/tensorflow:rocm-latest``
- PyTorch: ``rocm/pytorch:latest``
- Triton Inference Server: ``rocm/tritonserver:latest``

For example, running PyTorch with AMD GPUs:

.. code-block:: bash

   sudo docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/pytorch:latest

Step 5: Framework Adjustments
-----------------------------
To fully leverage AMD Instinct GPUs, frameworks like TensorFlow and PyTorch must use their ROCm-enabled versions. This ensures compatibility and optimized performance for machine learning workloads.

Compatibility Notes
-------------------
- Ensure Docker is version 25 or above for CDI compatibility.
- Some CUDA-specific applications may require minor modifications.
- ROCm supports a majority of ML frameworks, but always validate with your application stack.

Comparison:
-----------
.. list-table:: Feature Comparison: NVIDIA Docker vs AMD Container Toolkit
   :header-rows: 1
   :widths: 20 40 40

   * - **Feature**
     - **NVIDIA Docker**
     - **AMD Container Toolkit**
   * - GPU Enumeration
     - ``nvidia-smi`` - Lists available GPUs and their statuses.
     - ``rocm-smi`` - Lists AMD GPUs and exposes detailed hardware information.
   * - Container Runtime
     - ``nvidia-container-runtime`` - Manages container interactions with NVIDIA GPUs.
     - ``amd-container-runtime`` - Integrates AMD Instinct GPUs seamlessly with Docker.
   * - Environment Variable
     - ``NVIDIA_VISIBLE_DEVICES`` - Specifies which NVIDIA GPUs are visible inside containers.
     - ``AMD_VISIBLE_DEVICES`` - Specifies which AMD GPUs are visible inside containers.
   * - Framework Images
     - NVIDIA-specific images optimized for CUDA.
     - ROCm-optimized images designed for AMD GPUs.
   * - TensorFlow Support
     - CUDA TensorFlow - Supports TensorFlow operations on NVIDIA GPUs.
     - ROCm TensorFlow - Optimized TensorFlow builds for AMD GPUs.
   * - PyTorch Support
     - CUDA PyTorch - Optimized for NVIDIA architectures.
     - ROCm PyTorch - Optimized for AMD Instinct architectures.
   * - Configuration Toolkit
     - ``nvidia-ctk`` - NVIDIA's CLI for runtime configuration.
     - ``amd-ctk`` - AMD's CLI for Docker runtime integration and device management.
   * - Default Docker Runtime
     - ``nvidia runtime`` - Configures Docker to use NVIDIA GPUs by default.
     - ``amd runtime`` - Configures Docker to use AMD GPUs by default.


Testing and Validation
-----------------------
After migration, it is crucial to validate workloads with:

.. code-block:: bash

   sudo docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi

Ensure that all intended GPUs are detected and functioning as expected.


Next Steps
----------
Once migration is complete:

- Update your CI/CD pipelines to reflect runtime changes.
- Adjust Dockerfiles if specific runtime flags were set for NVIDIA.
- Monitor GPU usage using tools like ``rocm-smi`` and ``amd-ctk list``.

This completes the NVIDIA to AMD migration process, enabling you to leverage the full power of AMD Instinct GPUs in containerized workflows.
By following this migration guide, users can rapidly transition their GPU workloads to AMD Instinct platforms.
