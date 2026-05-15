# syntax=docker/dockerfile:1

FROM golang:1.22-bookworm AS build
WORKDIR /src
ENV GOWORK=off

COPY services/panel-api/go.mod services/panel-api/go.sum ./services/panel-api/

WORKDIR /src/services/panel-api
RUN go mod download

WORKDIR /src
COPY services/panel-api ./services/panel-api

WORKDIR /src/services/panel-api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/panel-api ./cmd/panel-api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bootstrap-admin ./cmd/bootstrap-admin

FROM debian:bookworm-slim AS runtime
RUN useradd -r -u 10001 -g nogroup lenker
COPY --from=build /out/panel-api /usr/local/bin/panel-api
COPY --from=build /out/bootstrap-admin /usr/local/bin/bootstrap-admin
USER 10001:65534
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/panel-api"]
