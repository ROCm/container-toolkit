Docker Compose
==============

The AMD Container Toolkit can be used with Docker Compose, enabling GPU access in multi-container applications.

Prerequisites
-------------

Before using this guide, ensure you have:

1. Completed the AMD Container Toolkit :doc:`Installation <installation>`.
2. Docker Compose installed (v2 or higher recommended).
3. Docker engine configured with the AMD container runtime, as described in the :doc:`Quick Start <quick-start-guide>`.

Example Docker Compose Configuration
------------------------------------

Below is a basic example of a Docker Compose file configured to use AMD GPUs:

.. code-block:: yaml

   services:
     pytorch:
       image: rocm/pytorch
       runtime: amd
       environment:
         - AMD_VISIBLE_DEVICES=all
       command: python -c "import torch; print('GPU available:', torch.cuda.is_available()); print('Number of GPUs:', torch.cuda.device_count())"

The key elements in this configuration are:

1. **runtime: amd** - Specifies that the AMD container runtime should be used for this service.
2. **AMD_VISIBLE_DEVICES** - Controls which GPUs are visible to the container.

GPU Visibility Control
----------------------

Control GPU visibility through environment variables. The ``AMD_VISIBLE_DEVICES`` variable can be set to:

- **AMD_VISIBLE_DEVICES=all** - Makes all GPUs visible to the container
- **AMD_VISIBLE_DEVICES=0,1** - Makes only GPU indices 0 and 1 visible
- **AMD_VISIBLE_DEVICES=0-3** - Makes GPU indices 0 through 3 visible
- **AMD_VISIBLE_DEVICES=0xEF2C1799A1F3E2ED** - Makes the GPU with this UUID visible (see ``amd-ctk gpu list``)

To run a service without GPUs, omit ``AMD_VISIBLE_DEVICES``.

When using Docker Compose to orchestrate multiple containers, you can specify the GPU resources for each service independently.
For example, if you have a training service and an inference service, you can assign different GPUs to each:

.. code-block:: yaml

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

   If the file selects GPUs with ``HIP_VISIBLE_DEVICES`` and explicit ``/dev/kfd`` and ``/dev/dri`` mappings, set ``AMD_VISIBLE_DEVICES`` to the GPUs the service needs before removing the mappings. ``HIP_VISIBLE_DEVICES`` still applies inside the container, but it indexes only the GPUs that ``AMD_VISIBLE_DEVICES`` exposes.

3. Remove explicit device mappings if present (not needed with AMD Container Toolkit):

   .. code-block:: diff

        services:
          myservice:
      -     devices:
      -       - /dev/kfd
      -       - /dev/dri

Next Steps
----------

- :doc:`Docker Swarm <docker-swarm>` to schedule GPU services across multiple nodes.
- :doc:`Troubleshooting <troubleshooting>` if a service does not see the expected GPUs.
