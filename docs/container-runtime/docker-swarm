# Docker Swarm Integration

### Purpose

Docker Swarm integration allows orchestrated GPU workloads to be deployed across multiple nodes by leveraging **GPU UUIDs** and Docker’s **generic resource** framework.

This ensures consistent GPU assignment, regardless of hardware order or node topology.

---

### Docker Daemon Configuration for Swarm

Configure each swarm node's Docker daemon with GPU resources in `/etc/docker/daemon.json`:

```json
{
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
### Deploy GPU Enabled Services

Deploy services with specific GPU requirements using docker-compose:

```yaml
# docker-compose.yml for Swarm deployment
version: '3.8'
services:
  rocm-service:
    image: rocm/dev-ubuntu-24.04
    command: rocm-smi
    runtime: amd
    deploy:
      replicas: 1
      resources:
        reservations:
          generic_resources:
            - discrete_resource_spec:
                kind: 'AMD_GPU'  # Matches daemon.json key
                value: 1
```

Deploy the service:
```bash
docker stack deploy -c docker-compose.yml rocm-stack
```


**Benefits of Swarm Integration**

**UUID-Based Scheduling**: Prevents mismatched device mapping across nodes.

**Cluster-Aware Resource Allocation**: Swarm recognizes GPU availability and schedules workloads accordingly.

**Scalable GPU Management**: Simplifies distributed workloads using multiple GPUs or nodes.

**Seamless Orchestration**: Integrates with Docker’s built-in scheduling and scaling mechanisms.
