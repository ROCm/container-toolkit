Framework Integration
======================

The AMD Container Toolkit is framework-agnostic but works seamlessly with popular machine learning, HPC, and AI frameworks that require GPU access, including:

- TensorFlow (ROCm builds)
- PyTorch (ROCm builds)
- ONNX Runtime
- OpenMPI + ROCm
- Custom AI/ML workflows

The examples below use :doc:`CDI <cdi-guide>` device notation (``--device amd.com/gpu=<entry>``). Ensure a CDI specification has been generated before running these commands.

TensorFlow
----------

Run ROCm-enabled TensorFlow with a single GPU:

.. code-block:: bash

   docker run --rm --device amd.com/gpu=0 tensorflow/tensorflow:rocm-latest

Or with all available GPUs:

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all tensorflow/tensorflow:rocm-latest

PyTorch
-------

Use ROCm-enabled PyTorch containers:

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all rocm/pytorch:latest

Triton Inference Server
-----------------------

Serving models with Triton using AMD GPUs is supported by adapting container images for ROCm:

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all <triton-rocm-image>

Best Practices
--------------

- Always use container images tested against the matching ROCm version.
- Prefer CDI device notation (``--device amd.com/gpu=<entry>``) for portability across container runtimes.
- Use ``amd-ctk cdi list`` to discover available device entries for multi-GPU setups.
