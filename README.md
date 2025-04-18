# Container Toolkit
Offers tools that streamline the use of AMD GPUs with containers.

## Building
To build debian package, use the following command.

```text
make pkg-deb
```

To build rpm package, use the following command.

```text
make pkg-rpm
```

The packages will be generated in the ```bin``` folder.

## Docker Usage
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