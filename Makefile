.PHONY: build-ci build-app cluster-create cluster-delete image-import deploy run

# Build the CI base image
build-ci:
	docker build -t minha-api-ci:latest -f ci.Dockerfile .

# Build the application image
build-app: build-ci
	docker build -t url-shortener:local .

# Create k3d cluster
cluster-create:
	k3d cluster create --config k3d-config.yaml

# Delete k3d cluster
cluster-delete:
	k3d cluster delete url-shortener-cluster

# Import image to k3d
image-import:
	k3d image import url-shortener:local -c url-shortener-cluster

# Install NGINX Ingress Controller
ingress-install:
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/cloud/deploy.yaml
	@echo "Waiting for Ingress Controller to be ready..."
	kubectl wait --namespace ingress-nginx \
	  --for=condition=ready pod \
	  --selector=app.kubernetes.io/component=controller \
	  --timeout=120s

# Setup Helm repos
helm-setup:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update

# Install Monitoring (Prometheus + Adapter)
monitoring-install: helm-setup
	helm upgrade --install prometheus prometheus-community/prometheus \
		--set alertmanager.enabled=false \
		--set pushgateway.enabled=false \
		--set server.global.scrape_interval=15s \
		--wait
	helm upgrade --install prometheus-adapter prometheus-community/prometheus-adapter \
		-f k8s/monitoring/adapter-values.yaml \
		--wait

# Deploy to k8s
deploy:
	kubectl apply -f k8s/

# Run everything (create cluster, build, import, install ingress, install monitoring, deploy)
run: cluster-create build-app image-import ingress-install monitoring-install deploy
