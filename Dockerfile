FROM golang:1.18.5-alpine as builder

WORKDIR /workspace
COPY go.* ./
RUN go mod download

COPY . /workspace

RUN CGO_ENABLED=0 go build -a -o nve-hydapi-exporter .

FROM alpine:3.15.5
WORKDIR /

COPY --from=builder /workspace/nve-hydapi-exporter .
USER 65532:65532

ENTRYPOINT ["/nve-hydapi-exporter"]
