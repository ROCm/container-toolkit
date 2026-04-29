# AMD Container Toolkit

AMD Container Toolkit offers tools to streamline the use of AMD GPUs with containers. The toolkit includes the following packages:

- `amd-container-runtime` — The AMD Container Runtime
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

- Ubuntu 22.04 or 24.04, or RHEL/CentOS 9
- Docker version 25 or later
- All `amd-ctk runtime configure` commands should be run as root/sudo

> **Note:** Docker Desktop on Linux is not supported for GPU workloads; see [troubleshooting](docs/container-runtime/troubleshooting.rst) for details.

## Quick Start

Install the toolkit on Ubuntu (for RHEL/CentOS and full details, see the [Quick Start Guide](docs/container-runtime/quick-start-guide.rst)):

1. Install prerequisites and add the repository:
   ```bash
   sudo apt update && sudo apt install -y wget gnupg2
   wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | sudo tee /etc/apt/keyrings/rocm.gpg > /dev/null
   echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ $(. /etc/os-release && echo $VERSION_CODENAME) main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list
   ```

2. Install the container toolkit:
   ```bash
   sudo apt update && sudo apt install amd-container-toolkit
   ```

3. Configure Runtime and run a GPU container:

   **Option A — AMD container runtime:**
   ```bash
   sudo amd-ctk runtime configure
   sudo systemctl restart docker
   ```
   Verify by running a container with all AMD GPUs:
   ```bash
   docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=all rocm/rocm-terminal rocm-smi
   ```

   **Option B — CDI (runtime-agnostic, no runtime configure needed):**
   ```bash
   sudo amd-ctk cdi generate --output=/etc/cdi/amd.json
   sudo amd-ctk cdi validate --path=/etc/cdi/amd.json
   ```
   Verify by running a container with all AMD GPUs:
   ```bash
   docker run --rm --device amd.com/gpu=all rocm/rocm-terminal rocm-smi
   ```

   > **Note:** CDI is supported by many container runtimes including Docker, Podman, and containerd.

## Usage

Select specific GPUs by index, range, or UUID:

```bash
docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0,1,2 rocm/rocm-terminal rocm-smi
docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0-3,5,8 rocm/rocm-terminal rocm-smi
docker run --rm --runtime=amd -e AMD_VISIBLE_DEVICES=0xEF2C1799A1F3E2ED rocm/rocm-terminal rocm-smi
```

List available GPUs and their UUIDs for use with `AMD_VISIBLE_DEVICES`:

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

Set the AMD runtime as Docker's default (avoids needing `--runtime=amd`):

```bash
sudo amd-ctk runtime configure --runtime=docker --set-as-default
```

For more on specific topics, see the detailed documentation:

- [Running Workloads](docs/container-runtime/running-workloads.rst) — CDI, `--gpus` flag, explicit device paths, GPU partitioning
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
