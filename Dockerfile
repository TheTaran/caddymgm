FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
RUN go build -o /out/caddymgm ./cmd/server

FROM alpine:3.20

RUN adduser -D -H -u 10001 caddymgm
WORKDIR /app
COPY --from=build /out/caddymgm /app/caddymgm
RUN mkdir -p /config /logs && chown -R caddymgm:caddymgm /config /logs

USER caddymgm
EXPOSE 8080
ENV CADDY_CONFIG_PATH=/config/Caddyfile
ENV CADDYMGM_ACCESS_LOG_DIR=/logs
ENTRYPOINT ["/app/caddymgm"]
