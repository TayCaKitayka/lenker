# syntax=docker/dockerfile:1

FROM golang:1.22-bookworm AS build
WORKDIR /src
ENV GOWORK=off

COPY services/node-agent/go.mod ./services/node-agent/

WORKDIR /src/services/node-agent
RUN go mod download

WORKDIR /src
COPY services/node-agent ./services/node-agent

WORKDIR /src/services/node-agent
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/node-agent ./cmd/node-agent

FROM debian:bookworm-slim AS runtime
RUN useradd -r -u 10001 -g nogroup lenker && mkdir -p /var/lib/lenker/node-agent && chown -R lenker:nogroup /var/lib/lenker
COPY --from=build /out/node-agent /usr/local/bin/node-agent
USER 10001:65534
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/node-agent"]
