# Overview
AMD Container Toolkit offers tools to streamline the use of AMD GPUs with containers. The toolkit includes the following packages.
- ```amd-container-runtime``` - The AMD Container Runtime
-  ```amd-ctk``` - The AMD Container Toolkit CLI

# Requirements
- Ubuntu 22.02 or 24.04, or RHEL/CentOS 9
- Docker version 25 or later

# Quick Start
Install the Container toolkit.

### Installing on Ubuntu
To install the AMD Container Toolkit on Ubuntu systems, follow these steps:

1. Ensure pre-requisites are installed
   ```bash
   apt update && apt install -y wget gnupg2
   ```

2. Add the GPG key for the repository:
   ```bash
   wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | tee /etc/apt/keyrings/rocm.gpg > /dev/null
   ```

3. Add the repository to your system. Replace `noble` with `jammy` if you are using Ubuntu 22.04:
   ```bash
   echo "deb [signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ noble main" > /etc/apt/sources.list.d/amd-container-toolkit.list
   ```

4. Update the package list and install the toolkit:
   ```bash
   apt update && apt install amd-container-toolkit
   ```

### Installing on RHEL/CentOS 9
To install the AMD Container Toolkit on RHEL/CentOS 9 systems, follow these steps:

1. Add the repository configuration:
   ```bash
   tee --append /etc/yum.repos.d/amd-container-toolkit.repo <<EOF
   [amd-container-toolkit]
   name=amd-container-toolkit
   baseurl=https://repo.radeon.com/amd-container-toolkit/el9/main/
   enabled=1
   priority=50
   gpgcheck=1
   gpgkey=https://repo.radeon.com/rocm/rocm.gpg.key
   EOF
   ```

2. Clean the package cache and install the toolkit:
   ```bash
   dnf clean all
   dnf install -y amd-container-toolkit
   ```

# Configuring Docker

1. Configure the AMD container runtime for Docker as follows. The following command modifies the docker configuration file, /etc/docker/daemon.json, so that Docker can use the AMD container runtime.

     ```text
     > sudo amd-ctk runtime configure
     ```

2. Restart the Docker daemon.

     ``` text
     > sudo systemctl restart docker
     ```

# Docker Runtime Integration
1. Configure Docker to use AMD container runtime.

``` text
> amd-ctk runtime configure --runtime=docker
```
2. Specify the required GPUs. There are 3 ways to do this.

     1. Using ```AMD_VISIBLE_DEVICES``` environment variable

          - To use all available GPUs,

          ```text
          > docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi
          ```

          - To use a subset of available GPUs,

          ```text
          > docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0,1,2 rocm/rocm-terminal rocm-smi
          ```

          - To use many contiguously numbered GPUs,

          ```text
          > docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0-3,5,8 rocm/rocm-terminal rocm-smi
          ```

     2. Using [CDI](https://github.com/cncf-tags/container-device-interface) style

          - First, generate the CDI spec.

          ```text
          > amd-ctk cdi generate --output=/etc/cdi/amd.json
          ```

          - Validate the generated CDI spec.

          ```text
          > amd-ctk cdi validate --path=/etc/cdi/amd.json
          ```

          - To use all available GPUs,

          ```text
          > docker run --rm --device amd.com/gpu=all rocm/rocm-terminal rocm-smi
          ```

          - To use a subset of available GPUs,

          ```text
          > docker run --rm --device amd.com/gpu=0 --device amd.com/gpu=1 rocm/rocm-terminal rocm-smi
          ```
          - Note that once the CDI spec, ```/etc/cdi/amd.json``` is available, ```runtime=amd``` is not required in the docker run command.

     3. Using explicit paths. Note that ```runtime=amd``` is not required here.

     ```text
     > docker run --device /dev/kfd --device /dev/dri/renderD128 --device /dev/dri/renderD129 rocm/rocm-terminal rocm-smi
     ```

3. List available GPUs.

```text
> amd-ctk cdi list
Found 1 AMD GPU device
amd.com/gpu=all
amd.com/gpu=0
  /dev/dri/card1
  /dev/dri/renderD128
```

4. Make AMD container runtime default runtime. Avoid specifying ```--runtime=amd``` option with the ```docker run``` command by setting the AMD container runtime as the default for Docker.

```text
> amd-ctk runtime configure --runtime=docker --set-as-default
```

5. Remove AMD container runtime as default runtime.

```text
> amd-ctk runtime configure --runtime=docker --unset-as-default
```

6. Remove AMD container runtime configuration in Docker (undo the earlier configuration).

``` text
> amd-ctk runtime configure --runtime=docker --remove
```

# Device discovery and enumeration

The following command can be used to list the GPUs available on the system and their enumberation. The GPUs are listed in the CDI format, but the same enumeration applies to usage with the OCI environment variable, ```AMD_VISIBLE_DEVICES```.

```text
> amd-ctk cdi list
Found 1 AMD GPU device
amd.com/gpu=all
amd.com/gpu=0
  /dev/dri/card1
  /dev/dri/renderD128
```

# Release notes
| Release  | Features                                                                     | Known Issues                                                                 |
|----------|------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| v1.1.0   | 1. GPU partitioning support<br>2. Full RPM package support<br>3. Support for range operator in the input string to AMD_VISIBLE_DEVICES ENV variable. | None                                                                         |
| v1.0.0   | Initial release                                                             | 1. Partitioned GPUs are not supported.<br>2. RPM builds are experimental.   |
 
## Building from Source
To build debian package, use the following command.

```text
make
make pkg-deb
```

To build rpm package, use the following command.

```text
make build-dev-container-rpm
make pkg-rpm
```

The packages will be generated in the ```bin``` folder.

# Documentation
For detailed documentation including installation guides and configuration options, see the [documentation](https://instinct.docs.amd.com/projects/container-toolkit/en/latest).

# License
This project is licensed under the Apache 2.0 License - see the [LICENSE](https://github.com/ROCm/container-toolkit/blob/main/LICENSE) file for details.


