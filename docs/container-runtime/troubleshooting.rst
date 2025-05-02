Troubleshooting
===============

Common Issues:
--------------

- **Driver Not Loaded**
   - Ensure the `amdgpu` module is loaded: `lsmod | grep amdgpu`.
   - Check for errors in kernel logs: `dmesg | grep amdgpu`.

- **Docker Daemon fails to start after runtime configuration**
  - Verify `/etc/docker/daemon.json` correctness.
  - Check Docker daemon logs: `sudo journalctl -u docker.service`.

- **GPU devices not visible inside container**
  - Confirm that the user is part of `video` and `render` groups.
  - Verify GPU accessibility via `rocm-smi` outside the container.

- **CDI specification errors**
  - Regenerate the CDI spec: `amd-ctk cdi generate`
  - Check `/etc/cdi/amd.json` for corruption.

Diagnostic Commands
-------------------

- List devices:

.. code-block:: bash

   amd-ctk cdi list

- Check runtime configuration:

.. code-block:: bash

   cat /etc/docker/daemon.json
