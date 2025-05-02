System Requirements
====================

Before installing or using the AMD Container Toolkit, ensure your environment meets the following prerequisites:

Operating Systems
-----------------
- Ubuntu 22.04 LTS (Jammy Jellyfish)
- Ubuntu 24.04 LTS (Noble Numbat)

Software Requirements:
-----------------------
- ROCm Software Stack: Version 6.3.x
- AMDGPU Driver: Version 6.10.5 (or corresponding to ROCm 6.3.x)
- Toolkit packages are tested with specific ROCm and AMDGPU driver versions.

Hardware Requirements:
-----------------------
- An AMD GPU supported by the ROCm 6.3.x release.
- CPU with virtualization support (if containers require nested environments).

Compatibility Matrix
--------------------

- Each AMD Container Toolkit release is tightly coupled with a specific ROCm and AMDGPU driver version. Please refer to the compatibility matrix before proceeding.

+--------------------------------------+---------------+-----------------------+
| Container Toolkit Debian Version     | ROCm Version  | AMDGPU Driver Version |
+--------------------------------------+---------------+-----------------------+
| amd-container-toolkit-1.2.0          | ROCm 6.3.x    | 6.10.5                |
+--------------------------------------+---------------+-----------------------+

Note
----
A mismatch between ROCm and driver versions may lead to runtime failures.

System Prerequisites
---------------------
- Kernel Headers
- Extra Kernel Modules
- Docker installed (docker.io package recommended)
- User must belong to the `render` and `video` groups for GPU access.
