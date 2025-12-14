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

## Architectural Decisions

### 1. Hexagonal Architecture (Ports and Adapters)

**What:**
The application is structured using Hexagonal Architecture, dividing the code into "Core" (Business Logic) and "Adapters" (Infrastructure).

**How:**

- **Core (`internal/core`)**: Contains the `Domain` entities (e.g., `URL`), `Ports` (interfaces like `URLRepository`, `URLService`), and `Services` (business logic implementation). This layer has no external dependencies.
- **Adapters (`internal/adapters`)**: Contains the implementation of the ports.
  - **Driving Adapters**: The HTTP Handler (`internal/adapters/handler/http`) drives the application.
  - **Driven Adapters**: The Repositories (`internal/adapters/repository`) implement the storage interface.

**Why:**

- **Decoupling**: The business logic is independent of the database or web framework.
- **Testability**: We can easily mock interfaces to test the core logic without spinning up a database or server.
- **Flexibility**: Switching from In-Memory to Redis or SQL requires only writing a new adapter, without changing the core logic.

### 2. Caching Strategy (Redis + Fallback)

**What:**
A hybrid caching mechanism using Redis as the primary cache and an In-Memory map as a fallback/persistent storage (for this demo).

**How:**
We implemented a **Decorator Pattern** in `internal/adapters/repository/cached`.

- **Write (Save)**:
    1. Saves to the Persistent storage (Source of Truth) first to ensure data safety.
    2. Asynchronously updates the Redis cache (Best Effort).
- **Read (Get)**:
    1. Tries to fetch from Redis first (Fast path).
    2. **Fallback**: If Redis is down or the key is missing, it fetches from Persistent storage.
    3. **Self-Healing**: If found in Persistent storage but not in Redis, it asynchronously populates Redis.

**Why:**

- **Performance**: Redis provides sub-millisecond access times for high-read workloads (redirects).
- **Resilience**: If the Redis cluster fails, the application automatically degrades to using the internal memory/database without downtime.

### 3. Collision-Resistant Hash Generation

**What:**
A deterministic and collision-resistant approach to generating short codes.

**How:**
The `generateUniqueShortCode` function in the URL service implements a multi-step process:

1. **Deterministic Hashing**: It creates a `SHA256` hash from the original URL combined with a salt (an integer that is incremented on retries). This ensures that the same URL will always produce the same sequence of potential hashes.
2. **Shortening with `hashids`**: It uses the `hashids` library to convert the first few bytes of the hash into a short, non-sequential, and URL-friendly string.
3. **Collision Handling**:
    - Before saving, it checks if the generated `shortCode` already exists in the repository.
    - If a collision is detected, it increments the salt and retries the process, generating a new, different hash.
    - This retry mechanism is attempted up to a `MaxRetries` limit to prevent infinite loops.

**Why:**

- **Predictability**: The same URL will always map to the same short code, preventing the creation of multiple short URLs for the same destination.
- **Collision Resistance**: The retry mechanism with a changing salt makes it highly unlikely that two different URLs will produce the same short code.
- **Security**: Using `hashids` with a secret salt makes it difficult for others to guess the sequence of short codes.

## Known Issues, Trade-offs, and Future Improvements

### Problems and Restrictions

- **In-Memory Fallback is Not Persistent**: The current fallback repository (`memoryRepository`) is not suitable for production. If all pods restart, all URL data will be lost unless it was successfully cached in Redis. This should be replaced with a persistent database like PostgreSQL or a distributed key-value store.
- **Single Point of Failure (Redis)**: Although the application has a fallback, if Redis is the primary data source and it fails, new URLs can't be reliably shortened across multiple replicas, as each pod would have its own in-memory state.
- **Stateless `hashids` Salt**: The salt for `hashids` is hardcoded. A more secure approach would be to manage this as a secret within Kubernetes.

### Trade-offs

- **Performance vs. Consistency (Cache)**: The cache is updated asynchronously ("best effort") to avoid blocking API responses. This means there's a brief moment where the cache might be stale after a write operation. This trade-off prioritizes write performance over immediate consistency.
- **Simplicity vs. Robustness (Repository)**: The repository interfaces are simple (`Get`, `Save`). They don't account for more complex scenarios like transactions or batch operations, which would be necessary for a more advanced system.

### Rejected Decisions

- **Using a Database as the First Choice**: We initially considered using a full-fledged SQL database like PostgreSQL. This was rejected for the initial version to favor simplicity and speed, using Redis and in-memory storage to get a working prototype faster.
- **Fully Random Short Codes**: Generating a completely random string for the short code was rejected because it doesn't guarantee that the same original URL will map to the same short URL. The deterministic hashing approach was chosen to provide this consistency.

### Known Failures

- **Hash Collision Exhaustion**: In the extremely rare event that a URL generates 10 short codes that all happen to be in use (collisions), the `Shorten` function will fail and return an error. The probability is astronomically low but not zero.
- **Data Race in In-Memory Repository**: The `memoryRepository` uses a `sync.RWMutex` for concurrency control. While generally safe, it's not as battle-tested as a proper database's transaction and isolation levels. Heavy concurrent writes could potentially reveal edge cases.

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
