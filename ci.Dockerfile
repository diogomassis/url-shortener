ARG GO_VERSION="1.23"

FROM golang:${GO_VERSION}-alpine AS ci-base

RUN apk add --no-cache git make curl ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
