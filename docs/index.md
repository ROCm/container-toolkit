---
myst:
  html_meta:
    "description": "AMD Container Toolkit documentation"
    "keywords": "AMD, ROCm, Container Runtime, GPU, HPC, Machine Learning"
---

# AMD Container Toolkit Documentation

The AMD Container Toolkit gives containers access to AMD Instinct GPUs. It generates Container Device Interface (CDI) specifications for any CDI-aware runtime and provides an OCI runtime for Docker, so ROCm workloads run in containers without manual device setup.

## Get started

1. Read the [Overview](container-runtime/overview.rst) to choose between CDI and the AMD container runtime.
2. Check the [System Requirements](container-runtime/requirements.rst).
3. Follow the [Installation](container-runtime/installation.rst) and [Quick Start](container-runtime/quick-start-guide.rst).

Coming from the NVIDIA Container Toolkit? See [Migrating from NVIDIA](container-runtime/migration-guide.rst).

## Run workloads

- [Running Workloads](container-runtime/running-workloads.rst) with Docker, Podman, nerdctl, and ctr
- [Container Device Interface](container-runtime/cdi-guide.rst)
- [GPU Partitioning](container-runtime/gpu-partitioning.rst)
- [GPU Tracker](container-runtime/gpu-tracker.md) for shared and exclusive GPU access
- [Docker Compose](container-runtime/docker-compose.rst) and [Docker Swarm](container-runtime/docker-swarm.md)
- [Framework Integration](container-runtime/framework-integration.rst)

## Help and reference

- [Troubleshooting](container-runtime/troubleshooting.rst)
- [Migrating from NVIDIA](container-runtime/migration-guide.rst)
- [Developer Guide](container-runtime/developer-guide.rst)
- [Release Notes](container-runtime/release-notes.rst)
