# syntax=docker/dockerfile:1

# renovate: datasource=docker depName=alpine versioning=docker
ARG ALPINE_VERSION="3.24"
# renovate: datasource=docker depName=crazymax/pure-ftpd versioning=docker
ARG CRAZYMAX_PUREFTPD_VERSION="1.0.54"

FROM alpine:${ALPINE_VERSION} AS pureftpd
RUN apk add --no-cache pure-ftpd && \
    adduser -D -h /home/foo foo && echo 'foo:bar' | chpasswd
COPY --from=fixtures /fixtures /fixtures
CMD ["pure-ftpd", "-l", "unix", "-E", "-H", "-4", "-p", "30000:30009"]

FROM crazymax/pure-ftpd:${CRAZYMAX_PUREFTPD_VERSION} AS crazymax-pureftpd
COPY --from=fixtures /fixtures /fixtures
RUN printf 'bar\nbar\n' | pure-pw useradd foo -u 1000 -g 1000 -d /fixtures -f /data/pureftpd.passwd
ENV AUTH_METHOD="puredb" \
    PASSIVE_IP="crazymax-pureftpd"
