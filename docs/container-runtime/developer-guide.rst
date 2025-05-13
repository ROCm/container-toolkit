Developer Guide
===============

The AMD Container Toolkit can be built and customized from the source code. This guide provides step-by-step instructions to set up your environment, install necessary dependencies, and build the toolkit packages for deployment.

System Preparation
------------------

To successfully build the AMD Container Toolkit from source, you need the following build dependencies installed:

1. **Build Essentials**

.. code-block:: bash

   sudo apt update
   sudo apt install build-essential cmake pkg-config

2. **Kernel Headers and Modules** (Required for DKMS module compilation)

.. code-block:: bash

   sudo apt install "linux-headers-$(uname -r)" "linux-modules-extra-$(uname -r)"

   **Note:** These are necessary for integrating the ROCm GPU drivers with your Linux kernel. Ensure the version matches your current kernel.

3. **Development Libraries**

.. code-block:: bash

   sudo apt install libssl-dev libelf-dev libudev-dev

4. **Docker CLI and Daemon** (Required for container runtime testing)

.. code-block:: bash

   sudo apt install docker.io

   # Start Docker and enable it on boot
   sudo systemctl enable docker
   sudo systemctl start docker

5. **ROCm Dependencies**

Ensure that ROCm is properly installed and configured:

.. code-block:: bash

   sudo apt update
   sudo apt install rocm-dev rocm-utils

Building the Toolkit
---------------------

To build Debian packages:

.. code-block:: bash

   make pkg-deb

The generated `.deb` files will be available under the `bin/` directory. These can be installed using:

.. code-block:: bash

   sudo dpkg -i bin/amd-container-toolkit-<version>.deb


To build RPM packages:

.. code-block:: bash

   make pkg-rpm

The RPM packages will also be located in the `bin/` directory. For installation:

.. code-block:: bash

   sudo rpm -i bin/amd-container-toolkit-<version>.rpm

Contribution Guidelines
------------------------

Contributions to the AMD Container Toolkit are welcomed and encouraged. Follow these guidelines to ensure smooth collaboration:

1. **Coding Standards**:
   - Adhere to the coding conventions outlined in the `developer README`.
   - Maintain clear, concise, and well-structured code.

2. **Testing Requirements**:
   - All changes must be tested with Docker and ROCm environments.
   - Use `amd-ctk list` and Docker integration tests to validate GPU access.

3. **Pull Request Requirements**:
   - Include detailed descriptions of changes.
   - Reference any related issues or bug fixes.
   - Attach testing logs or screenshots if applicable.

Advanced Configuration
----------------------

For developers looking to extend the runtime or integrate custom modules, make sure you:

- Rebuild the kernel modules if kernel headers are updated.

.. code-block:: bash

   sudo dkms install -m amdgpu -v <version>

- Restart Docker to load new configurations:

.. code-block:: bash

   sudo systemctl restart docker

Next Steps
----------

- Deploy the built packages in a development environment for further testing.
- Validate compatibility with your ROCm-based applications.
- Document any discrepancies or runtime anomalies.

By following these steps, you will have a robust, production-ready build of the AMD Container Toolkit, optimized for high-performance containerized workloads.
