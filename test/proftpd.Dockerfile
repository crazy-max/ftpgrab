# syntax=docker/dockerfile:1

ARG ALPINE_VERSION="3.24"

FROM alpine:${ALPINE_VERSION}
RUN apk add --no-cache proftpd && \
    adduser -D -h /home/foo foo && echo 'foo:bar' | chpasswd && \
    mkdir -p /run/proftpd
COPY --from=fixtures /fixtures /fixtures
COPY <<EOF /etc/proftpd/proftpd.conf
ServerName integration
ServerType standalone
DefaultServer on
Port 21
UseIPv6 off
UseReverseDNS off
User nobody
Group nogroup
AuthOrder mod_auth_unix.c
RequireValidShell off
PassivePorts 30000 30009
WtmpLog off
TransferLog NONE
<Limit WRITE>
  DenyAll
</Limit>
EOF
CMD ["proftpd", "-n", "-c", "/etc/proftpd/proftpd.conf"]
