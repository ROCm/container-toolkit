# Docker Swarm

## Purpose

Docker Swarm integration allows orchestrated GPU workloads to be deployed across multiple nodes by leveraging **GPU UUIDs** and Docker’s resource framework.

This guide assumes a Swarm cluster is already initialized (`docker swarm init` on the manager, `docker swarm join` on each worker) and the toolkit is installed on every GPU node.

## Docker Daemon Configuration for Swarm

Swarm services cannot select a runtime per service, so the AMD container runtime must be the default runtime on every GPU node. Register it as the default:

```bash
sudo amd-ctk runtime configure --set-as-default
```

List the GPU UUIDs on each node:

```bash
amd-ctk gpu list
```

Then advertise the GPUs as generic resources in `/etc/docker/daemon.json` on that node. Each entry has the form `<kind>=<GPU UUID>`:

```json
{
  "default-runtime": "amd",
  "features": {
    "cdi": true
  },
  "runtimes": {
    "amd": {
      "path": "amd-container-runtime",
      "args": []
    }
  },
  "node-generic-resources": [
    "AMD_GPU=0x378041e1ada6015",
    "AMD_GPU=0xef39dad16afb86ad"
  ]
}
```

When a task reserves a resource, Docker passes the assigned UUIDs to the container as `DOCKER_RESOURCE_<kind>`. The AMD container runtime reads any `DOCKER_RESOURCE_*` variable, so the kind name (`AMD_GPU` above) can be chosen freely as long as it matches the service definition.

After updating the configuration, restart the Docker daemon:
```bash
sudo systemctl restart docker
```
## Deploy GPU Enabled Services

Deploy services with specific GPU requirements using docker-compose:

**Using generic resources:**

```yaml
# docker-compose.yml for Swarm deployment
version: '3.8'
services:
  rocm-service:
    image: rocm/dev-ubuntu-24.04
    command: amd-smi
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

This form does not use Swarm scheduling. Every replica gets the GPUs listed in `AMD_VISIBLE_DEVICES` on the node it lands on.

```yaml
# docker-compose.yml for Swarm deployment with environment variable
version: '3.8'
services:
  rocm-service:
    image: rocm/dev-ubuntu-24.04
    command: amd-smi
    environment:
      - AMD_VISIBLE_DEVICES=all
    deploy:
      replicas: 1
```

Deploy the service:
```bash
docker stack deploy -c docker-compose.yml rocm-stack
```

Check that the task started and sees the GPUs:
```bash
docker service ps rocm-stack_rocm-service
docker service logs rocm-stack_rocm-service
```

If the service does not see the expected GPUs, see [Troubleshooting](troubleshooting.rst).
