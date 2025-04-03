# container-toolkit
Offers tools that streamline the use of AMD GPUs with containers.

## Docker Usage
1. Configure Docker to use AMD container runtime.

``` text
> amd-ctk runtime configure --runtime=docker
```

2. Specify the required GPUs. There are 3 ways to do this.

2.1. Using AMD_VISIBLE_DEVICES environment variable.

To use all available GPUs,

```text
> docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi
```

To use a subset of available GPUs,

```text
> docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0,1,2 rocm/rocm-terminal rocm-smi
```

2.2. Using CDI style

To use all available GPUs,

```text
> docker run --rm --runtime=amd --device amd.com/gpu=all rocm/rocm-terminal rocm-smi
```

To use a subset of available GPUs,

```text
> docker run --rm --runtime=amd --device amd.com/gpu=0 --device amd.com/gpu=1 rocm/rocm-terminal rocm-smi
```

2.3. Using explicit paths for each required GPU

```text
> docker run --device /dev/kfd --device /dev/dri/renderD128 --device /dev/dri/renderD129 rocm/rocm-terminal rocm-smi
```

3. To see the list of all available GPUs,

```text
> amd-ctk gpu list
Found 1 AMD GPU device
amd.com/gpu=all
amd.com/gpu=0
  /dev/dri/card1
  /dev/dri/renderD128
```

4. To avoid specifying "--runtime=amd" option with the "docker run" command, set the AMD container runtime as the default for Docker.

```text
> amd-ctk runtime configure --runtime=docker --set-as-default
```
