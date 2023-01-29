FROM golang:1.19.5-alpine3.17 as builder

RUN apk --update add ca-certificates
RUN echo 'nve:*:65532:' > /tmp/group && \
    echo 'nve:*:65532:65532:nve:/:/nve-hydapi-exporter' > /tmp/passwd

WORKDIR /workspace
COPY go.* ./
RUN go mod download

COPY . /workspace

RUN CGO_ENABLED=0 go build -a -o nve-hydapi-exporter .

FROM scratch

LABEL org.opencontainers.image.title="nve-hydapi-exporter" \
      org.opencontainers.image.description="Prometheus exporter for NVE hydrological data" \
      org.opencontainers.image.authors="Terje Sannum <terje@offpiste.org>" \
      org.opencontainers.image.url="https://github.com/terjesannum/nve-hydapi-exporter" \
      org.opencontainers.image.source="https://github.com/terjesannum/nve-hydapi-exporter"

WORKDIR /
EXPOSE 8080

COPY --from=builder /tmp/passwd /tmp/group /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /workspace/nve-hydapi-exporter .

USER 65532:65532

ENTRYPOINT ["/nve-hydapi-exporter"]
