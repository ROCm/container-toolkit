# Overview
AMD Container Toolkit offers tools to streamline the use of AMD GPUs with containers. The toolkit includes the following packages.
- ```amd-container-runtime``` - The AMD Container Runtime
-  ```amd-ctk``` - The AMD Container Toolkit CLI

# Requirements
- Ubuntu 22.04 or 24.04, or RHEL/CentOS 9
- Docker version 25 or later (on Linux, use Docker Engine rather than Docker Desktop for GPU device access; see [troubleshooting](docs/container-runtime/troubleshooting.rst) if you see `/dev/kfd` errors)
- All the 'amd-ctk runtime configure' commands should be run as root/sudo

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

     2. Using [CDI](docs/container-runtime/cdi-guide.rst) style

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
If this command is run as root, the container-toolkit logs go to /var/log/amd-container-runtime.log, otherwise they go to the user's home directory.

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

# GPU UUID Support

The AMD Container Toolkit supports GPU selection using unique identifiers (UUIDs) in addition to device indices. This enables more precise and reliable GPU targeting, especially in multi-GPU systems and orchestrated environments.

## Getting GPU UUIDs

You can obtain GPU UUIDs using different tools:

### Using ROCm SMI
```bash
rocm-smi --showuniqueid
```

This will display output similar to:
```
GPU[0]          : Unique ID: 0xef2c1799a1f3e2ed
GPU[1]          : Unique ID: 0x1234567890abcdef
```

### Using AMD-SMI
You can also use `amd-smi` to get the ASIC_SERIAL, which serves as the GPU UUID:

```bash
amd-smi static -aB
```

This will display output similar to:
```
GPU: 0
    ASIC:
        MARKET_NAME: AMD Instinct MI210
        VENDOR_ID: 0x1002
        VENDOR_NAME: Advanced Micro Devices Inc. [AMD/ATI]
        SUBVENDOR_ID: 0x1002
        DEVICE_ID: 0x740f
        SUBSYSTEM_ID: 0x0c34
        REV_ID: 0x02
        ASIC_SERIAL: 0xD1CC3F11CFDD5112
        OAM_ID: N/A
        NUM_COMPUTE_UNITS: 104
        TARGET_GRAPHICS_VERSION: gfx90a
    BOARD:
        MODEL_NUMBER: 102-D67302-00
        PRODUCT_SERIAL: 692231000131
        FRU_ID: 113-HPED67302000B.009
        PRODUCT_NAME: Instinct MI210
        MANUFACTURER_NAME: AMD
```

Use the `ASIC_SERIAL` value (e.g., `0xD1CC3F11CFDD5112`) as the GPU UUID in your container configurations.

## Using UUIDs with Environment Variables

Both `AMD_VISIBLE_DEVICES` and `DOCKER_RESOURCE_*` environment variables support UUID specification:

### Using AMD_VISIBLE_DEVICES
```bash
# Use specific GPUs by UUID
docker run --rm --runtime=amd \
  -e AMD_VISIBLE_DEVICES=0xef2c1799a1f3e2ed,0x1234567890abcdef \
  rocm/dev-ubuntu-24.04 rocm-smi

# Mix device indices and UUIDs
docker run --rm --runtime=amd \
  -e AMD_VISIBLE_DEVICES=0,0xef2c1799a1f3e2ed \
  rocm/dev-ubuntu-24.04 rocm-smi
```

### Using DOCKER_RESOURCE_* Variables
```bash
# Docker Swarm generic resource format
docker run --rm --runtime=amd \
  -e DOCKER_RESOURCE_GPU=0xef2c1799a1f3e2ed \
  rocm/dev-ubuntu-24.04 rocm-smi
```

## Docker Swarm Integration

GPU UUID support significantly improves Docker Swarm deployments by enabling precise GPU allocation across cluster nodes.

### Docker Daemon Configuration for Swarm

Configure each swarm node's Docker daemon with GPU resources in `/etc/docker/daemon.json`:

