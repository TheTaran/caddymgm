FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
RUN go build -o /out/caddymgm ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache acme.sh ca-certificates openssl curl \
  && adduser -D -H -u 10001 caddymgm
WORKDIR /app
COPY --from=build /out/caddymgm /app/caddymgm
RUN mkdir -p /config /logs /caddymgm-config /caddymgm-data && chown -R caddymgm:caddymgm /config /logs /caddymgm-config /caddymgm-data

USER caddymgm
EXPOSE 8080
ENV CADDY_CONFIG_PATH=/config/Caddyfile
ENV CADDYMGM_ACCESS_LOG_DIR=/logs
ENV CADDYMGM_WEB_LISTEN=:8080
ENV CADDYMGM_SETTINGS_PATH=/caddymgm-config/caddymgm-settings.json
ENV CADDYMGM_WEB_ACME_HOME=/caddymgm-data/acme
ENV CADDYMGM_WEB_TLS_DIR=/caddymgm-data/tls
ENTRYPOINT ["/app/caddymgm"]
