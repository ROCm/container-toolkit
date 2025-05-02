Migration Guide: NVIDIA to AMD
==============================

Migrating container workflows from NVIDIA GPUs (using `nvidia-docker`) to AMD Instinct GPUs (using AMD Container Toolkit) involves the following key changes:

Step 1: Environment Variable Updates
-------------------------------------
Replace ``NVIDIA_VISIBLE_DEVICES`` with ``AMD_VISIBLE_DEVICES``.

Step 2: Update Runtime Configuration
-------------------------------------
Instead of ``nvidia`` runtime, configure Docker for ``amd`` runtime.

Step 3: Container Images
------------------------
Use ROCm-compatible containers instead of CUDA-based containers images for ML/AI workloads (e.g., `rocm/pytorch`, `rocm/tensorflow`).

Step 4: Framework Adjustments
-----------------------------
Switch to ROCm-enabled versions of TensorFlow, PyTorch, etc.

Compatibility Notes
-------------------
- Not all CUDA-specific frameworks may have direct ROCm counterparts yet.
- Validation and testing of applications is highly recommended after migration.

Comparison:
-----------

.. list-table:: Feature Comparison: NVIDIA Docker vs AMD Container Toolkit
    :header-rows: 1

    * - Feature
      - NVIDIA Docker
      - AMD Container Toolkit
    * - GPU Enumeration
      - ``nvidia-smi``
      - ``rocm-smi``
    * - Container Runtime
      - ``nvidia-container-runtime``
      - ``amd-container-runtime``
    * - Environment Variable
      - ``NVIDIA_VISIBLE_DEVICES``
      - ``AMD_VISIBLE_DEVICES``
    * - Framework Images
      - NVIDIA-specific
      - ROCm-optimized

By following this migration guide, users can rapidly transition their GPU workloads to AMD Instinct platforms.
