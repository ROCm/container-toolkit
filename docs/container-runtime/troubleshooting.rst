Troubleshooting
===============

This page covers common problems with the AMD Container Toolkit, grouped by where they show up. Start with the host checks, then move to the section for the injection method you use.

Host Checks
-----------

Driver Not Loaded
~~~~~~~~~~~~~~~~~

If the AMD GPU driver is not detected, verify that the ``amdgpu`` module is loaded:

.. code-block:: bash

   lsmod | grep amdgpu

If the module is not present, check whether the driver package is installed:

.. code-block:: bash

   sudo dkms status

If ``amdgpu`` is not listed, follow the `AMD GPU Driver installation guide <https://instinct.docs.amd.com/projects/amdgpu-docs/en/latest/install/package-manager-index.html>`_. If it is listed, load the module manually:

.. code-block:: bash

   sudo modprobe amdgpu

If you encounter errors, check the kernel logs for driver loading issues:

.. code-block:: bash

   dmesg | grep amdgpu

Confirm that the toolkit can see the GPUs:

.. code-block:: bash

   amd-ctk gpu list

Permission Denied Errors
~~~~~~~~~~~~~~~~~~~~~~~~

If ROCm tools on the host or rootless Podman containers cannot access the GPUs:

- Verify GPU accessibility using ``rocm-smi`` or ``amd-smi`` outside the container.
- Ensure the user belongs to the ``render`` and ``video`` groups:

  .. code-block:: bash

     groups $USER

  If you are not a member, add yourself to the groups, then log out and back in:

  .. code-block:: bash

     sudo usermod -a -G render,video $USER

Docker Desktop and /dev/kfd Access
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Docker Desktop on Linux is not supported for GPU workloads. You may see:

.. code-block:: text

   docker: Error response from daemon: error gathering device information while adding custom device "/dev/kfd": no such file or directory

**Why:** Docker Desktop on Linux runs the Docker daemon inside a VM (or similar isolated context). That VM does not have the host's ``/dev/kfd`` and ``/dev/dri`` devices mounted, so containers started by that daemon cannot access them.

**Workaround:** Install Docker via the ``docker.io`` package or Docker's official repository so the daemon runs on the host and can expose these devices to containers. Alternatively, quit Docker Desktop and use Docker installed on the host.

This applies to any run that relies on host GPU devices (e.g. ``docker run --device=/dev/kfd --device=/dev/dri ...`` or ``docker run --runtime=amd -e AMD_VISIBLE_DEVICES=...``).

CDI Issues
----------

CDI Not Enabled in Docker
~~~~~~~~~~~~~~~~~~~~~~~~~

If Docker reports ``could not select device driver "cdi" with capabilities: []``, CDI is not enabled in Docker. Docker 28.2 and newer enable CDI by default. Docker 25.0 to 28.1 require ``"features": {"cdi": true}`` in ``/etc/docker/daemon.json``, which ``sudo amd-ctk runtime configure`` adds. Restart Docker after changing the file.

Unresolvable CDI Devices
~~~~~~~~~~~~~~~~~~~~~~~~

If Docker reports ``unresolvable CDI devices amd.com/gpu=...``, the CDI specification is missing or does not define that device.

1. Check that you are using a valid device name. List the names with:

   .. code-block:: bash

      amd-ctk cdi list

2. Regenerate the specification:

   .. code-block:: bash

      sudo amd-ctk cdi generate

3. Ensure your container runtime reads CDI specifications from the directory containing ``amd.json``. Most runtimes default to ``/etc/cdi`` and ``/var/run/cdi``. If you generated the specification at a custom location, configure the runtime to read from it.

4. If the error persists, restart Docker:

   .. code-block:: bash

      sudo systemctl restart docker

Containers Do Not See the Expected GPUs
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

1. Validate the specification against the hardware:

   .. code-block:: bash

      sudo amd-ctk cdi validate

   If it prints ``CDI spec is invalid``, the specification no longer matches the GPUs on the host. This happens after adding or removing GPUs or changing the partitioning scheme. Regenerate it with ``sudo amd-ctk cdi generate``.

2. Check that the container runtime process can read the specification:

   .. code-block:: bash

      ls -l /etc/cdi/amd.json

   If you are using a custom location, ensure the permissions allow the runtime to access it.

