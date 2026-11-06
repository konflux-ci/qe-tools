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

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.8-1793862054@sha256:ee7fd2977589d1f776d703e99a03b2ee1c332868bb2b2e2f921c3c1abc55ce9b

LABEL konflux.additional-tags="latest"

USER 65532:65532

WORKDIR /qe-tools

COPY --from=builder /opt/app-root/src/qe-tools /usr/bin/
COPY --from=builder /opt/app-root/src/config config