```json
{
  "default-runtime": "amd",
  "runtimes": {
    "amd": {
      "path": "amd-container-runtime",
      "runtimeArgs": []
    }
  },
  "node-generic-resources": [
    "AMD_GPU=0x378041e1ada6015",
    "AMD_GPU=0xef39dad16afb86ad",
    "GPU_COMPUTE=0x583de6f2d99dc333"
  ]
}
```

After updating the configuration, restart the Docker daemon:
```bash
sudo systemctl restart docker
```

### Service Definition

Deploy services with specific GPU requirements using docker-compose:

**Using generic resources:**

```yaml
# docker-compose.yml for Swarm deployment
version: '3.8'
services:
  rocm-service:
    image: rocm/dev-ubuntu-24.04
    command: rocm-smi
    deploy:
      replicas: 1
      resources:
        reservations:
          generic_resources:
            - discrete_resource_spec:
                kind: 'AMD_GPU'  # Matches daemon.json key
                value: 1
```

**Using environment variables:**

```yaml
# docker-compose.yml for Swarm deployment with environment variable
version: '3.8'
services:
  rocm-service:
    image: rocm/dev-ubuntu-24.04
    command: rocm-smi
    environment:
      - AMD_VISIBLE_DEVICES=all
    deploy:
      replicas: 1
```

Deploy the service:
```bash
docker stack deploy -c docker-compose.yml rocm-stack
```

# GPU Tracker

Currently, barebones Docker provides no way to track access of GPUs in containers. Additionally, by default, multiple containers in Docker can be granted access to the same GPU simultaneously. GPU Tracker is an extremely lightweight feature of AMD Container Toolkit that solves these issues.

GPU Tracker state is initialized during AMD Container Toolkit installation and is by default disabled. Users can enable or disable the GPU Tracker feature by using the `enable` or `disable` CLIs. When enabled, the GPU Tracker automatically maintains the state of GPUs and the containers that they are made accessible to, only if the containers are launched and granted access to the GPUs using the `AMD_VISILE_DEVICES` environment variable. When the container process completes execution or is stopped, the GPU Tracker state is automatically updated to reflect GPUs released by the specific container.

**NOTE:** GPU Tracker feature is currenly supported only if containers are started using the `docker run` command and GPUs are made accessible in containers using the `AMD_VISIBLE_DEVICES` environment variable. If containers are started and granted access to GPUs in any other manner, GPU Tracker feature is not supported.

GPU Tracker provides CLIs that can be used to control the accessibility of GPUs in containers. The accessibility of GPUs can be set to either `shared` or `exclusive`.
- The `shared` accessibility indicates that the GPU can be made accessibile to multiple containers simultaneously. By default, all GPUs are granted the `shared` accessibility to reflect the default Docker behavior.
- The `exclusive` accessibility indicates that the GPU can be made accessible to at most one container at any point of time.

GPU Tracker status can be queried at any point of time using the `status` command and reset using the `reset` CLIs.

```text
> sudo amd-ctk gpu-tracker -h
NAME:
   AMD Container Toolkit CLI gpu-tracker - GPU Tracker related commands

USAGE:
   amd-ctk gpu-tracker [gpu-ids] [accessibility]

     Arguments:
       gpu-ids        Comma-separated list of GPU IDs (comma separated list, range operator, all)
       accessibility  Must be either 'exclusive' or 'shared'

     Examples:
       amd-ctk gpu-tracker 0,1,2 exclusive
       amd-ctk gpu-tracker 0,1-2 shared
       amd-ctk gpu-tracker all shared

   OR

   amd-ctk gpu-tracker [command] [options]

COMMANDS:
   disable  Disable the GPU Tracker
   enable   Enable the GPU Tracker
   reset    Reset the GPU Tracker
   status   Show Status of GPUs
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help
```

## Using GPU Tracker

Let us assume that the node has 4 GPUs as indicated below