3. Inspect the generated specification:

   .. code-block:: bash

      cat /etc/cdi/amd.json

CDI Validation Errors
~~~~~~~~~~~~~~~~~~~~~

If ``amd-ctk cdi validate`` fails with an error:

* Check that GPU devices are properly detected by the system (verify with ``amd-ctk gpu list``, ``amd-smi``, or ``rocm-smi``).
* Ensure GPU drivers are correctly installed.
* Regenerate the specification to reflect the current system state.
* Use the ``--debug`` flag for verbose output:

  .. code-block:: bash

     sudo amd-ctk --debug cdi validate

AMD Container Runtime Issues
----------------------------

Docker Does Not Recognize the AMD Runtime
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If Docker reports ``unknown or invalid runtime name: amd``, validate the Docker configuration:

.. code-block:: bash

   cat /etc/docker/daemon.json

``amd-ctk runtime configure`` writes the following entries:

.. code-block:: json

   {
      "features": {
          "cdi": true
      },
      "runtimes": {
          "amd": {
              "path": "amd-container-runtime",
              "args": []
          }
      }
   }

If the configuration is missing or incorrect, regenerate it and restart Docker:

.. code-block:: bash

   sudo amd-ctk runtime configure
   sudo systemctl restart docker

Docker Daemon Restart Failure
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If Docker fails to restart after configuring the AMD container runtime, inspect the Docker logs:

.. code-block:: bash

   sudo journalctl -u docker

Look for errors related to:

- Container runtime conflicts
- GPU device issues
- Invalid JSON or unknown keys in ``/etc/docker/daemon.json``

Container Fails to Start with a Generic Runtime Error
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If ``docker run --runtime=amd`` fails with ``amd-container-runtime did not terminate successfully``, the specific cause is in the runtime log (see :ref:`log-file-reference`). Common causes:

- ``Rootless mode is not supported``: the runtime was started by a non-root Docker daemon. In this case the log is written to ``$LOGDIR`` if set, otherwise to the home directory of the user the runtime runs as (``~/amd-container-runtime.log``), instead of ``/var/log``. Use rootful Docker, or use CDI instead.
- ``GPUs [...] are exclusive and already in use``: the :doc:`GPU Tracker <gpu-tracker>` is enabled and another container holds an exclusive GPU. Check with ``sudo amd-ctk gpu-tracker status``.

.. _log-file-reference:

Logs and Debug Output
---------------------

The AMD container runtime (``amd-container-runtime``) logs events and errors to ``/var/log/amd-container-runtime.log``.

You can view logs in real-time using:

.. code-block:: bash

   sudo tail -f /var/log/amd-container-runtime.log

This log captures detailed interactions between Docker and the AMD container runtime, including:

- Runtime initialization
- GPU device injection and allocation
- OCI specification modifications
- Exclusive GPU enforcement errors

To write the log elsewhere, set ``LOGDIR`` in the Docker daemon's environment (for example, with a systemd drop-in for ``docker.service``).

If a container fails to start with the AMD container runtime, this log contains the specific error (e.g. ``GPUs [0] are exclusive and already in use``), even when Docker only shows a generic runtime failure message.

The ``amd-ctk`` CLI tool prints errors directly to the terminal (not to a log file). For verbose debug output from ``amd-ctk``, use the ``--debug`` (or ``-d``) flag:

.. code-block:: bash

   sudo amd-ctk --debug gpu-tracker status
   sudo amd-ctk --debug cdi validate

This prints debug-level messages to stderr, which can help diagnose GPU enumeration, tracker state, or CDI specification issues.

Diagnostic Commands
-------------------

- List GPUs detected on the host: ``amd-ctk gpu list``
- List CDI device names: ``amd-ctk cdi list``
- Check the Docker configuration: ``cat /etc/docker/daemon.json``
- Inspect Docker logs: ``sudo journalctl -u docker``
- Check that the driver is loaded: ``lsmod | grep amdgpu``
- Verify GPU access from a container: ``docker run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi``
- If ROCm is installed on the host: ``rocminfo`` and ``amd-smi``

Getting Help
------------

If these steps do not resolve your issue, open an issue on `GitHub <https://github.com/ROCm/container-toolkit/issues>`_ with the output of the diagnostic commands above and the relevant lines from ``/var/log/amd-container-runtime.log``.
