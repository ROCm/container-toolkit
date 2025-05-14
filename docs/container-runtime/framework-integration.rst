Framework Integration
======================

The AMD Container Toolkit is framework-agnostic but works seamlessly with popular machine learning, HPC, and AI frameworks that require GPU access, including:

- TensorFlow (ROCm builds)
- PyTorch (ROCm builds)
- ONNX Runtime
- OpenMPI + ROCm
- Custom AI/ML workflows

Typical Example:
----------------

- Enabling easy container-based development and deployment across AMD GPU systems.

1. TensorFlow
--------------

Run ROCm-enabled TensorFlow:

.. code-block:: bash

   sudo docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0 tensorflow/tensorflow:rocm-latest

2. PyTorch
-----------

Use ROCm-enabled PyTorch containers:

.. code-block:: bash

   sudo docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/pytorch:latest

3.Triton Inference Server
-------------------------

Serving models with Triton using AMD GPUs is supported by adapting container images for ROCm.

Best Practices
--------------

- Always use container images tested against the matching ROCm version.
- Use environment variables or CDI device selection carefully in multi-GPU setups.
