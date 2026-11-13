# syntax=docker/dockerfile:1

ARG ALPINE_VERSION="3.24"

FROM alpine:${ALPINE_VERSION}
RUN apk add --no-cache openssh && \
    adduser -D -h /home/foo foo && echo 'foo:bar' | chpasswd && \
    ssh-keygen -A
COPY --from=fixtures /fixtures /fixtures
COPY <<EOF /etc/ssh/sshd_config
PasswordAuthentication yes
PermitRootLogin no
Subsystem sftp internal-sftp
EOF
CMD ["/usr/sbin/sshd", "-D", "-e"]
