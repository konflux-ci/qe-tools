FROM registry.access.redhat.com/ubi9/go-toolset:9.8-1793294290@sha256:56881c8ed2aa8c3b8493a89f62c8dc9f575747708380d2bae861224ccefd7d5c AS builder

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go
COPY config/ config/
COPY tools/ tools/
COPY pkg/ pkg/
COPY cmd/ cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o qe-tools main.go

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.8-1780378819@sha256:2a7e516b217a8e9d18021f40a061f22274d6d677d11e3e368a672827e149050d

LABEL konflux.additional-tags="latest"

USER 65532:65532

WORKDIR /qe-tools

COPY --from=builder /opt/app-root/src/qe-tools /usr/bin/
COPY --from=builder /opt/app-root/src/config config


