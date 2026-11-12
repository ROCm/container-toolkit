# AMD Container Toolkit

The AMD Container Toolkit gives containers access to AMD GPUs. It generates Container Device Interface (CDI) specifications for Docker, Podman, and containerd, and provides an OCI runtime for Docker.

[![CI](https://github.com/ROCm/container-toolkit/actions/workflows/ci.yml/badge.svg)](https://github.com/ROCm/container-toolkit/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ROCm/container-toolkit)](https://github.com/ROCm/container-toolkit/releases)
[![Docs](https://img.shields.io/badge/docs-instinct.docs.amd.com-blue)](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/)

The `amd-container-toolkit` package installs two binaries:

- `amd-ctk`: the command-line tool that lists GPUs, manages CDI specifications, configures Docker, and controls the GPU Tracker.
- `amd-container-runtime`: the AMD container runtime for Docker.

## Features

- **CDI specifications**: `amd-ctk cdi generate` describes every GPU for any CDI-aware runtime. See the [CDI Guide](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/cdi-guide.html).
- **AMD container runtime**: select GPUs in Docker by index, range, or UUID with `AMD_VISIBLE_DEVICES`, or with `--gpus` on Docker 28.3 or newer.
- **GPU partitioning**: assign individual GPU partitions to containers.
- **GPU Tracker**: opt-in shared or exclusive GPU access across containers that use the AMD container runtime. See [GPU Tracker](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/gpu-tracker.html).
- **Orchestration**: Docker Compose and Docker Swarm support.

## Requirements

- Ubuntu 22.04, 24.04, or 26.04, or RHEL 9
- AMDGPU driver installed on the host ([AMD GPU Driver installation guide](https://instinct.docs.amd.com/projects/amdgpu-docs/en/latest/install/package-manager-index.html))
- Docker 25.0 or newer, or another CDI-aware runtime

See [System Requirements](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/requirements.html) for container runtime versions and limits.

## Quick Start

On Ubuntu:

```bash
sudo apt update && sudo apt install -y wget gpg
sudo mkdir --parents --mode=0755 /etc/apt/keyrings
wget https://repo.radeon.com/rocm/rocm.gpg.key -O - | gpg --dearmor | sudo tee /etc/apt/keyrings/rocm.gpg > /dev/null
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/rocm.gpg] https://repo.radeon.com/amd-container-toolkit/apt/ $(. /etc/os-release && echo $VERSION_CODENAME) main" | sudo tee /etc/apt/sources.list.d/amd-container-toolkit.list
sudo apt update && sudo apt install -y amd-container-toolkit

sudo amd-ctk cdi generate
docker run --rm --device amd.com/gpu=all rocm/dev-ubuntu-24.04 amd-smi
```

On Docker 25.0 to 28.1, run `sudo amd-ctk runtime configure && sudo systemctl restart docker` before `docker run` to enable CDI.

For RHEL and the AMD container runtime, see [Installation](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/installation.html) and [Quick Start](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/quick-start-guide.html).

## Documentation

- [Running Workloads](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/running-workloads.html): GPU selection with Docker, Podman, nerdctl, and ctr
- [GPU Partitioning](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/gpu-partitioning.html)
- [Docker Compose](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/docker-compose.html) and [Docker Swarm](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/docker-swarm.html)
- [Migrating from NVIDIA](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/migration-guide.html)
- [Troubleshooting](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/troubleshooting.html)
- [Release Notes](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/release-notes.html)

## Building from Source

```bash
make                                          # build the binaries in the builder container
make pkg-deb                                  # build the Debian package into bin/
make build-dev-container-rpm && make pkg-rpm  # build the RPM package into bin/
make test                                     # run the unit tests (needs Go on the host)
```

See the [Developer Guide](https://instinct.docs.amd.com/projects/container-toolkit/en/latest/container-runtime/developer-guide.html) for host builds and contribution guidelines.

## Support

Report bugs and request features in [GitHub Issues](https://github.com/ROCm/container-toolkit/issues). Report security vulnerabilities privately; see [SECURITY.md](SECURITY.md).

## License

Apache 2.0. See [LICENSE](LICENSE).
