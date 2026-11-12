Overview
========

The AMD Container Toolkit gives containers access to AMD Instinct GPUs. It discovers the GPUs on a host and injects the device nodes a container needs, so ROCm workloads run in containers without manual ``--device`` flags.

The toolkit consists of two components:

- **amd-ctk**: a command-line tool that lists GPUs, generates and validates CDI specifications, registers the AMD container runtime with Docker, and manages the :doc:`GPU Tracker <gpu-tracker>`.
- **amd-container-runtime**: an OCI runtime for Docker (a wrapper around ``runc``) that adds the requested GPUs to each container's specification.

Two Ways to Inject GPUs
-----------------------

**Container Device Interface (CDI), recommended.** ``amd-ctk cdi generate`` writes a CDI specification (``/etc/cdi/amd.json``) that describes every GPU on the host. Any CDI-aware runtime (Docker, Podman, containerd) reads it and injects the requested devices, for example ``--device amd.com/gpu=0``. No AMD-specific runtime is involved at container start.

**amd-container-runtime (Docker only).** Docker hands the container's OCI specification to ``amd-container-runtime``, which reads ``AMD_VISIBLE_DEVICES`` (or ``--gpus``), adds the matching GPU devices, and passes the updated specification to ``runc``. This path also supports the :doc:`GPU Tracker <gpu-tracker>` for shared and exclusive GPU access.

.. code-block:: text

   CDI path
   +-----------------+      +----------------------+      +-------------------+
   | docker / podman | ---> | runtime reads        | ---> | container sees    |
   | --device        |      | /etc/cdi/amd.json    |      | requested GPUs    |
   | amd.com/gpu=0   |      | and injects devices  |      |                   |
   +-----------------+      +----------------------+      +-------------------+

   amd-container-runtime path
   +-----------------+      +---------------+      +-----------------------+      +------+
   | docker run      | ---> | Docker daemon | ---> | amd-container-runtime | ---> | runc |
   | --runtime=amd   |      |   OCI spec    |      | adds GPU devices to   |      |      |
   | -e AMD_VISIBLE_ |      |               |      | the OCI spec          |      |      |
   |   DEVICES=0     |      |               |      |                       |      |      |
   +-----------------+      +---------------+      +-----------------------+      +------+

Use Cases
---------

- Running ROCm machine learning and HPC containers on AMD Instinct GPUs.
- Assigning specific GPUs or GPU partitions to different containers.
- Scheduling GPU workloads with :doc:`Docker Compose <docker-compose>` and :doc:`Docker Swarm <docker-swarm>`.

Where to Go Next
----------------

- :doc:`System Requirements <requirements>` and :doc:`Installation <installation>` to set up the toolkit.
- :doc:`Quick Start <quick-start-guide>` to run your first GPU container with either method.
- :doc:`Running Workloads <running-workloads>` for GPU selection with each container runtime.
