ARG CADDY_VERSION=2.11.4

FROM caddy:${CADDY_VERSION}-builder AS builder

# The official builder image includes Go outside its default PATH.
ENV PATH="/usr/local/go/bin:${PATH}"
ARG CADDY_GEOIP_MODULE_VERSION=v0.6.0

RUN set -eux; \
    for attempt in 1 2 3; do \
      xcaddy build --output /usr/bin/caddy \
        --with github.com/shift72/caddy-geo-ip@${CADDY_GEOIP_MODULE_VERSION} && exit 0; \
      echo "Caddy Geo-IP build attempt ${attempt} failed; retrying..." >&2; \
      sleep $((attempt * 5)); \
    done; \
    exit 1

FROM caddy:${CADDY_VERSION}-alpine

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
