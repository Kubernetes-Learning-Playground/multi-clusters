# apiserver的Dockerfile文件
FROM golang:1.20.7-alpine3.17 as builder

WORKDIR /app

# copy modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# cache modules
RUN go mod download

# copy source code
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY resources/ resources/
COPY config.yaml config.yaml

# build
RUN CGO_ENABLED=0 go build \
    -a -o multi-cluster-operator cmd/main.go

FROM alpine:3.13
WORKDIR /app

USER root
COPY --from=builder --chown=root:root /app/multi-cluster-operator .
ENTRYPOINT ["./multi-cluster-operator"]