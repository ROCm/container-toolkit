Migrating from NVIDIA
=====================

Migrating container workflows from the NVIDIA Container Toolkit to the AMD Container Toolkit mostly means renaming devices, runtimes, and environment variables, and switching to ROCm-enabled images. Install the toolkit first as described in :doc:`Installation <installation>`.

Step 1: CDI Device Names
------------------------

If you use CDI with NVIDIA GPUs, generate the AMD CDI specification and change the device kind from ``nvidia.com/gpu`` to ``amd.com/gpu``:

.. code-block:: bash

   sudo amd-ctk cdi generate

.. code-block:: diff

   - docker run --rm --device nvidia.com/gpu=all <image-name>
   + docker run --rm --device amd.com/gpu=all <image-name>

CDI works with Docker, Podman, and containerd. Docker 25.0 to 28.1 need CDI enabled in Docker first; see :doc:`System Requirements <requirements>`.

Step 2: Runtime and Environment Variable
----------------------------------------

If you use the NVIDIA runtime with Docker, register the AMD container runtime and switch the runtime name and environment variable. Unlike ``nvidia-ctk runtime configure``, ``amd-ctk runtime configure`` supports Docker only; for containerd or CRI-O, use CDI (Step 1).

.. code-block:: bash

   sudo amd-ctk runtime configure
   sudo systemctl restart docker

.. code-block:: diff

   - docker run --rm --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=0,1 <image-name>
   + docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0,1 <image-name>

``AMD_VISIBLE_DEVICES`` accepts ``all``, GPU indices, ranges such as ``0-3``, and GPU UUIDs. To set AMD as the default runtime:

.. code-block:: bash

   sudo amd-ctk runtime configure --set-as-default
   sudo systemctl restart docker

Step 3: Command Line Utility
----------------------------

``nvidia-ctk`` is replaced with ``amd-ctk``:

.. code-block:: diff

   - nvidia-ctk cdi list
   + amd-ctk cdi list

To list GPUs with their UUIDs, use:

.. code-block:: bash

   amd-ctk gpu list

Step 4: Container Images and Frameworks
---------------------------------------

Containers built for NVIDIA GPUs often rely on CUDA-based images. Switch to ROCm-enabled images, for example:

- TensorFlow: ``rocm/tensorflow:latest``
- PyTorch: ``rocm/pytorch:latest``

For example, running PyTorch with AMD GPUs:

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all rocm/pytorch:latest

Some CUDA-specific applications need minor modifications. ROCm supports most ML frameworks, but always validate with your application stack.

Comparison
----------

.. list-table:: Feature Comparison: NVIDIA Container Toolkit vs AMD Container Toolkit
   :header-rows: 1
   :widths: 20 40 40

   * - **Feature**
     - **NVIDIA Container Toolkit**
     - **AMD Container Toolkit**
   * - GPU Enumeration
     - ``nvidia-smi``
     - ``amd-smi``, ``rocm-smi``, or ``amd-ctk gpu list``
   * - CDI Device Kind
     - ``nvidia.com/gpu``
     - ``amd.com/gpu``
   * - Container Runtime
     - ``nvidia-container-runtime``
     - ``amd-container-runtime``
   * - Environment Variable
     - ``NVIDIA_VISIBLE_DEVICES``
     - ``AMD_VISIBLE_DEVICES``
   * - Configuration CLI
     - ``nvidia-ctk``
     - ``amd-ctk``
   * - Framework Images
     - CUDA-based images
     - ROCm-based images

Testing and Validation
----------------------

After migration, validate that the intended GPUs are visible:

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi

Then update your CI/CD pipelines, Dockerfiles, and Compose files that set NVIDIA runtime flags. For Compose files, see :doc:`Docker Compose <docker-compose>`.
