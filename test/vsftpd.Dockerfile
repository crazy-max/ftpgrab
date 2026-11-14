# syntax=docker/dockerfile:1

ARG ALPINE_VERSION="3.24"

FROM alpine:${ALPINE_VERSION}
RUN apk add --no-cache vsftpd && \
    adduser -D -h /home/foo foo && echo 'foo:bar' | chpasswd
COPY --from=fixtures /fixtures /fixtures
COPY --from=fixtures /symlink-fixtures /symlink-fixtures
COPY <<EOF /etc/vsftpd/vsftpd.conf
listen=YES
listen_ipv6=NO
background=NO
anonymous_enable=NO
local_enable=YES
write_enable=NO
pasv_min_port=30000
pasv_max_port=30009
seccomp_sandbox=NO
EOF
CMD ["vsftpd", "/etc/vsftpd/vsftpd.conf"]
