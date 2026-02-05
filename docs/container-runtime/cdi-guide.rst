==========================================
Support for Container Device Interface
==========================================

Overview
========

The `Container Device Interface <https://github.com/cncf-tags/container-device-interface>`_ (CDI) is a standardized specification for exposing specialized hardware devices, such as AMD GPUs, to containers in a runtime-agnostic manner. This works consistently across different container runtimes.

CDI eliminates the need for runtime-specific hooks or shims, like ``amd-container-runtime``, by allowing container runtimes to natively understand and inject device resources into containers.

The ``amd-ctk`` tool provides commands to generate and manage CDI specifications for AMD GPU devices on your system.

Prerequisites
=============

Before using CDI with AMD GPUs, ensure:

* AMD GPU drivers are properly installed on the host system
* The ``amd-ctk`` tool is installed
* Your container runtime supports CDI

Generating CDI Specifications
==============================

To generate a CDI specification for AMD GPUs on your system, run:

.. code-block:: bash

    sudo amd-ctk cdi generate

This command:

* Scans the system for available AMD GPU devices
* Creates a CDI specification file at ``/etc/cdi/amd.json``
* Defines device nodes, mount points, and environment variables needed for each GPU

**Custom Output Location**

To generate the specification in a different location, use the ``--output`` flag:

.. code-block:: bash

    amd-ctk cdi generate --output /path/to/custom/amd.json

Validating CDI Specifications
==============================

To verify that your CDI specification matches the actual GPU hardware on the system, run:

.. code-block:: bash

    sudo amd-ctk cdi validate

This command:

* Reads the CDI specification from ``/etc/cdi/amd.json``
* Scans the system for available AMD GPU devices
* Verifies that the devices defined in the specification accurately reflect the hardware present on the host

**Custom Specification Path**

To validate a specification at a different location, use the ``--path`` flag:

.. code-block:: bash

    amd-ctk cdi validate --path /path/to/custom/amd.json

.. note::

    The ``amd-ctk`` tool requires appropriate permissions to read and write CDI specification files. When operating on the default location (``/etc/cdi``), it requires elevated privileges, hence ``sudo`` is typically needed.

    If you want to operate on a different user-owned location (using the ``--output`` or ``--path`` flags for generation or validation respectively), ``sudo`` can be omitted, provided the user has necessary read/write permissions for that location.

    When using a custom output location, ensure your container runtime is configured to read CDI specifications from that directory. Most runtimes default to ``/etc/cdi`` and ``/var/run/cdi``.

.. important::

    Regenerate the CDI specification whenever you:
    
    * Add or remove GPU devices
    * Modify GPU partitioning or configuration

Troubleshooting
===============

Containers Cannot Access GPUs
------------------------------

If containers do not see the expected GPU devices:

1. **Validate the specification:**

   .. code-block:: bash

       sudo amd-ctk cdi validate

   If the validation fails, it indicates a mismatch between the CDI specification and the actual hardware. You may need to regenerate the specification in such cases. 

2. **Verify runtime configuration:**

   Ensure your container runtime is configured to read CDI specifications from the directory containing ``amd.json``. Check the runtime's CDI configuration settings.

3. **Check file permissions:**

   .. code-block:: bash

       ls -l /etc/cdi/amd.json

   The file should be readable by the container runtime process. If you're using a custom location, ensure the permissions allow the runtime to access it.

4. **Regenerate if hardware changed:**

   If you've added, removed, or reconfigured GPUs, regenerate the specification:

   .. code-block:: bash

       sudo amd-ctk cdi generate

5. **Verify device names:**

   Ensure you're using the correct CDI device names (e.g., ``amd.com/gpu=0``) while requesting devices.

Validation Errors
-----------------

If ``amd-ctk cdi validate`` reports errors:

* Check that GPU devices are properly detected by the system (verify with ``rocm-smi``, ``amd-smi`` or similar tools)
* Ensure GPU drivers are correctly installed
* Regenerate the specification to reflect the current system state
