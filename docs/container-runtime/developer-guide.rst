Developer Guide
===============

The AMD Container Toolkit is written in Go and can be built from source. Packages are built inside builder containers, so the build host only needs Docker, ``make``, and ``git``.

System Preparation
------------------

Install the build host dependencies:

.. code-block:: bash

   sudo apt update
   sudo apt install docker.io make git

Start Docker and enable it on boot:

.. code-block:: bash

   sudo systemctl enable --now docker

To build the binaries directly on the host instead of in a builder container, also install Go at the version listed in ``go.mod``.

No GPU, AMDGPU driver, or ROCm installation is needed to build the toolkit.

Building the Binaries
---------------------

Build the binaries inside the Debian builder container:

.. code-block:: bash

   make

This builds the builder image from ``tools/deb/base-image`` and places ``amd-container-runtime`` and ``amd-ctk`` in ``bin/deb/``.

To build directly on the host with a local Go toolchain:

.. code-block:: bash

   make container-toolkit container-toolkit-ctk

Building the Packages
---------------------

To build a Debian package, run ``make`` first so the builder image exists, then:

.. code-block:: bash

   make pkg-deb

The target Ubuntu release defaults to 22.04 (``jammy``). Set ``UBUNTU_VERSION`` to ``noble`` (24.04) or ``resolute`` (26.04) to build for another release:

.. code-block:: bash

   make pkg-deb UBUNTU_VERSION=noble

The package is written to ``bin/amd-container-toolkit_<ubuntu-version>_amd64.deb`` and can be installed with:

.. code-block:: bash

   sudo apt install ./bin/amd-container-toolkit_22.04_amd64.deb

To build an RPM package, first build the RPM builder image, then the package:

.. code-block:: bash

   make build-dev-container-rpm
   make pkg-rpm

The RPM is copied to ``bin/`` and can be installed with:

.. code-block:: bash

   sudo dnf install ./bin/amd-container-toolkit-*.x86_64.rpm

Testing
-------

``make test`` and ``make lint`` run on the host and need Go at the version listed in ``go.mod``. ``make test`` also needs ``bin/deb/amd-ctk``, so run ``make`` or ``make container-toolkit-ctk`` first.

Run the unit tests:

.. code-block:: bash

   make test

Run the linter:

.. code-block:: bash

   make lint

Contribution Guidelines
------------------------

Contributions to the AMD Container Toolkit are welcomed and encouraged. Follow these guidelines to ensure smooth collaboration:

1. **Coding Standards**:

   - Format code with ``gofmt`` and keep ``make lint`` clean.
   - Maintain clear, concise, and well-structured code.

2. **Testing Requirements**:

   - Add or update unit tests for code changes and run ``make test``.
   - For runtime or CDI changes, validate GPU access on a GPU host with ``amd-ctk gpu list``, ``amd-ctk cdi validate``, and a container run.

3. **Pull Request Requirements**:

   - Include detailed descriptions of changes.
   - Reference any related issues or bug fixes.
   - Attach testing logs or screenshots if applicable.
