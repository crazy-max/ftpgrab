# syntax=docker/dockerfile:1

ARG SFTPGO_VERSION="v2.7.6"

FROM drakkan/sftpgo:${SFTPGO_VERSION}-alpine
USER root
COPY --from=fixtures /fixtures /fixtures
COPY --from=fixtures --chown=1000:1000 /ssh/ssh_host_ed25519_key /etc/sftpgo/ssh_host_ed25519_key
COPY --from=fixtures --chown=1000:1000 /ssh/ssh_host_ecdsa_key /etc/sftpgo/ssh_host_ecdsa_key
COPY <<EOF /etc/sftpgo/users.json
{
  "version": 17,
  "users": [
    {
      "username": "foo",
      "password": "bar",
      "status": 1,
      "home_dir": "/fixtures",
      "permissions": {"/": ["list", "download"]}
    }
  ]
}
EOF
ENV SFTPGO_DATA_PROVIDER__DRIVER="memory" \
    SFTPGO_DATA_PROVIDER__NAME="" \
    SFTPGO_HTTPD__BINDINGS__0__PORT=0 \
    SFTPGO_FTPD__BINDINGS__0__PORT=2121 \
    SFTPGO_SFTPD__HOST_KEYS="/etc/sftpgo/ssh_host_ed25519_key,/etc/sftpgo/ssh_host_ecdsa_key"
USER 1000:1000
CMD ["sftpgo", "serve", "--loaddata-from", "/etc/sftpgo/users.json", "--log-level", "warn"]
