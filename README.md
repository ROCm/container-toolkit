# Overview
ROCm Container Toolkit offers tools to streamline the use of AMD GPUs with containers. The toolkit includes the following packages.
- ```amd-container-runtime``` - The AMD Container Runtime
-  ```amd-ctk``` - The AMD Container Toolkit CLI

# Installation
## Debian Package Install
### System Requirements

Before installing the AMD Container Toolkit, you need to install the following:

*   **Operating System**: Ubuntu 22.04 or Ubuntu 24.04
    
*   **ROCm Version**: 6.3.x (specific to each .deb pkg)
    
Each Debian package release of the AMD Container Toolkit is dependent on a specific version of the ROCm amdgpu driver. Please see table below for more information:

| Container Toolkit Debian Version | ROCm Version | AMDGPU Driver Version |
|----------------------------------|--------------|-----------------------|
| amd-container-toolkit-1.2.0 | ROCm 6.3.x | 6.10.5 |

### Installation

#### Step 1: Install System Prerequisites

1.  Update the system:
    
    ```text
    > sudo apt update
    > sudo apt install "linux-headers-$(uname \-r)" "linux-modules-extra-$(uname \-r)"
    ```

2.  Add user to required groups:
    
    ```text
    > sudo usermod \-a \-G render,video $LOGNAME
    ```

#### Step 2: Install AMDGPU Driver

Note

For the most up-to-date information on installing dkms drivers please see the [ROCm Install Quick Start](https://rocm.docs.amd.com/projects/install-on-linux/en/latest/install/quick-start.html) page. The below instructions are the most current instructions as of ROCm 6.2.4.

1.  Download the driver from the Radeon repository ([repo.radeon.com](https://repo.radeon.com/amdgpu-install)) for your operating system. For example if you want to get the latest ROCm 6.3.4 drivers for Ubuntu 22.04 you would run the following command:
    
    ```text
    > wget https://repo.radeon.com/amdgpu-install/6.3.4/ubuntu/jammy/amdgpu-install\_6.3.60304-1\_all.deb
    ```
    
    Please note that the above url will be different depending on what version of the drivers you will be installing and type of Operating System you are using.
    
2.  Install the driver:
    
    ```text
    > sudo apt install ./amdgpu-install\_6.3.60304-1\_all.deb
    > sudo apt update
    > amdgpu-install \--usecase\=dkms
    ```

3.  Load the driver module:
    
    ```text
    > sudo modprobe amdgpu
    ```

#### Step 3: Install the APT Prerequisites for Container Toolkit.

1.  Update the package list and install necessary tools, keyrings and keys:
    
    ```text
    > sudo apt update
    > sudo apt install vim wget gpg
    
    \# Create the keyrings directory with the appropriate permissions:
    > sudo mkdir \--parents \--mode\=0755 /etc/apt/keyrings
    
    \# Download the ROCm GPG key and add it to the keyrings:
    > wget https://repo.radeon.com/rocm/rocm.gpg.key \-O \- | gpg \--dearmor | sudo tee /etc/apt/keyrings/rocm.gpg \> /dev/null
    ```

2.  Edit the sources list to add the Container Toolkit repository:
    
    ```text
    **For Ubuntu 22.04**, add the following line:
    
    > deb \[arch\=amd64 signed-by\=/etc/apt/keyrings/rocm.gpg\] https://repo.radeon.com/amd-container-toolkit/apt/1.2.0 jammy main
    
    **For Ubuntu 24.04**, add the following line:
    
    > deb \[arch\=amd64 signed-by\=/etc/apt/keyrings/rocm.gpg\] https://repo.radeon.com/amd-container-toolkit/apt/1.2.0 noble main
    ```
    
3.  Update the package list again:
    
    ```text
    > sudo apt update
    ```

#### Step 4: Install the Prerequisites for Container Toolkit

1. Install the Container toolkit.

    ```text
    > sudo apt install amd-container-toolkit
    ```

2. Install docker.

## Debian Package Install

TBD

# Configuring Docker

1. Configure the AMD container runtime for Docker as follows.

     ```text
     > sudo amd-ctk configure runtime
     ```

The above command modifies the docker configuration file, /etc/docker/daemon.json, so that Docker can use the AMD container runtime.

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

     2. Using [CDI](https://github.com/cncf-tags/container-device-interface) style

          - First, generate the CDI spec.

          ```text
          > amd-ctk cdi generate --output=/etc/cdi/amd.json
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

4. Make AMD container runtime default runtime.

Avoid specifying ```--runtime=amd``` option with the ```docker run``` command by setting the AMD container runtime as the default for Docker.

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

# Architecture overview
The Container Toolkit achieves the following.

1. Provide containers access to all or a subset of the GPUs available on the host system, via the environment variable, ```AMD_VISIBLE_DEVICES```. The AMD container runtime is a wrapper on the low level runtime ```runc`` that intercepts and updates the [OCI spec](https://github.com/opencontainers/runtime-spec/blob/main/spec.md) that is generated by the container daemon and passed to ```runc```. Specifically, it injects the GPU devices, requested by the container CLI, into the OCI spec. The diagram below illustrates this for Docker.

     ```text
     +------------+                       +---------------+             +-----------------------+
     |            |                       |               |             |                       |
     | Docker CLI |  AMD_VISIBLE_DEVICES  | Docker Daemon |  OCI SPEC   | AMD Container Runtime |
     |            |          -->          |               |     -->     |                       |
     +------------+                       +---------------+             +-----------------------+
                                                                                    |
                                                                              UPDATED OCI SPEC
                                                                            INJECTED GPU DEVICES
                                                                                    |
                                                                                    v
                                                                         +----------------------+
                                                                         |         RUNC         +
                                                                         +----------------------+
                                                                                    |
                                                                                    v
                                                                         +----------------------+
                                                                         |   CONTAINER PROCESS  +
                                                                         +----------------------+
                                                                                    |
                                                                                    v
                                                                         +----------------------+
                                                                         |      GPU DRIVER      +
                                                                         +----------------------+
     ```

2. Generates the CDI spec for AMD GPUs available on the host system. This enables Users to specify the GPU devices needed in the container using the CDI style. Most containers today have native support for CDI.

3. Provides CLI to configure Docker to use CDI and ```AMD_VISIBLE_DEVICES``` environment variable.

# Troubleshooting

TBD

# Release notes

TBD

# Developer Guide
## Building from Source
To build debian package, use the following command.

```text
make pkg-deb
```

To build rpm package, use the following command.

```text
make pkg-rpm
```

The packages will be generated in the ```bin``` folder.