ARG BASE_IMAGE="minha-api-ci"
ARG CI_IMAGE_TAG="latest"
FROM ${BASE_IMAGE}:${CI_IMAGE_TAG} AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/api-server \
    ./cmd/http.go

FROM scratch AS final

LABEL org.opencontainers.image.title="Go Fiber API" \
    org.opencontainers.image.source="https://github.com/diogomassis/url-shortener"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/api-server /api-server

USER 65534:65534

EXPOSE 3000

ENTRYPOINT ["/api-server"]
