# URL Shortener Project

A high-performance URL Shortener service built with Go (Fiber), designed to run on Kubernetes with a complete observability and autoscaling stack.

## Overview

This project demonstrates a production-ready setup for a Go application on Kubernetes. It includes a local development environment using **k3d**, **NGINX Ingress Controller** for traffic management, **Prometheus** for monitoring, and **Horizontal Pod Autoscaler (HPA)** based on custom metrics (QPS).

## Architecture

* **Application**: Go (Fiber) REST API.
* **Containerization**: Docker (Multi-stage builds).
* **Orchestration**: Kubernetes (k3d for local dev).
* **Ingress**: NGINX Ingress Controller with Rate Limiting.
* **Observability**: Prometheus & Prometheus Adapter.
* **Autoscaling**: HPA based on Requests Per Second (QPS).

## Project Structure

```project
.
├── cmd/                    # Application source code
│   └── http.go             # Main entry point
├── k8s/                    # Kubernetes manifests
│   ├── deployment.yaml     # App deployment configuration
│   ├── service.yaml        # ClusterIP service definition
│   ├── ingress.yaml        # Ingress rules & rate limiting
│   ├── hpa.yaml            # Horizontal Pod Autoscaler config
│   └── monitoring/         # Monitoring stack configuration
│       └── adapter-values.yaml # Prometheus Adapter rules
├── template/               # Reusable K8s configuration template
├── ci.Dockerfile           # Base image for CI/Build
├── Dockerfile              # Multi-stage application Dockerfile
├── k3d-config.yaml         # k3d cluster configuration
└── Makefile                # Automation scripts
```

## Prerequisites

Ensure you have the following tools installed:

* [Docker](https://www.docker.com/)
* [k3d](https://k3d.io/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Helm](https://helm.sh/)
* [Make](https://www.gnu.org/software/make/)

## Getting Started

The entire environment can be spun up with a single command:

```bash
make run
```

This command will:

1. Create a local Kubernetes cluster using `k3d`.
2. Build the application Docker image.
3. Import the image into the cluster.
4. Install **NGINX Ingress Controller**.
5. Install **Prometheus** and **Prometheus Adapter**.
6. Deploy the application and HPA configurations.

### Accessing the Application

Once the deployment is complete, the application is accessible at:

**<http://localhost:8081/>**

### Useful Commands

* **Stop/Delete Cluster**:

    ```bash
    make cluster-delete
    ```

* **Re-deploy Application**:

    ```bash
    make build-app image-import deploy
    ```

## Key Features

### 1. Ingress & Rate Limiting

Traffic is managed by NGINX Ingress Controller. Rate limiting is configured in `k8s/ingress.yaml` to protect the service:

* **Limit**: 5 requests per second (RPS) per IP.
* **Burst**: Multiplier of 2 (allows bursts up to 10 requests).

### 2. Autoscaling (HPA)

The application scales automatically based on traffic load (QPS).

* **Metric**: `http_requests_qps` (derived from Prometheus).
* **Target**: 10 QPS per pod.
* **Scale Range**: 1 to 10 replicas.

### 3. Monitoring

* **Prometheus**: Scrapes metrics from the application at `/metrics`.
* **Prometheus Adapter**: Converts Prometheus metrics into Kubernetes Custom Metrics API for HPA.

## Template Usage

A reusable template is provided in the `template/` directory for bootstrapping new projects with the same infrastructure stack.

To use it:

1. Copy the `template` folder to your new project.
2. Replace placeholders:
    * `{{PROJECT_NAME}}`: Your project name.
    * `{{APP_PORT}}`: Your application port.

```bash
# Example replacement
find template -type f -exec sed -i 's/{{PROJECT_NAME}}/my-new-app/g' {} +
```
