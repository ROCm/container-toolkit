=================
Running Workloads
=================

This page describes how to use the AMD Container Toolkit to run GPU-accelerated workloads through different container runtimes and CLIs. You can inject AMD GPUs into containers in two ways: via a **CDI specification** (recommended) or via the **AMD container runtime** (``amd-container-runtime``).

Through CDI
===========

The :doc:`Container Device Interface (CDI) <cdi-guide>` lets you expose AMD GPUs to containers in a runtime-agnostic way. Once a CDI specification is generated and available (e.g. at ``/etc/cdi/amd.json``), any CDI-aware runtime and CLI can inject GPUs using the ``--device amd.com/gpu=<entry>`` pattern.

Prerequisites
-------------

Generate the CDI specification as shown in :doc:`Quick Start <quick-start-guide>` (Option A). See the :doc:`CDI guide <cdi-guide>` for custom paths and validation.

Docker
------

Use ``--device amd.com/gpu=<entry>``. You do **not** need ``--runtime=amd`` when using CDI.

Docker 25.0 to 28.1 need CDI enabled first; see :doc:`System Requirements <requirements>`.

.. code-block:: bash

   docker run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi

.. tip::

   Starting with Docker **29.3.0**, you can also use ``--gpus <entry>`` for AMD GPUs with CDI, without ``--runtime=amd``. ``<entry>`` can be ``all``, a count (e.g. ``2``), or specific devices (e.g. ``'"device=0,1,2"'``).

   .. code-block:: bash

      docker run --rm --gpus all rocm/dev-ubuntu-24.04 amd-smi

Podman
------

.. code-block:: bash

   podman run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi

.. tip::

   Starting with Podman **6.0.0**, you can also use ``--gpus <entry>`` for AMD GPUs with CDI. ``<entry>`` can be ``all`` to request all GPUs.

   .. code-block:: bash

      podman run --rm --gpus all rocm/dev-ubuntu-24.04 amd-smi

.. note::

   To access the GPU inside the container, the process must run under the video and render groups. When running in rootless mode, ensure the user starting the Podman container is a member of these groups on the host, and use the ``--group-add keep-groups`` flag to pass these supplementary groups to the container process.

nerdctl
-------

nerdctl works with containerd and supports CDI via ``--device``.

.. code-block:: bash

   nerdctl run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi

.. tip::

   Starting with nerdctl **2.3.0**, you can also use ``--gpus <entry>`` for AMD GPUs with CDI. ``<entry>`` can be ``all`` or a count (e.g. ``2``).

   .. code-block:: bash

      nerdctl run --rm --gpus all rocm/dev-ubuntu-24.04 amd-smi

ctr
---

ctr is containerd's native CLI and supports CDI via ``--device``. It requires root and does not pull images on ``run``, so pull the image first.

.. code-block:: bash

   sudo ctr image pull docker.io/rocm/dev-ubuntu-24.04:latest
   sudo ctr run --rm --device amd.com/gpu=all docker.io/rocm/dev-ubuntu-24.04:latest mycontainer amd-smi

.. tip::

   Starting with containerd **2.3.0**, ``ctr`` also supports the ``--gpus`` flag for AMD GPUs with CDI. GPUs are requested by index (e.g. ``--gpus 0``).

   .. code-block:: bash

      sudo ctr run --rm --gpus 0 docker.io/rocm/dev-ubuntu-24.04:latest mycontainer amd-smi

Requesting Specific GPUs
------------------------

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

   nerdctl and ctr use the containerd backend; Docker and Podman use their own runtimes. All of the above rely on the same CDI specification (e.g. ``/etc/cdi/amd.json``) and ``amd-ctk cdi list`` for ``<entry>`` values.

Through the AMD Container Runtime
=================================

The **amd-container-runtime** is a custom OCI runtime that injects AMD GPUs into containers. At this time it is supported only with **Docker**. For setup (registering the runtime and restarting Docker), see the :doc:`Quick Start <quick-start-guide>`.

**All GPUs:**

.. code-block:: bash

   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/dev-ubuntu-24.04 amd-smi

For particular GPUs, use exact GPU indices, ranges, or UUIDs with ``AMD_VISIBLE_DEVICES``:

.. code-block:: bash

   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0,1 rocm/dev-ubuntu-24.04 amd-smi
   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0-3,5 rocm/dev-ubuntu-24.04 amd-smi
   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0xEF2C1799A1F3E2ED rocm/dev-ubuntu-24.04 amd-smi

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

To run every container with the AMD container runtime without passing ``--runtime=amd``, set it as the Docker default:

.. code-block:: bash

   sudo amd-ctk runtime configure --set-as-default
   sudo systemctl restart docker

``amd-ctk runtime configure`` supports Docker only and accepts these options:

- ``--set-as-default``: set ``amd`` as Docker's ``default-runtime``.
- ``--unset-as-default``: remove ``default-runtime`` from ``daemon.json``, whichever runtime it names.
- ``--remove``: remove the ``amd`` runtime, ``default-runtime`` if it is ``amd``, and the ``cdi`` feature entry. On Docker 25.0 to 28.1, this disables CDI.
- ``--config-path <file>``: edit a Docker configuration file other than ``/etc/docker/daemon.json``.

Restart Docker after any change.

Using the ``--gpus`` Flag
-------------------------

Starting with Docker **28.3.0**, the standardized ``--gpus`` flag can be used with ``--runtime=amd`` instead of ``-e AMD_VISIBLE_DEVICES``.

- All GPUs:

  .. code-block:: bash

     docker run --rm --runtime=amd --gpus all rocm/dev-ubuntu-24.04 amd-smi
     docker run --rm --runtime=amd --gpus device=all rocm/dev-ubuntu-24.04 amd-smi

- The first 2 GPUs (indices 0 and 1):

  .. code-block:: bash

     docker run --rm --runtime=amd --gpus 2 rocm/dev-ubuntu-24.04 amd-smi

- Specific GPUs by index, starting from 0. The ``device=`` prefix is required:

  .. code-block:: bash

     docker run --rm --runtime=amd --gpus device=2 rocm/dev-ubuntu-24.04 amd-smi
     docker run --rm --runtime=amd --gpus '"device=1,2,3"' rocm/dev-ubuntu-24.04 amd-smi

.. note::
   Without ``device=``, a comma-separated list such as ``--gpus 1,2,3`` is read as a count, and only the last number is used. That example requests **3 GPUs**, not GPUs 1, 2, and 3.

For troubleshooting, see the :doc:`Troubleshooting <troubleshooting>` guide.
