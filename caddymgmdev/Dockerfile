FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
RUN go build -o /out/caddymgm ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
  && adduser -D -H -u 10001 caddymgm
WORKDIR /app
COPY --from=build /out/caddymgm /app/caddymgm
RUN mkdir -p /config /logs /caddymgm-config && chown -R caddymgm:caddymgm /config /logs /caddymgm-config

USER caddymgm
EXPOSE 8080
ENV CADDY_CONFIG_PATH=/config/Caddyfile
ENV CADDYMGM_ACCESS_LOG_DIR=/logs
ENV CADDYMGM_WEB_LISTEN=:8080
ENV CADDYMGM_SETTINGS_PATH=/caddymgm-config/caddymgm-settings.json
ENTRYPOINT ["/app/caddymgm"]
