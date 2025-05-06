Quick Start Guide
=================

This section provides a step-by-step guide to install the AMD Container Toolkit and configure your system for Docker-based GPU container workloads. The steps below are tailored for ease of use, production-readiness, and ensuring compatibility across AMD Instinct GPU-enabled systems.

Step 1: Install System Prerequisites
------------------------------------
- Update your system and install necessary packages to ensure kernel compatibility:

.. code-block:: bash

   sudo apt update
   sudo apt install "linux-headers-$(uname -r)" "linux-modules-extra-$(uname -r)"

- Add your user to the required groups for GPU device access:

.. code-block:: bash

   sudo usermod -a -G render,video $LOGNAME

Step 2: Install the AMDGPU Driver
---------------------------------

- Refer to the latest ROCm documentation for driver installation here, [ROCm Install Quick Start](https://rocm.docs.amd.com/projects/install-on-linux/en/latest/install/quick-start.html).
- Download the AMDGPU driver installer package from the [Radeon Repository](https://repo.radeon.com/amdgpu-install).
- Install the downloaded package.
- Load the driver.

.. code-block:: bash

   #Example (for Ubuntu 22.04, ROCm 6.3.4)
   wget https://repo.radeon.com/amdgpu-install/6.3.4/ubuntu/jammy/amdgpu-install_6.3.60304-1_all.deb
   sudo apt install ./amdgpu-install_6.3.60304-1_all.deb
   sudo apt update
   amdgpu-install --usecase=dkms
   sudo modprobe amdgpu

Step 3: Configure Repositories
-------------------------------

- Install required dependencies:

.. code-block:: bash

   sudo apt update
   sudo apt install vim wget gpg

- Create keyrings directory

.. code-block:: bash

   sudo mkdir --parents --mode=0755 /etc/apt/keyrings

- Install GPG keys and repository links:

.. code-block:: bash

   sudo mkdir --parents --mode=0755 /etc/apt/keyrings
   wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | sudo tee /etc/apt/keyrings/rocm.gpg > /dev/null

- Add the AMD Container Toolkit repository.

Ubuntu 22.04:

.. code-block:: bash

   echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/1.2.0 jammy main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list

Ubuntu 24.04:

.. code-block:: bash

   echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/1.2.0 noble main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list

- Update package index and install the toolkit:

.. code-block:: bash

   sudo apt update

Step 4: Install Toolkit and Docker
----------------------------------

.. code-block:: bash

   sudo apt install amd-container-toolkit
   #Install Docker (if not already installed)
   sudo apt install docker.io

.. important::

   Please note — the **Docker version must be 25 or above**. The Container Device Interface (CDI) format, used by modern container runtimes to abstract and expose GPUs, is not supported in older Docker versions. Without Docker 25+, CDI functionality such as dynamic device enumeration and CDI-style run commands will not work as intended.

   You can verify your Docker version using:

   .. code-block:: bash

      docker --version

If you are on an earlier Docker version, please upgrade to at least Docker 25 before proceeding with toolkit configuration and GPU-based workloads.

Step 5: Configure Docker Runtime for AMD GPUs
---------------------------------------------

- Register the AMD container runtime and restart the Docker daemon:

.. code-block:: bash

   sudo amd-ctk configure runtime
   sudo systemctl restart docker

This configuration ensures that Docker is aware of the AMD container runtime and is able to support GPU-accelerated workloads using AMD Instinct devices.
