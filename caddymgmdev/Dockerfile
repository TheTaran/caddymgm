FROM ghcr.io/maxmind/geoipupdate:v8.0.0 AS geoipupdate

FROM golang:1.27.1-alpine AS build

WORKDIR /src
ARG VERSION=dev
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
RUN go build -ldflags "-X main.version=${VERSION}" -o /out/caddymgm ./cmd/server

FROM alpine:3.24

RUN apk upgrade --no-cache \
  && apk add --no-cache ca-certificates \
  && adduser -D -H -u 10001 caddymgm
WORKDIR /app
COPY --from=build /out/caddymgm /app/caddymgm
COPY --from=geoipupdate /usr/bin/geoipupdate /usr/local/bin/geoipupdate
RUN mkdir -p /config /logs /caddymgm && chown -R caddymgm:caddymgm /config /logs /caddymgm

USER caddymgm
EXPOSE 8080
ENV CADDY_CONFIG_PATH=/config/Caddyfile
ENV CADDYMGM_ACCESS_LOG_DIR=/logs
ENV CADDYMGM_WEB_LISTEN=:8080
ENV CADDYMGM_SETTINGS_PATH=/caddymgm/caddymgm-settings.json
ENTRYPOINT ["/app/caddymgm"]
