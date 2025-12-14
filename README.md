# URL Shortener

A high-performance URL Shortener service built with Go and the Fiber framework. This project is designed for cloud-native environments, containerized with Docker, and orchestrated with Kubernetes. It includes a complete local development stack using k3d, with monitoring via Prometheus, and autoscaling based on custom metrics (HPA).

## Architecture Overview

The project follows a Hexagonal Architecture (Ports and Adapters) to ensure a clean separation of concerns, making it maintainable and testable.

- **Application**: A REST API built with Go and [Fiber](https://gofiber.io/), a high-performance web framework.
- **Containerization**: Multi-stage Docker builds for lean, production-ready images.
- **Orchestration**: Kubernetes manifests for deployment, service discovery, ingress, and autoscaling.
- **Local Cluster**: [k3d](https://k3d.io/) is used to run a lightweight local Kubernetes cluster.
- **Ingress**: [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/) manages external access and provides rate limiting.
- **Observability**:
  - **Prometheus**: Scrapes application metrics for monitoring.
  - **Prometheus Adapter**: Exposes custom metrics to the Kubernetes API for autoscaling.
- **Autoscaling**: A Horizontal Pod Autoscaler (HPA) automatically scales the application based on requests per second (QPS).

## Project Structure

```folder
.
├── cmd/                    # Main application entrypoint
│   └── http.go
├── internal/               # Core business logic and adapters
│   ├── adapters/           # Implementations of ports (e.g., HTTP handlers, repositories)
│   ├── core/               # Core domain, services, and ports (interfaces)
├── k8s/                    # Kubernetes manifests
│   ├── deployment.yaml     # App deployment
│   ├── service.yaml        # ClusterIP service
│   ├── ingress.yaml        # Ingress rules with rate limiting
│   ├── hpa.yaml            # Horizontal Pod Autoscaler
│   └── monitoring/
│       └── adapter-values.yaml # Prometheus Adapter configuration
├── template/               # Reusable template for new projects
├── ci.Dockerfile           # Dockerfile for the build environment
├── Dockerfile              # Multi-stage application Dockerfile
├── go.mod                  # Go module dependencies
├── k3d-config.yaml         # k3d cluster configuration
└── Makefile                # Automation scripts for development and deployment
```

## Prerequisites

Ensure you have the following tools installed on your system:

- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [k3d](https://k3d.io/) (v5.0.0 or newer)
- [Helm](https://helm.sh/)
- [Make](https://www.gnu.org/software/make/)

## Getting Started

A `Makefile` is provided to automate all the necessary steps. To spin up the entire environment, simply run:

```bash
make run
```

This single command will perform the following actions:

1. Create a local Kubernetes cluster using `k3d`.
2. Build the application's Docker image.
3. Import the image into the `k3d` cluster.
4. Install the NGINX Ingress Controller.
5. Install Prometheus and the Prometheus Adapter for custom metrics.
6. Deploy the URL shortener application and its associated Kubernetes resources.

### Accessing the Application

Once the deployment is complete, the application will be accessible at **<http://localhost:8081/>**.

The port is mapped from the `k3d` cluster to your local machine as defined in `k3d-config.yaml`.

### Cleaning Up

To delete the cluster and remove all created resources, run:

```bash
make clean
```

## API Endpoints

### Shorten URL

Creates a new short URL.

- **Endpoint**: `POST /api/v1`
- **Request Body**:

  ```json
  {
    "url": "https://your-long-url.com/with/a/very/long/path"
  }
  ```

- **Success Response** (`201 Created`):

  ```json
  {
    "id": "c2a7b3a0-8a1e-4b3a-9f3e-3e1e3a0c2a7b",
    "original_url": "https://your-long-url.com/with/a/very/long/path",
    "short_code": "aBcDeF",
    "created_at": "2025-12-14T10:00:00Z",
    "access_count": 0
  }
  ```

### Redirect to Original URL

Redirects to the original URL associated with a short code.

- **Endpoint**: `GET /{shortCode}`
- **Example**: `GET /aBcDeF`
- **Success Response**: `308 Permanent Redirect` to the original URL.
- **Error Response** (`404 Not Found`): If the short code does not exist.

### Metrics

Exposes application metrics for Prometheus.

- **Endpoint**: `GET /metrics`

## Key Features Explained

### Autoscaling (HPA)

The application is configured to scale automatically based on traffic.

- **Metric**: `http_requests_qps` (a custom metric representing requests per second).
- **Target**: 10 QPS per pod.
- **Replicas**: Scales between 1 (min) and 10 (max) pods.

The `prometheus-adapter` queries Prometheus for the `rate(fiber_http_requests_total[1m])` metric and exposes it to Kubernetes, allowing the HPA to function.

### Ingress and Rate Limiting

The NGINX Ingress Controller manages external traffic. To prevent abuse, rate limiting is enabled in `k8s/ingress.yaml`:

- **Limit**: 5 requests per second (RPS) per client IP.
- **Burst**: Allows short bursts of up to 10 requests.

### Makefile Commands

The `Makefile` provides several useful commands for managing the project:

| Command                | Description                                                              |
| ---------------------- | ------------------------------------------------------------------------ |
| `make run`             | **(Recommended)** Creates the cluster and deploys everything from scratch. |
| `make clean`           | Deletes the cluster and cleans up local Docker images.                   |
| `make build-app`       | Builds only the application's Docker image.                              |
| `make cluster-create`  | Creates the `k3d` cluster.                                               |
| `make cluster-delete`  | Deletes the `k3d` cluster.                                               |
| `make image-import`    | Imports the local Docker image into the cluster.                         |
| `make deploy`          | Applies the Kubernetes manifests from the `k8s/` directory.              |
| `make monitoring-install` | Installs Prometheus and the Prometheus Adapter.                        |

## Template for New Projects

The `template/` directory contains a reusable set of configurations to bootstrap a new project with a similar stack. To use it, copy the directory and replace the placeholder values.

```bash
# Example of replacing placeholders
cp -r template/ my-new-project/
cd my-new-project/
find . -type f -exec sed -i 's/url-shortener/my-new-app/g' {} +
find . -type f -exec sed -i 's/3000/8000/g' {} +
```
