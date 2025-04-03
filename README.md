# container-toolkit
Offers tools that streamline the use of AMD GPUs with containers.

## Docker Usage
### Configure Docker to use AMD container runtime.

``` text
> amd-ctk runtime configure --runtime=docker
```

### Specify the required GPUs. There are 3 ways to do this.

#### AMD_VISIBLE_DEVICES Environment Variable.

To use all available GPUs,

```text
> docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi
```

To use a subset of available GPUs,

```text
> docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0,1,2 rocm/rocm-terminal rocm-smi
```

#### CDI Style

To use all available GPUs,

```text
> docker run --rm --runtime=amd --device amd.com/gpu=all rocm/rocm-terminal rocm-smi
```

To use a subset of available GPUs,

```text
> docker run --rm --runtime=amd --device amd.com/gpu=0 --device amd.com/gpu=1 rocm/rocm-terminal rocm-smi
```

#### Explicit Paths

```text
> docker run --device /dev/kfd --device /dev/dri/renderD128 --device /dev/dri/renderD129 rocm/rocm-terminal rocm-smi
```

### List available GPUs.

```text
> amd-ctk gpu list
Found 1 AMD GPU device
amd.com/gpu=all
amd.com/gpu=0
  /dev/dri/card1
  /dev/dri/renderD128
```

### Make AMD container runtime default runtime.
Avoid specifying "--runtime=amd" option with the "docker run" command by setting the AMD container runtime as the default for Docker.

```text
> amd-ctk runtime configure --runtime=docker --set-as-default
```
