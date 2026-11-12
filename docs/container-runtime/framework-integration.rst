Framework Integration
======================

The AMD Container Toolkit is framework-agnostic. Any ROCm-enabled container image can use the GPUs it injects, including:

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

   docker run --rm --device amd.com/gpu=0 rocm/tensorflow:latest \
     python3 -c "import tensorflow as tf; print(tf.config.list_physical_devices('GPU'))"

Or with all available GPUs:

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all rocm/tensorflow:latest \
     python3 -c "import tensorflow as tf; print(tf.config.list_physical_devices('GPU'))"

PyTorch
-------

Use ROCm-enabled PyTorch containers:

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all --ipc=host rocm/pytorch:latest \
     python3 -c "import torch; print(torch.cuda.device_count())"

For interactive work, replace the command with ``-it ... bash``. Multi-process data loaders need shared memory, so pass ``--ipc=host`` or a larger ``--shm-size``.

Best Practices
--------------

- Always use container images tested against the matching ROCm version.
- Prefer CDI device notation (``--device amd.com/gpu=<entry>``) for portability across container runtimes.
- Use ``amd-ctk cdi list`` to discover available device entries for multi-GPU setups.
