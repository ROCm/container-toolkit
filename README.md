# AMD Container Toolkit

AMD Container Toolkit offers tools to streamline the use of AMD GPUs with containers. The toolkit includes the following packages:

- `amd-container-runtime` — The AMD container runtime
- `amd-ctk` — The AMD Container Toolkit CLI

## Key Features

- **Docker runtime integration** — Inject AMD GPUs into containers via the `amd-container-runtime` OCI runtime
- **CDI support** — Generate and manage [Container Device Interface](docs/container-runtime/cdi-guide.rst) specs for runtime-agnostic GPU access
- **GPU selection** — Target GPUs by index, range, or UUID using `AMD_VISIBLE_DEVICES`
- **Docker Swarm & Compose** — Orchestrate GPU workloads across nodes with UUID-based scheduling
- **GPU Tracker** — Lightweight opt-in tracking of container-to-GPU assignments with shared/exclusive access modes

## Documentation

For comprehensive documentation including installation, configuration, CDI, Swarm, troubleshooting, and migration guides, see the [official documentation](https://instinct.docs.amd.com/projects/container-toolkit/en/latest).

## Requirements

- Ubuntu 22.04, 24.04, or 26.04, or RHEL 9
- Docker version 25 or later
- All `amd-ctk runtime configure` commands should be run as root/sudo

> **Note:** Docker Desktop on Linux is not supported for GPU workloads; see [troubleshooting](docs/container-runtime/troubleshooting.rst) for details.

## Quick Start

Install the toolkit on Ubuntu (for RHEL 9, the AMDGPU driver, and full details, see [Installation](docs/container-runtime/installation.rst)):

1. Install prerequisites and add the repository:
   ```bash
   sudo apt update && sudo apt install -y wget gpg
   sudo mkdir --parents --mode=0755 /etc/apt/keyrings
   wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | sudo tee /etc/apt/keyrings/rocm.gpg > /dev/null
   echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ $(. /etc/os-release && echo $VERSION_CODENAME) main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list
   ```

2. Install the container toolkit:
   ```bash
   sudo apt update && sudo apt install amd-container-toolkit
   ```

3. Configure Runtime and run a GPU container:

   **Option A: CDI (recommended, runtime-agnostic, no runtime configure needed on Docker 28.2+):**
   ```bash
   sudo amd-ctk cdi generate
   sudo amd-ctk cdi validate
   ```
   Verify by running a container with all AMD GPUs:
   ```bash
   docker run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi
   ```

   > **Note:** CDI is supported by many container runtimes including Docker, Podman, and containerd. Docker 28.2 and newer enable CDI by default. On Docker 25.0 to 28.1, add `"features": {"cdi": true}` to `/etc/docker/daemon.json` (or run `sudo amd-ctk runtime configure`) and restart Docker.

   **Option B: AMD container runtime (Docker only):**
   ```bash
   sudo amd-ctk runtime configure
   sudo systemctl restart docker
   ```
   Verify by running a container with all AMD GPUs:
   ```bash
   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/dev-ubuntu-24.04 amd-smi
   ```

## Usage

### Using CDI

List available CDI device entries:

```bash
amd-ctk cdi list
```

Select GPUs by index using `--device amd.com/gpu=<entry>`:

```bash
docker run --rm --device amd.com/gpu=0 rocm/dev-ubuntu-24.04 amd-smi
docker run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi
```

### Using amd-container-runtime (Docker only)

Select GPUs by index, range, or UUID with `AMD_VISIBLE_DEVICES`:

```bash
docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0,1,2 rocm/dev-ubuntu-24.04 amd-smi
docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0-3,5,8 rocm/dev-ubuntu-24.04 amd-smi
docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0xEF2C1799A1F3E2ED rocm/dev-ubuntu-24.04 amd-smi
```

List available GPUs and their UUIDs:

```bash
amd-ctk gpu list
```

This will display output similar to:
```
Found 2 AMD GPU devices
---------------------------------------------------------------------------
GPU Id    UUID                     DRM Devices
---------------------------------------------------------------------------
0         0xEF2C1799A1F3E2ED       /dev/dri/renderD128
1         0x1234567890ABCDEF       /dev/dri/renderD129
```

Set the AMD container runtime as Docker's default (avoids needing `--runtime=amd`):

```bash
sudo amd-ctk runtime configure --set-as-default
sudo systemctl restart docker
```

For more on specific topics, see the detailed documentation:

- [Running Workloads](docs/container-runtime/running-workloads.rst) — CDI with each runtime, `AMD_VISIBLE_DEVICES`, `--gpus` flag
- [GPU Partitioning](docs/container-runtime/gpu-partitioning.rst) — Running containers on GPU partitions
- [CDI Guide](docs/container-runtime/cdi-guide.rst) — Generating, validating, and managing CDI specs
- [Docker Compose](docs/container-runtime/docker-compose.rst) — Multi-service GPU access
- [Docker Swarm](docs/container-runtime/docker-swarm.md) — UUID-based GPU scheduling across nodes
- [GPU Tracker](docs/container-runtime/gpu-tracker.md) — Shared/exclusive GPU access tracking

## Building from Source

To build a Debian package:

```bash
make && make pkg-deb
```

To build an RPM package:

```bash
make build-dev-container-rpm && make pkg-rpm
```

Packages are generated in the `bin` folder.

## Release Notes

See the [Release Notes](docs/container-runtime/release-notes.rst) for version history, compatibility matrix, and upgrade notes.

## License

This project is licensed under the Apache 2.0 License — see the [LICENSE](https://github.com/ROCm/container-toolkit/blob/main/LICENSE) file for details.
