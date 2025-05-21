Docker Compose Usage
====================

The AMD Container Toolkit can be used with Docker Compose, enabling GPU access in multi-container applications.

Prerequisites
-------------

Before using this guide, ensure you have:

1. Completed the AMD Container Toolkit installation as described in the Quick Start Guide.
2. Docker Compose installed (v2 or higher recommended).
3. Docker engine properly configured with the AMD runtime.

Example Docker Compose Configuration
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Below is a basic example of a Docker Compose file configured to use AMD GPUs:

.. code-block:: yaml

   version: '3'
   services:
     pytorch:
       image: rocm/pytorch
       runtime: amd
       environment:
         - AMD_VISIBLE_DEVICES=all
       command: python -c "import torch; print('GPU available:', torch.cuda.is_available()); print('Number of GPUs:', torch.cuda.device_count())"

The key elements in this configuration are:

1. **runtime: amd** - Specifies that the AMD Container Runtime should be used for this service.
2. **AMD_VISIBLE_DEVICES** - Controls which GPUs are visible to the container.

GPU Visibility Control
----------------------

Control GPU visibility through environment variables. The `AMD_VISIBLE_DEVICES` variable can be set to:

- **AMD_VISIBLE_DEVICES=all** - Makes all GPUs visible to the container
- **AMD_VISIBLE_DEVICES=0,1** - Makes only GPU indices 0 and 1 visible
- **AMD_VISIBLE_DEVICES=none** - Disables GPU visibility

When using Docker Compose to orchestrate multiple containers, you can specify the GPU resources for each service independently.
For example, if you have a training service and an inference service, you can assign different GPUs to each:

.. code-block:: yaml

   version: '3'
   services:
     training:
       image: rocm/tensorflow
       runtime: amd
       environment:
         - AMD_VISIBLE_DEVICES=0

     inference:
       image: rocm/pytorch
       runtime: amd
       environment:
         - AMD_VISIBLE_DEVICES=1

Converting Existing Docker Compose Files
----------------------------------------

If you have existing Docker Compose files using NVIDIA GPUs or other GPU setups, use these guidelines to migrate to the AMD Container Toolkit:

1. Replace runtime specifications:

   .. code-block:: diff

      - runtime: nvidia
      + runtime: amd

2. Update environment variables:

   .. code-block:: diff

      - NVIDIA_VISIBLE_DEVICES: all
      + AMD_VISIBLE_DEVICES: all

      - HIP_VISIBLE_DEVICES: 0,1
      + AMD_VISIBLE_DEVICES: 0,1

3. Remove explicit device mappings if present (not needed with AMD Container Toolkit):

   .. code-block:: diff

      services:
        myservice:
          # Remove these lines
          - devices:
          -   - /dev/kfd
          -   - /dev/dri
