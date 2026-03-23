# Enterprise Internal Developer Platform (IDP)

Open-source platform for accelerating service and ML delivery with built-in governance, security, and compliance.

## Architecture

- **Core Platform**: Catalog, Workflows, Actions (Go/Python)
- **ML Platform**: Model Registry, Adapters (Python/MLflow)
- **Developer Tools**: Portal (React), CLI (Go)
- **Infrastructure**: Kubernetes, ArgoCD, Tekton, Terraform
- **Observability**: Prometheus, Grafana, Loki, Tempo

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Identity | Keycloak |
| API Gateway | Kong |
| Workflows | Temporal |
| Database | PostgreSQL |
| Message Queue | NATS |
| Event Streaming | Kafka |
| ML Registry | MLflow |
| GitOps | ArgoCD |
| Pipelines | Tekton + GitLab CI |
| Observability | Prometheus Stack |

## Quick Start
```bash
# Prerequisites
- Kubernetes cluster (k3s, EKS, GKE, AKS)
- kubectl configured
- helm 3.x
- terraform 1.5+

# 1. Deploy infrastructure
cd infrastructure/terraform/environments/dev
terraform init
terraform apply

# 2. Deploy platform services
kubectl apply -f infrastructure/argocd/projects/
kubectl apply -f infrastructure/argocd/applications/

# 3. Access portal
kubectl port-forward -n platform svc/portal 3000:80
# Open http://localhost:3000
```

## Documentation

See [docs/](./docs) for comprehensive documentation.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)

## License

MIT License - See [LICENSE](./LICENSE)
