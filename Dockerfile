FROM registry.access.redhat.com/ubi9/go-toolset:9.8-1780373831@sha256:bed6eea46459b3afbe9c46de513172ab08001a8a846ad29c0d7b124d6f81eeba AS builder

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go
COPY config/ config/
COPY tools/ tools/
COPY pkg/ pkg/
COPY cmd/ cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o qe-tools main.go

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.8-1793253163@sha256:660adca8516eaf4d033616118fe87c5df60af51e6346c528f3aaa225b1a5cc6d

LABEL konflux.additional-tags="latest"

USER 65532:65532

WORKDIR /qe-tools

COPY --from=builder /opt/app-root/src/qe-tools /usr/bin/
COPY --from=builder /opt/app-root/src/config config


