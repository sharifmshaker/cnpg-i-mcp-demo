# cnpg-i-mcp-demo

A demo [CNPG-I](https://github.com/cloudnative-pg/cnpg-i) plugin for [CloudNativePG](https://cloudnative-pg.io) that attaches a **Model Context Protocol (MCP) server** to a Postgres cluster, letting an AI agent (e.g. Claude) query the database directly through the operator's plugin interface.

It accompanies the talk **"Extending Cloud-Native Postgres with CNPG-I Plugins"** ([SCaLE 23x](https://www.socallinuxexpo.org/scale/23x), March 2026).

📝 **Written companion:** [Extending CloudNativePG with the CNPG-I Plugin System](https://sharebearbeta.com/posts/extending-cloudnativepg-cnpg-i-plugin-system.html) — a walkthrough of the plugin interface and how the pieces below fit together.

## What it demonstrates

- How to build a CNPG-I plugin that implements the **Identity**, **Lifecycle**, and **Reconciler** gRPC services.
- Using **Lifecycle** hooks to inject a sidecar (the MCP server) into Postgres pods and modify the cluster's services.
- Exposing a running Postgres cluster to an AI agent over MCP, with no changes to the core operator.

## Architecture

The plugin registers with the CloudNativePG operator over a gRPC / gRPC-streaming interface and:

- **Identity** (`internal/identity`) — advertises the plugin's name, capabilities, and metadata.
- **Lifecycle** (`internal/lifecycle`) — injects an MCP server sidecar into Postgres pods and wires up the connection.
- **Reconciler** (`internal/reconciler`) — manages the supporting service so the MCP endpoint is reachable.

```
main.go
├── cmd/plugin           # cobra command wiring the gRPC server
├── internal/identity    # plugin registration + capabilities
├── internal/lifecycle   # pod/service mutation (sidecar injection)
├── internal/reconciler  # service reconciliation
├── internal/config      # plugin parameter parsing
├── kubernetes/          # kustomize manifests (deployment, RBAC, cert-manager certs)
└── doc/                 # runnable demo steps + example Cluster manifests
```

## Prerequisites

- Go 1.25+
- Docker
- [kind](https://kind.sigs.k8s.io/) (local Kubernetes)
- CloudNativePG installed in the cluster ([cnpg-playground](https://github.com/cloudnative-pg/cnpg-playground) is handy)
- [cert-manager](https://cert-manager.io) (the plugin uses mTLS between the operator and the plugin server)

## Quick start

Full, copy-pasteable steps are in [`doc/demo/steps.md`](./doc/demo/steps.md). In short:

```bash
# Build and load the plugin image into your kind cluster
docker build -t cnpg-i-mcp-demo:latest .
kind load docker-image --name k8s-local cnpg-i-mcp-demo:latest

# Deploy the plugin and an example cluster
kubectl apply -k kubernetes/
kubectl apply -f doc/examples/cluster-example.yaml

# Port-forward the MCP endpoint and register it with Claude
kubectl port-forward svc/pg-local-rw 8888:8888
claude mcp add postgres-local --transport sse http://localhost:8888/sse
```

Then ask the agent about your data, e.g. *"How many customers do I have in postgres-local?"*

> **Note:** the `DATABASE_URI` in `postgres-mcp-deployment.yaml` is a placeholder — replace it with your cluster's connection string before deploying.

> ⚠️ **Demo only:** the MCP sidecar runs in `--access-mode=unrestricted` and connects as the `postgres` superuser, so the AI agent can execute arbitrary read/write SQL. This is intentional for the demo. In production, connect with a least-privilege role and use a restricted access mode.

## License

Apache License 2.0 — see [`LICENSE`](./LICENSE).
