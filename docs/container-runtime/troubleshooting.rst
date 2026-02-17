Troubleshooting
===============

The AMD Container Toolkit is designed to integrate smoothly into Docker-based environments. However, issues may arise due to system configurations, driver installations, or runtime settings. This guide aims to provide detailed, step-by-step troubleshooting methods to identify and resolve common issues effectively.

Common Issues:
--------------

1. **Driver Not Loaded**
------------------------

If the AMD GPU driver is not detected, verify that the `amdgpu` module is loaded:

.. code-block:: bash

   lsmod | grep amdgpu

If the module is not present, attempt to load it manually:

.. code-block:: bash

   sudo modprobe amdgpu

If you encounter errors, check the kernel logs for driver loading issues:

.. code-block:: bash

   dmesg | grep amdgpu

This will provide information about any problems during the driver initialization.

2. **Permission Denied Errors**
--------------------------------

If GPU devices are not visible inside containers:

- Verify GPU accessibility using `rocm-smi` outside the container.
- Ensure the user belongs to the following groups:

  - `render`
  - `video`

Verify your group membership:

.. code-block:: bash

   groups $USER

If you are not a member, add yourself to the necessary groups:

.. code-block:: bash

   sudo usermod -a -G render,video $USER

**Note:** Log out and back in for the changes to take effect.

3. **Docker Daemon Restart Failure**
------------------------------------

If Docker fails to restart after configuring the AMD runtime, inspect the Docker logs:

.. code-block:: bash

   sudo journalctl -u docker

Look for errors related to:

- Container runtime conflicts
- GPU device issues
- Improper `/etc/docker/daemon.json` configuration

Verify that the runtime path is correctly set for AMD:

.. code-block:: bash

   cat /etc/docker/daemon.json

4. **Runtime Configuration Issues**
------------------------------------

If Docker does not recognize the AMD runtime, validate the Docker configuration:

.. code-block:: bash

   cat /etc/docker/daemon.json

Ensure the runtime is set correctly:

.. code-block:: json

   {
      "runtimes": {
          "amd": {
              "path": "/usr/bin/amd-container-runtime",
              "runtimeArgs": []
          }
      }
   }

If the configuration is missing or incorrect, regenerate it and restart Docker:

.. code-block:: bash

   sudo amd-ctk configure runtime
   sudo systemctl restart docker


5. **CDI Specification Not Applied**
-------------------------------------

If Docker does not recognize the GPU under CDI specifications, regenerate the CDI configuration:

.. code-block:: bash

   sudo amd-ctk cdi generate --output=/etc/cdi/amd.json

Check the integrity of the generated specification:

.. code-block:: bash

   cat /etc/cdi/amd.json

If issues persist, restart Docker:

.. code-block:: bash

   sudo systemctl restart docker

6. **Docker Desktop and /dev/kfd Access**
-------------------------------------------------

Docker Desktop on Linux is not supported for GPU workloads. You may see:

.. code-block:: text

   docker: Error response from daemon: error gathering device information while adding custom device "/dev/kfd": no such file or directory

**Why:** Docker Desktop on Linux runs the Docker daemon inside a VM (or similar isolated context). That VM does not have the host's ``/dev/kfd`` and ``/dev/dri`` devices mounted, so containers started by that daemon cannot access them.

**Workaround:** Install Docker via the ``docker.io`` package or Docker's official repository so the daemon runs on the host and can expose these devices to containers. Alternatively, quit Docker Desktop and use Docker installed on the host.

This applies to any run that relies on host GPU devices (e.g. ``docker run --device=/dev/kfd --device=/dev/dri ...`` or ``docker run --runtime=amd -e AMD_VISIBLE_DEVICES=...``).

Log File Reference
------------------

The AMD Container Toolkit logs runtime events and errors to the following location:

   **/var/log/amd-container-runtime.log**

You can view logs in real-time using:

.. code-block:: bash

   sudo tail -f /var/log/amd-container-runtime.log

This log captures detailed interactions between Docker and the AMD container runtime, including:

- Runtime initialization
- GPU device injection
- OCI specification modifications
- CDI specification usage

If you experience issues that are not easily diagnosed, refer to this log file for real-time insights and deeper debugging.

Diagnostic Commands
-------------------

- **List Available Devices:**

   .. code-block:: bash

      amd-ctk cdi list

- **Check Runtime Configuration:**

   .. code-block:: bash

      cat /etc/docker/daemon.json

- **Inspect Docker Logs:**

   .. code-block:: bash

      sudo journalctl -u docker

Next Steps
----------

If the above steps do not resolve your issue:

- Validate your amdgpu driver installation with:

.. code-block:: bash

   rocminfo

- Verify GPU accessibility with:

.. code-block:: bash

   rocm-smi

- Consult the official AMD Container Toolkit documentation or reach out to the support community for advanced troubleshooting.
