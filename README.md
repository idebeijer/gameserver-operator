<div align="center">

  <h1>GameServer Operator</h1>

**A Kubernetes operator for managing game servers with security and simplicity in mind**

[![License](https://img.shields.io/github/license/idebeijer/gameserver-operator?style=for-the-badge)](https://github.com/idebeijer/gameserver-operator/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/idebeijer/gameserver-operator?style=for-the-badge)](https://github.com/idebeijer/gameserver-operator/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/idebeijer/gameserver-operator?style=for-the-badge)](https://goreportcard.com/report/github.com/idebeijer/gameserver-operator)
[![CI](https://img.shields.io/github/actions/workflow/status/idebeijer/gameserver-operator/test.yml?branch=main&style=for-the-badge)](https://github.com/idebeijer/gameserver-operator/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/idebeijer/gameserver-operator?style=for-the-badge)](https://go.dev/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.22+-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)](https://kubernetes.io/)

</div>

## Overview

GameServer Operator provides a single Kubernetes API for deploying and managing more than 270 game servers through [LinuxGSM](https://linuxgsm.com/). A `GameServer` custom resource defines the game, storage, and networking while the operator handles container lifecycle, configuration, and service exposure.

## Features

- **Universal game support** – Deploy any LinuxGSM-supported game server using a single CRD.
- **Security-first defaults** – Non-root user, restricted capabilities, seccomp profile, and no privilege escalation.

## Requirements

- Kubernetes 1.22+ (operator uses [Server-Side Apply](https://kubernetes.io/docs/reference/using-api/server-side-apply/)).
- Helm 3.8+.
- `kubectl` configured for the target cluster.

## Installation

### Helm (recommended)

```bash
helm install gameserver-operator \
  oci://ghcr.io/idebeijer/charts/gameserver-operator \
  --namespace gameserver-operator-system \
  --create-namespace
```

### From source

```bash
make docker-build docker-push IMG=<registry>/gameserver-operator:<tag>
make deploy IMG=<registry>/gameserver-operator:<tag>
```

The first command builds and pushes an image that matches any local changes; the second command deploys that image and installs the CRDs into the target cluster.

## Quick Start

1. Install the operator using the [Helm chart](#helm-recommended) or `make deploy`.
2. Apply a `GameServer` manifest that describes the game you want to run.

### Example: Minecraft server

```yaml
apiVersion: games.idebeijer.github.io/v1alpha1
kind: GameServer
metadata:
  name: my-minecraft-server
spec:
  gameName: mc
  service:
    type: LoadBalancer
    ports:
      - name: game
        port: 25565
        protocol: TCP
  storage:
    size: 20Gi
```

Apply these manifests with `kubectl apply -f <file>` (or `-` for stdin).

## Development Workflow

```bash
make install        # Install or update CRDs in the active kubecontext
make run            # Run the controller locally against cluster in kubeconfig
make deploy IMG=... # Build manifests and deploy the manager to the cluster
make undeploy       # Remove the manager and associated resources
```

More details in [CONTRIBUTING.md](./CONTRIBUTING.md).

## Documentation

Additional examples and configuration snippets live in the [`config/samples`](./config/samples) directory. The [`pkg/specs`](./pkg/specs) package contains the programmatic representation of supported options.

## Known Limitations

- Some actions might be unsupported due to security restrictions. To name a few:
  - Binding to ports below 1024.
  - Games that require `CAP_NET_ADMIN` or `CAP_SYS_ADMIN` will not work due to dropped capabilities.
  - LinuxGSM cron-based automations (updates, backups) are disabled under the default security profile. (working on a k8s native solution)

Those limitations are intentional to maintain a secure default. If it appears that a game cannot run under these restrictions, please open an issue.
Otherwise, in the future the `GameServer` spec may include options for customizing security settings per game.

## Stability and API Changes

⚠️ **Major version zero (v0.x.x).** The CRDs and APIs may change between releases and backwards compatibility is not guaranteed.

## Roadmap

- [ ] Automatic Minecraft modded server setup (CurseForge).
- [ ] CLI tool for managing game servers.
- [ ] Web UI for server management.
- [ ] Command forwarding service (native LinuxGSM commands).
- [ ] Configurable security contexts per game. (if it turns out some games need it)
- [ ] Native backup support using CronJobs.
- [ ] Auto-update scheduling.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and guidelines.

## License

Licensed under the MIT License. See [LICENSE](./LICENSE) for details.
