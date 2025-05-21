Quick Start Guide
=================

This section provides a step-by-step guide to install the AMD Container Toolkit and configure your system for Docker-based GPU container workloads. The steps below are tailored for ease of use, production-readiness, and ensuring compatibility across AMD Instinct GPU-enabled systems.

Prerequisites
-------------

Before installing the AMD Container Toolkit, ensure the following dependencies are installed.

**Docker** - The toolkit is designed to work with Docker, so ensure you have Docker installed on your system. the **Docker version must be 25 or above**. The Container Device Interface (CDI) format, used by modern container runtimes to abstract and expose GPUs, is not supported in older Docker versions. Without Docker 25+, CDI functionality such as dynamic device enumeration and CDI-style run commands will not work as intended.

   .. code-block:: bash

      sudo apt-get install docker.io

You can verify your Docker version using:

   .. code-block:: bash

      docker --version

If you are on an earlier Docker version, please upgrade to at least Docker 25 before proceeding with toolkit configuration and GPU-based workloads.      

**jq** - Required during uninstallation to parse configuration settings cleanly.

   .. code-block:: bash

      sudo apt-get install jq

Step 1: Update System and Group Settings
----------------------------------------
- Update your system:

.. code-block:: bash

   sudo apt update

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

   wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | sudo tee /etc/apt/keyrings/rocm.gpg > /dev/null

- Add the AMD Container Toolkit repository.

Ubuntu 22.04:

.. code-block:: bash

   echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ jammy main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list

Ubuntu 24.04:

.. code-block:: bash

   echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ noble main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list

- Update package index and install the toolkit:

.. code-block:: bash

   sudo apt update

Step 4: Install Toolkit and Docker
----------------------------------

.. code-block:: bash

   sudo apt install amd-container-toolkit

Step 5: Configure Docker Runtime for AMD GPUs
---------------------------------------------

- Register the AMD container runtime and restart the Docker daemon:

.. code-block:: bash

   sudo amd-ctk configure runtime
   sudo systemctl restart docker

This configuration ensures that Docker is aware of the AMD container runtime and is able to support GPU-accelerated workloads using AMD Instinct devices.

Step 6: Verify Container Runtime Installation
---------------------------------------------

To run Docker containers with access to AMD GPUs, you need to specify the AMD runtime and visible GPUs. Here are some examples you can use to verify the installation:

**Run a container with access to all available AMD GPUs:**
  .. code-block:: bash
     docker run --runtime=amd -e AMD_VISIBLE_DEVICES=all --runtime=amd rocm/rocm-terminal amd-smi monitor
      GPU  POWER   GPU_T   MEM_T   GFX_CLK   GFX%   MEM%   ENC%   DEC%      VRAM_USAGE
        0  137 W   41 °C   36 °C   142 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        1  139 W   39 °C   33 °C   135 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        2  138 W   42 °C   34 °C   145 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        3  141 W   39 °C   33 °C   139 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        4  140 W   42 °C   36 °C   146 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        5  137 W   38 °C   33 °C   133 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        6  139 W   43 °C   36 °C   151 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB
        7  137 W   41 °C   34 °C   141 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB

**To run a container with access to a specific AMD GPU (i.e., the first GPU):**
  .. code-block:: bash
     docker run --runtime=amd -e AMD_VISIBLE_DEVICES=0 --runtime=amd rocm/rocm-terminal amd-smi monitor
     GPU  POWER   GPU_T   MEM_T   GFX_CLK   GFX%   MEM%   ENC%   DEC%      VRAM_USAGE
       0  140 W   42 °C   36 °C   146 MHz    0 %    0 %    N/A    0 %    0.3/192.0 GB

Uninstallation Guide
--------------------

To remove the `amd-container-toolkit`, you must have `jq` installed. The uninstallation script relies on it to parse configuration files.

.. code-block:: bash

   sudo apt-get install jq

Then proceed with the removal:

.. code-block:: bash

   sudo apt-get remove --purge amd-container-toolkit

If you encounter issues, inspect the logs:

.. code-block:: bash

   sudo journalctl -u apt

   sudo tail -f /var/log/amd-container-runtime.log


If you continue to face errors, you may need to force the removal:

.. code-block:: bash

   sudo dpkg --remove --force-all amd-container-toolkit

   sudo apt-get autoremove
