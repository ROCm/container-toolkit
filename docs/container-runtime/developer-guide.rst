Developer Guide
===============

Building from Source
--------------------

The AMD Container Toolkit can be built manually from the source code.

To build Debian packages:

.. code-block:: bash

   make pkg-deb

To build RPM packages:

.. code-block:: bash

   make pkg-rpm

Generated files will be located under the ``bin/`` directory.

Contribution Guidelines
------------------------
- Follow the coding conventions outlined in the developer README.
- Ensure all changes are tested with Docker and ROCm environments.
- Create pull requests with detailed change descriptions.
