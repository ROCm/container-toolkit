Installation
============

This page covers installing, upgrading, and removing the AMD Container Toolkit packages. Check the :doc:`System Requirements <requirements>` first.

Step 1: Install the AMDGPU Driver
---------------------------------

The toolkit needs the AMDGPU driver on the host. ROCm user-space libraries come from your container images and are not needed on the host.

Install the driver by following the `AMD GPU Driver installation guide <https://instinct.docs.amd.com/projects/amdgpu-docs/en/latest/install/package-manager-index.html>`_ for your distribution, including the reboot it requires. Then confirm that the driver is installed and loaded:

.. code-block:: bash

   sudo dkms status
   lsmod | grep amdgpu

.. note::
   Container and immutable OS images can install the toolkit packages without a
   loaded AMDGPU driver. Load the driver on the target host before generating
   CDI specifications or running GPU workloads.

Step 2: Configure the Repository
--------------------------------

Ubuntu:

.. code-block:: bash

   sudo apt update
   sudo apt install wget gpg
   sudo mkdir --parents --mode=0755 /etc/apt/keyrings
   wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | sudo tee /etc/apt/keyrings/rocm.gpg > /dev/null
   echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ $(. /etc/os-release && echo $VERSION_CODENAME) main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list
   sudo apt update

RHEL 9:

.. code-block:: bash

   sudo tee /etc/yum.repos.d/amd-container-toolkit.repo <<EOF
   [amd-container-toolkit]
   name=amd-container-toolkit
   baseurl=https://repo.radeon.com/amd-container-toolkit/el9/main/
   enabled=1
   priority=50
   gpgcheck=1
   gpgkey=https://repo.radeon.com/rocm/rocm.gpg.key
   EOF

Step 3: Install the Toolkit
---------------------------

Ubuntu:

.. code-block:: bash

   sudo apt install amd-container-toolkit

RHEL 9:

.. code-block:: bash

   sudo dnf clean all
   sudo dnf install -y amd-container-toolkit

Verify the installation:

.. code-block:: bash

   amd-ctk version
   amd-ctk gpu list

The package installs ``amd-ctk`` and ``amd-container-runtime``. Continue with the :doc:`Quick Start <quick-start-guide>` to configure your container runtime and run a GPU container.

Upgrading
---------

Upgrade the package with your package manager:

.. code-block:: bash

   # Ubuntu
   sudo apt update && sudo apt install --only-upgrade amd-container-toolkit

   # RHEL
   sudo dnf upgrade amd-container-toolkit

After upgrading, validate the CDI specification with ``sudo amd-ctk cdi validate`` and regenerate it if needed.

Uninstalling
------------

Ubuntu:

.. code-block:: bash

   sudo apt-get remove --purge amd-container-toolkit

RHEL:

.. code-block:: bash

   sudo dnf remove amd-container-toolkit

Removing the package:

- Removes ``/etc/cdi/amd.json`` if it is an AMD GPU specification. Other files in ``/etc/cdi`` and specifications at custom paths are left in place.
- Removes the ``amd`` runtime, ``default-runtime`` (if set to ``amd``), and the ``cdi`` feature entry from ``/etc/docker/daemon.json``. Restart Docker afterwards. On Docker 25.0 to 28.1, this disables CDI; add the entry back if other CDI specifications are still in use.
- Removes the :doc:`GPU Tracker <gpu-tracker>` state in ``/var/log/gpu-tracker.json`` on ``purge`` (Ubuntu) or on removal (RHEL).

If a cleanup step fails, the package is still removed and a message names the file to clean up manually.

To undo the Docker runtime registration without uninstalling the package, run ``sudo amd-ctk runtime configure --remove`` and restart Docker.
