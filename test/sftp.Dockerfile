# syntax=docker/dockerfile:1

ARG ALPINE_VERSION="3.24"

FROM alpine:${ALPINE_VERSION}
RUN apk add --no-cache openssh && \
    adduser -D -h /home/foo foo && echo 'foo:bar' | chpasswd && \
    ssh-keygen -A
COPY --from=fixtures /ssh/ssh_host_ed25519_key /etc/ssh/ssh_host_ed25519_key
COPY --from=fixtures /fixtures /fixtures
COPY --from=fixtures /symlink-fixtures /symlink-fixtures
COPY <<EOF /etc/ssh/sshd_config
PasswordAuthentication yes
HostKey /etc/ssh/ssh_host_ed25519_key
HostKey /etc/ssh/ssh_host_rsa_key
HostKey /etc/ssh/ssh_host_ecdsa_key
PermitRootLogin no
Subsystem sftp internal-sftp
EOF
CMD ["/usr/sbin/sshd", "-D", "-e"]