```text
> rocm-smi


========================================= ROCm System Management Interface =========================================
=================================================== Concise Info ===================================================
Device  Node  IDs              Temp    Power  Partitions          SCLK    MCLK     Fan  Perf  PwrCap  VRAM%  GPU%
              (DID,     GUID)  (Edge)  (Avg)  (Mem, Compute, ID)
====================================================================================================================
0       4     0x740f,   12261  33.0°C  42.0W  N/A, N/A, 0         800Mhz  1600Mhz  0%   auto  300.0W  0%     0%
1       5     0x740f,   13566  38.0°C  40.0W  N/A, N/A, 0         800Mhz  1600Mhz  0%   auto  300.0W  0%     0%
2       3     0x740f,   57300  34.0°C  42.0W  N/A, N/A, 0         800Mhz  1600Mhz  0%   auto  300.0W  0%     0%
3       2     0x740f,   1997   38.0°C  41.0W  N/A, N/A, 0         800Mhz  1600Mhz  0%   auto  300.0W  0%     0%
====================================================================================================================
=============================================== End of ROCm SMI Log ================================================
```
  1. Show GPU Tracker Status:

      Once AMD Container Toolkit, is installed, the GPU Tracker is initialized and the status can be queried using the `status` CLI. If GPU Tracker is enabled, by default it can be seen that GPUs are granted the `shared` accessibility.

      ```text
      > amd-ctk gpu-tracker status
      ------------------------------------------------------------------------------------------------------------------------
      GPU Id    UUID                     Accessibility       Container Ids
      ------------------------------------------------------------------------------------------------------------------------
      0         0xEA35F57CC80DEB35       Shared              -
      1         0x89CAA15875FF5A43       Shared              -
      2         0x6E32F10EFC982B4C       Shared              -
      3         0x12FE4F7FDAF06B9        Shared              -
      ```

      If GPU Tracked feature is not enabled, then a message indicating this is printed.

      ```text
      > amd-ctk gpu-tracker status
      GPU Tracker is disabled
      ```

  2. Enabling GPU Tracker:

      GPU Tracker can be enabled using the `enable` CLI. When GPU Tracker is newly enabled, it starts tracking usage of GPUs in containers with no prior knowledge of GPUs state. If GPU Tracker is already currently enabled, then nothing happens and a message indicating this is printed.

      ```text
      > amd-ctk gpu-tracker status
      GPU Tracker is disabled

      > amd-ctk gpu-tracker enable
      GPU Tracker has been enabled

      > amd-ctk gpu-tracker enable
      GPU Tracker is already enabled
      ```

  3. Disabling GPU Tracker:

      GPU Tracker can be disabled using the `disable` CLI. If GPU Tracker is again enabled in the future, all the GPUs state related information will be lost.

      ```text
      > amd-ctk gpu-tracker disable
      GPU Tracker has been disabled

      > amd-ctk gpu-tracker status
      GPU Tracker is disabled
      ```

  4. Granting access to GPUs in Docker containers:

      If GPU Tracker is enabled before launching container, it automatically tracks the usage of GPUs in containers as indicated below.

      ```text
      > docker run --runtime=amd -itd -e AMD_VISIBLE_DEVICES=0-2 rocm/rocm-terminal bash
      36b012bb34c96149a6ef5b28623e6e75cf9f71eb2b824b2c8f44e0449c7a1aa8

      > docker run --runtime=amd -itd -e AMD_VISIBLE_DEVICES=1,3 rocm/rocm-terminal bash
      90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd

      > amd-ctk gpu-tracker status
      ------------------------------------------------------------------------------------------------------------------------
      GPU Id    UUID                     Accessibility       Container Ids
      ------------------------------------------------------------------------------------------------------------------------
      0         0xEA35F57CC80DEB35       Shared              36b012bb34c96149a6ef5b28623e6e75cf9f71eb2b824b2c8f44e0449c7a1aa8
      1         0x89CAA15875FF5A43       Shared              36b012bb34c96149a6ef5b28623e6e75cf9f71eb2b824b2c8f44e0449c7a1aa8
                                                             90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
      2         0x6E32F10EFC982B4C       Shared              36b012bb34c96149a6ef5b28623e6e75cf9f71eb2b824b2c8f44e0449c7a1aa8
      3         0x12FE4F7FDAF06B9        Shared              90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd

      > docker rm -f 36b012bb34c96149a6ef5b28623e6e75cf9f71eb2b824b2c8f44e0449c7a1aa8
      36b012bb34c96149a6ef5b28623e6e75cf9f71eb2b824b2c8f44e0449c7a1aa8

      > amd-ctk gpu-tracker status
      ------------------------------------------------------------------------------------------------------------------------
      GPU Id    UUID                     Accessibility       Container Ids
      ------------------------------------------------------------------------------------------------------------------------
      0         0xEA35F57CC80DEB35       Shared              -
      1         0x89CAA15875FF5A43       Shared              90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
      2         0x6E32F10EFC982B4C       Shared              -
      3         0x12FE4F7FDAF06B9        Shared              90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
      ```

  5. Setting GPUs to have `exclusive` accessibility:

      If GPU Tracker is enabled, GPUs can be set to have exclusive access in containers. If the user tries to make GPUs exclusive when GPU Tracker is disabled, nothing happens and a message indicating that GPU Tracker is disabled is printed.

      ```text
      > amd-ctk gpu-tracker 1-3 exclusive
      GPUs [1 2 3] have been made exclusive

      > amd-ctk gpu-tracker status
      ------------------------------------------------------------------------------------------------------------------------
      GPU Id    UUID                     Accessibility       Container Ids
      ------------------------------------------------------------------------------------------------------------------------
      0         0xEA35F57CC80DEB35       Shared              -
      1         0x89CAA15875FF5A43       Exclusive           90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
      2         0x6E32F10EFC982B4C       Exclusive           -
      3         0x12FE4F7FDAF06B9        Exclusive           90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd

      > docker run --runtime=amd -itd -e AMD_VISIBLE_DEVICES=0-2 rocm/rocm-terminal bash
      d23ff3dce1839cbf8ce7ad362641ab85e80b315c319edf73b269c460e348053a
      docker: Error response from daemon: failed to create task for container: failed to create shim task: OCI runtime create failed: unable to retrieve OCI runtime error (open /run/containerd/io.containerd.runtime.v2.task/moby/d23ff3dce1839cbf8ce7ad362641ab85e80b315c319edf73b269c460e348053a/log.json: no such file or directory): amd-container-runtime did not terminate successfully: exit status 1: GPUs [0 2] allocated
      GPUs [1] are exclusive and already in use
      Released GPUs [2 0] used by container d23ff3dce1839cbf8ce7ad362641ab85e80b315c319edf73b269c460e348053a
      : unknown.

      > amd-ctk gpu-tracker status
      ------------------------------------------------------------------------------------------------------------------------
      GPU Id    UUID                     Accessibility       Container Ids
      ------------------------------------------------------------------------------------------------------------------------
      0         0xEA35F57CC80DEB35       Shared              -
      1         0x89CAA15875FF5A43       Exclusive           90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
      2         0x6E32F10EFC982B4C       Exclusive           -
      3         0x12FE4F7FDAF06B9        Exclusive           90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
      ```

      In the above example, GPUs 1,2 and 3 have been granted `exclusive` access.

      When a new container `d23ff3dce1839cbf8ce7ad362641ab85e80b315c319edf73b269c460e348053a` that requests access to GPUs 0,1 and 2 is launched, the following happens:
      - The new container is created.
      - The new container is granted access to GPU 0 as no container is currently using GPU 0.
      - GPUs 1 is already being used by container `90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd`. Hence, the new container is not granted access to it as GPU 1 has `exclusive` accessibility.
      - The new container is granted access to GPU 2 as no container is currently using GPU 2 though GPU 2 has `exclusive` accessibility.
      - The container is not started since it has not been granted access to the required GPU resources.
      - The resources that have been granted to the new container are released.

      **NOTE:**

      - Even though the new container `d23ff3dce1839cbf8ce7ad362641ab85e80b315c319edf73b269c460e348053a`  is not successfully started, it is still visible when we run `docker ps -a` command.

          ```text
          > docker ps -a
          CONTAINER ID   IMAGE                                                                                 COMMAND                  CREATED          STATUS                    PORTS     NAMES
          d23ff3dce183   rocm/rocm-terminal                                                                    "bash"                   11 seconds ago   Created                             funny_gagarin
          90cb29e11e83   rocm/rocm-terminal                                                                    "bash"                   45 seconds ago   Up 44 seconds                       practical_williams
          ```

          This is because Docker has already created the container when the runtime errors out due to non-availability of resources. This behavior is similar to behavior exhibited by Docker when a container fails to start in any stage after the container is created in Docker. In such cases also, the container is visible in the `docker ps -a` command output with status as `Created` as depicted below.

          ```text
          > docker run -itd ubuntu incorrect_command
          94f11c132e8cd0a35d05bcc8bcaf77264563998d07f6ad5c73798cf9ddd94726
          docker: Error response from daemon: failed to create task for container: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: exec: "incorrect_command": executable file not found in $PATH: unknown.

          > docker ps -a
          CONTAINER ID   IMAGE                                                                                 COMMAND                  CREATED          STATUS                    PORTS     NAMES
          94f11c132e8c   ubuntu                                                                                "incorrect_command"      17 seconds ago   Created                             elastic_ardinghelli
          ```

      - Only GPUs that are currently not being used by more than 1 container can be set to have `exclusive` accessibility.

          ```text
          > amd-ctk gpu-tracker status
          ------------------------------------------------------------------------------------------------------------------------
          GPU Id    UUID                     Accessibility       Container Ids
          ------------------------------------------------------------------------------------------------------------------------
          0         0xEA35F57CC80DEB35       Shared              8463b475b55b104b30edec8ddf6249b6214b27127106aa0ff4a8a514b856810e
          1         0x89CAA15875FF5A43       Shared              90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
                                                                 8463b475b55b104b30edec8ddf6249b6214b27127106aa0ff4a8a514b856810e
          2         0x6E32F10EFC982B4C       Exclusive           8463b475b55b104b30edec8ddf6249b6214b27127106aa0ff4a8a514b856810e
          3         0x12FE4F7FDAF06B9        Exclusive           90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd

          > amd-ctk gpu-tracker 1 exclusive
          GPUs [1] have not been made exclusive because more than one container is currently using it
          ```

  6. Setting GPUs to have `shared` accessibility:

      If GPU Tracker is enabled, GPUs can be set to have shared access in containers. If the user tries to make GPUs shared when GPU Tracker is disabled, nothing happens and a message indicating that GPU Tracker is disabled is printed. By default when GPU Tracker is disabled, GPUs have shared accessibility.

      ```text
      > amd-ctk gpu-tracker status
      ------------------------------------------------------------------------------------------------------------------------
      GPU Id    UUID                     Accessibility       Container Ids
      ------------------------------------------------------------------------------------------------------------------------
      0         0xEA35F57CC80DEB35       Shared              -
      1         0x89CAA15875FF5A43       Exclusive           90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
      2         0x6E32F10EFC982B4C       Exclusive           -
      3         0x12FE4F7FDAF06B9        Exclusive           90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd

      > amd-ctk gpu-tracker 1 shared
      GPUs [1] have been made shared

      > docker run --runtime=amd -itd -e AMD_VISIBLE_DEVICES=0-2 rocm/rocm-terminal bash
      a8ce87c99727107ab467508bd431a170b148001fe8a866fcf96d5cc6af9a7f5e

      > amd-ctk gpu-tracker status
      ------------------------------------------------------------------------------------------------------------------------
      GPU Id    UUID                     Accessibility       Container Ids
      ------------------------------------------------------------------------------------------------------------------------
      0         0xEA35F57CC80DEB35       Shared              a8ce87c99727107ab467508bd431a170b148001fe8a866fcf96d5cc6af9a7f5e
      1         0x89CAA15875FF5A43       Shared              90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
                                                             a8ce87c99727107ab467508bd431a170b148001fe8a866fcf96d5cc6af9a7f5e
      2         0x6E32F10EFC982B4C       Exclusive           a8ce87c99727107ab467508bd431a170b148001fe8a866fcf96d5cc6af9a7f5e
      3         0x12FE4F7FDAF06B9        Exclusive           90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd
      ```

      In the above example, GPU 1 has been set to `shared` access from the previous `exclusive` access.

      When a new container `a8ce87c99727107ab467508bd431a170b148001fe8a866fcf96d5cc6af9a7f5e` that requests access to GPUs 0,1 and 2 is launched, the following happens:
      - The new container is created.
      - The new container is granted access to GPU 0 as no container is currently using GPU 0.
      - GPUs 1 is already being used by container `90cb29e11e83aa3ae497c68c90e1f0894b85262188c1ef9c7284457a9bc35ffd`. However, the new container is granted access to GPU 1 as GPU 1 has `shared` accessibility.
      - The new container is granted access to GPU 2 as no container is currently using GPU 2 though GPU 2 has `exclusive` accessibility.
      - The container is successfully started since it has been granted access to the required GPU resources.

  7. Resetting GPU Tracker Status:

      Resetting GPU Tracker clears the GPU Tracker state, i.e. the accessibility of all GPU is set to `shared` and all information about which GPUs have been made accessible in containers is cleared. If GPU Tracker is enabled, then after the reset operation also the GPU Tracker is enabled. Conversely, if GPU Tracker is cdisabled, then after the reset operation also the GOU Tracker remains disabled.

      Resetting GPU Tracker is primarily useful in cases where GPU Tracker is enabled and the partitioning scheme of the GPUs has been altered. Changing the partitioning scheme of the GPUs invalidated the CDI Spec and GPU Tracker state. In these cases, it is required to:
      - Stop all running containers
      - Reset GPU Tracker
      - Regenerate CDI Spec
      - Restart containers

      If GPU Tracker is disabled when the partitioning scheme of the GPUs have been altered, then GPU Tracker need not be reset. However, it is recommended to still perform the other actions. It makes no difference if GPU Tracker is reset when it is disabled.

      ```text
      > amd-ctk gpu-tracker status
      GPUs info is invalid. Please reset GPU Tracker.

      > amd-ctk gpu-tracker reset
      GPU Tracker has been reset
      Since GPU Tracker was enabled, it is recommended to stop and restart running containers to get the most accurate GPU Tracker status

      > sudo amd-ctk cdi generate
      Generated CDI spec: /etc/cdi/amd.json

      > sudo docker run --runtime=amd -itd -e AMD_VISIBLE_DEVICES=0-2 rocm/rocm-terminal bash
      988135dafcd94bf98fbd92ca97f4a07c9bcfff0521359ee9bc8a6973cc3e25ce

      > sudo amd-ctk gpu-tracker status
      ------------------------------------------------------------------------------------------------------------------------
      GPU Id    UUID                     Accessibility       Container Ids
      ------------------------------------------------------------------------------------------------------------------------
      0         0xEA35F57CC80DEB35       Shared              988135dafcd94bf98fbd92ca97f4a07c9bcfff0521359ee9bc8a6973cc3e25ce
      1         0x89CAA15875FF5A43       Shared              988135dafcd94bf98fbd92ca97f4a07c9bcfff0521359ee9bc8a6973cc3e25ce
      2         0x6E32F10EFC982B4C       Shared              988135dafcd94bf98fbd92ca97f4a07c9bcfff0521359ee9bc8a6973cc3e25ce
      3         0x12FE4F7FDAF06B9        Shared              -
      ```


# Release notes
| Release  | Features                                                                     | Known Issues                                                                 |
|----------|------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| v1.2.0   | 1. GPU Tracker feature support<br>2. Docker Swarm support | None                                                                         |
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


