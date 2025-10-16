# Contributing to GameServer Operator

Thank you for your interest in contributing to GameServer Operator! This document provides guidelines and instructions for setting up your development environment and contributing to the project.

## Development Prerequisites

- Go version 1.24.0+
- Docker
- Access to a Kubernetes v1.22+ cluster (or use [kind](https://kind.sigs.k8s.io/) or [minikube](https://minikube.sigs.k8s.io/))
- Make

## Getting Started

1. **Clone the repository**

```bash
git clone https://github.com/idebeijer/gameserver-operator.git
cd gameserver-operator
```

2. **Make**

The project relies on `make` for various tasks. You can view available commands with:

```bash
make help
```

3. **Run tests**

```bash
make test        # Run unit tests
make test-e2e    # Run end-to-end tests (requires kind)
```

## Development Workflow

### Code Changes

When making changes to the API:

```bash
# Update generated code
make generate

# Update CRD manifests
make manifests
```

### Running the Controller Locally

For rapid development, you can run the controller locally against a Kubernetes cluster:

```bash
# Install CRDs into the cluster (uses kubectl; requires kubeconfig to be set to target cluster)
make install

# Run the controller locally (requires kubeconfig to be set to target cluster)
make run
```

### Building and Deploying

To build and deploy the controller to a cluster:

```bash
# Build and push the Docker image
make docker-build docker-push IMG=<your-registry>/gameserver-operator:tag

# Deploy to the cluster
make deploy IMG=<your-registry>/gameserver-operator:tag
```

### Testing Your Changes in a Cluster

1. **Create a test GameServer**

```bash
kubectl apply -f config/samples/games_v1alpha1_gameserver.yaml
```

2. **Check the logs**

```bash
kubectl logs -n gameserver-operator-system deployment/gameserver-operator-controller-manager -f
```

3. **Verify resources**

```bash
kubectl get gameserver
kubectl get statefulset
kubectl get service
```

### Cleanup

To remove the controller and CRDs from your cluster:

```bash
# Delete sample instances
kubectl delete -k config/samples/

# Undeploy the controller
make undeploy

# Uninstall CRDs
make uninstall
```

## Project Distribution

### Building the Installer Bundle

Generate a single YAML file containing all resources:

```bash
make build-installer IMG=<your-registry>/gameserver-operator:tag
```

This creates `dist/install.yaml` which users can apply directly:

```bash
kubectl apply -f dist/install.yaml
```

### Building the Helm Chart

The project uses Kubebuilder's Helm plugin to generate Helm charts:

```bash
kubebuilder edit --plugins=helm/v1-alpha
```

The generated chart is located in `dist/chart/`.

**Note:** After making changes to the project:

- Ensure to run `make manifests` to update CRDs.
- Regenerate the Helm chart (using `kubebuilder edit --plugins=helm/v1-alpha` again if necessary).

The following files will not be updates unless `--force` is used:

- `dist/chart/values.yaml`
- `dist/chart/templates/manager/manager.yaml`

The following files are never updated automatically and must be maintained manually:

- `dist/chart/Chart.yaml`
- `dist/chart/templates/_helpers.tpl`
- `dist/chart/.helmignore`

## Code Style

- Follow standard Go formatting (`gofmt`)
- Run linter before committing: `make lint`
- Add meaningful comments for exported functions and types
- Write tests for new functionality

## Commit Messages

Follow conventional commit format, e.g.:

- `feat: add new feature`
- `fix: resolve bug`
- `docs: update documentation`
- `test: add or update tests`
- `refactor: code refactoring`
- `chore: maintenance tasks`

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes
4. Run tests and linting (`make test lint`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to your fork (`git push origin feat/amazing-feature`)
7. Open a Pull Request

## Questions?

Feel free to open an issue for questions or discussion about the project.

## Additional Resources

- [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [LinuxGSM Documentation](https://docs.linuxgsm.com/)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
