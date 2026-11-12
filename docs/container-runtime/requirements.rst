System Requirements
===================

This page lists what the AMD Container Toolkit needs on the host.

Supported Operating Systems
---------------------------

- Ubuntu 22.04 LTS (Jammy Jellyfish)
- Ubuntu 24.04 LTS (Noble Numbat)
- Ubuntu 26.04 LTS (Resolute Raccoon)
- RHEL 9

Ubuntu uses the Debian package and RHEL uses the RPM package. See :doc:`Installation <installation>` for both.

AMDGPU Driver and ROCm
----------------------

- The **AMDGPU driver** must be installed and loaded on the host. See the `AMD GPU Driver installation prerequisites <https://instinct.docs.amd.com/projects/amdgpu-docs/en/latest/install/detailed-install/prerequisites.html>`_ for supported kernels and the kernel headers DKMS needs.
- ROCm user-space libraries come from the container image and do not need to be installed on the host. The host driver must be compatible with the ROCm version in your container images; a mismatch can cause runtime failures.
- ROCm 6.4.1 or newer is required inside the container to view and verify partitioned GPUs. See :doc:`GPU Partitioning <gpu-partitioning>`.

Container Runtimes
------------------

.. list-table::
   :header-rows: 1
   :widths: 50 50

   * - Feature
     - Minimum version
   * - CDI with Docker
     - Docker 25.0. Docker 25.0 to 28.1 also need ``"features": {"cdi": true}`` in ``/etc/docker/daemon.json``, which ``sudo amd-ctk runtime configure`` adds. Docker 28.2 and newer enable CDI by default.
   * - ``amd-container-runtime`` with Docker
     - Docker 25.0
   * - ``--gpus`` with ``--runtime=amd``
     - Docker 28.3.0
   * - ``--gpus`` with CDI, without ``--runtime=amd``
     - Docker 29.3.0
   * - Other runtimes (Podman, nerdctl, containerd)
     - Any version with CDI support. See :doc:`Running Workloads <running-workloads>` for ``--gpus`` support in each.

Install Docker from the ``docker.io`` package or Docker's official repositories. Check your version with ``docker --version``.

.. note::
   Docker Desktop on Linux is not supported for GPU workloads. See :doc:`Troubleshooting <troubleshooting>` for details.

.. important::
   The ``amd-container-runtime`` must run as root, so it requires rootful Docker. Rootless Docker is not supported. CDI does not have this restriction; for example, rootless Podman can use the CDI specification.

User Permissions
----------------

Users running ROCm tools on the host, or running rootless Podman containers, must belong to the ``render`` and ``video`` groups:

.. code-block:: bash

   sudo usermod -aG render,video $USER

Log out and back in for the group change to take effect.

Examples run ``docker`` without ``sudo`` and assume your user is in the ``docker`` group. Otherwise, prefix ``docker`` with ``sudo``.

Commands that write to system locations (``/etc/cdi``, ``/etc/docker/daemon.json``) and all ``amd-ctk gpu-tracker`` commands must be run as root.
