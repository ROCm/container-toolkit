System Requirements
====================

Before installing or using the AMD Container Toolkit, ensure your environment meets the following prerequisites:

Operating Systems
-----------------
- Ubuntu 22.04 LTS (Jammy Jellyfish)
- Ubuntu 24.04 LTS (Noble Numbat)

Compatibility Matrix
--------------------
- Please refer to the compatibility matrix before proceeding.

.. list-table:: Compatibility Matrix
    :header-rows: 1
    :widths: 30 20

    * - Container Toolkit Debian Version
      - Docker Version
    * - amd-container-toolkit-1.0.0
      - 25+

Note
----
A mismatch between ROCm and driver versions may lead to runtime failures.

System Prerequisites
---------------------
- Kernel Headers
- Extra Kernel Modules
- Docker installed (docker.io package recommended)
- User must belong to the `render` and `video` groups for GPU access.
