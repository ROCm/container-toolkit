================
Running Workloads
================

This page describes how to use the AMD Container Toolkit to run GPU-accelerated workloads through different container runtimes and CLIs. You can inject AMD GPUs into containers in two ways: via **CDI specs** (recommended) or via the **amd-container-runtime**.

Through CDI
===========

The :doc:`Container Device Interface (CDI) <cdi-guide>` lets you expose AMD GPUs to containers in a runtime-agnostic way. Once a CDI specification is generated and available (e.g. at ``/etc/cdi/amd.json``), any CDI-aware runtime and CLI can inject GPUs using the ``--device amd.com/gpu=<entry>`` pattern.

Prerequisites
-------------

Ensure CDI specs are set up on your system. Refer to the :doc:`CDI guide <cdi-guide>` for details.

Docker
~~~~~~

Use ``--device amd.com/gpu=<entry>``. You do **not** need ``--runtime=amd`` when using CDI.

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all rocm/rocm-terminal rocm-smi

.. tip::

   Starting with Docker **29.3.0**, you can also use ``--gpus <entry>`` directly with CDI, without ``--runtime=amd``. ``<entry>`` can be ``all``, a count (e.g. ``2``), or specific devices (e.g. ``'"device=0,1,2"'``).

   .. code-block:: bash

      docker run --rm --gpus all rocm/rocm-terminal rocm-smi

Podman
~~~~~~

.. code-block:: bash

   podman run --rm --device amd.com/gpu=all rocm/rocm-terminal rocm-smi

.. note::

   To access the GPU inside the container, the process must run under the video and render groups. When running in rootless mode, ensure the user starting the Podman container is a member of these groups on the host, and use the ``--group-add keep-groups`` flag to pass these supplementary groups to the container process.

nerdctl
~~~~~~~

nerdctl works with containerd and supports CDI via ``--device``.

.. code-block:: bash

   nerdctl run --rm --device amd.com/gpu=all rocm/rocm-terminal rocm-smi

ctr
~~~

ctr is containerd's native CLI and supports CDI via ``--device``.

.. code-block:: bash

   ctr run --rm --device amd.com/gpu=all docker.io/rocm/rocm-terminal:latest mycontainer rocm-smi

Requesting specific GPUs
~~~~~~~~~~~~~~~~~~~~~~~~

To request specific GPUs instead of all, use ``--device amd.com/gpu=<entry>`` with the available entry for the corresponding GPU(s) on your machine. List valid entries with:

.. code-block:: bash

   amd-ctk cdi list

Example output:

.. code-block:: text

   Found 2 AMD GPU devices
   amd.com/gpu=all
   amd.com/gpu=0
     /dev/dri/renderD128
   amd.com/gpu=1
     /dev/dri/renderD129

Use the listed device names (e.g. ``all``, ``0``, ``1``) as ``<entry>`` in the CLI commands above.

.. note::

   nerdctl and ctr use the containerd backend; Docker and Podman use their own runtimes. All of the above rely on the same CDI spec (e.g. ``/etc/cdi/amd.json``) and ``amd-ctk cdi list`` for ``<entry>`` values.

Through amd-container-runtime
=============================

The **amd-container-runtime** is a custom OCI runtime that injects AMD GPUs into containers. At this time it is supported only with **Docker**. For setup (registering the runtime and restarting Docker), see the :doc:`Quick Start Guide <quick-start-guide>`.

**All GPUs:**

.. code-block:: bash

   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi

For particular GPUs, use exact GPU indices, ranges, or UUIDs with ``AMD_VISIBLE_DEVICES``:

.. code-block:: bash

   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0,1 rocm/rocm-terminal rocm-smi
   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0-3,5 rocm/rocm-terminal rocm-smi
   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0xEF2C1799A1F3E2ED rocm/rocm-terminal rocm-smi

Use ``amd-ctk gpu list`` to discover available GPUs and their UUIDs:

.. code-block:: bash

   amd-ctk gpu list

Example output:

.. code-block:: text

   Found 2 AMD GPU devices
   ---------------------------------------------------------------------------
   GPU Id    UUID                     DRM Devices
   ---------------------------------------------------------------------------
   0         0xEF2C1799A1F3E2ED       /dev/dri/renderD128
   1         0x1234567890ABCDEF       /dev/dri/renderD129

.. note::

   Docker 28.3.0+ supports the standardized ``--gpus`` flag (e.g. ``--gpus all`` or ``--gpus device=0,1``) as an alternative to ``-e AMD_VISIBLE_DEVICES=all``.

For setup and installation, see the :doc:`Quick Start Guide <quick-start-guide>`. For troubleshooting, see the :doc:`Troubleshooting <troubleshooting>` guide.
