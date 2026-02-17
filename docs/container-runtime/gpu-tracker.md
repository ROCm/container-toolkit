# GPU Tracker

Currently, barebones Docker provides no way to track access of GPUs in containers. Additionally, by default, multiple containers in Docker can be granted access to the same GPU simultaneously. GPU Tracker is an extremely lightweight feature of AMD Container Toolkit that solves these issues.

GPU Tracker state is initialized during AMD Container Toolkit installation and is by default disabled. Users can enable or disable the GPU Tracker feature by using the `enable` or `disable` CLIs. When enabled, the GPU Tracker automatically maintains the state of GPUs and the containers that they are made accessible to, only if the containers are launched and granted access to the GPUs using the `AMD_VISIBLE_DEVICES` environment variable. When the container process completes execution or is stopped, the GPU Tracker state is automatically updated to reflect GPUs released by the specific container.

**NOTE:** GPU Tracker feature is currently supported only if containers are started using the `docker run` command and GPUs are made accessible in containers using the `AMD_VISIBLE_DEVICES` environment variable. If containers are started and granted access to GPUs in any other manner, GPU Tracker feature is not supported.

GPU Tracker provides CLIs that can be used to control the accessibility of GPUs in containers. The accessibility of GPUs can be set to either `shared` or `exclusive`.
- The `shared` accessibility indicates that the GPU can be made accessible to multiple containers simultaneously. By default, all GPUs are granted the `shared` accessibility to reflect the default Docker behavior.
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

      Resetting GPU Tracker clears the GPU Tracker state, i.e. the accessibility of all GPUs is set to `shared` and all information about which GPUs have been made accessible in containers is cleared. If GPU Tracker is enabled, then after the reset operation also the GPU Tracker is enabled. Conversely, if GPU Tracker is disabled, then after the reset operation also the GPU Tracker remains disabled.

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

