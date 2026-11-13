# syntax=docker/dockerfile:1

ARG ALPINE_VERSION="3.24"

FROM alpine:${ALPINE_VERSION} AS fixtures
RUN apk add --no-cache openssh-keygen && \
    mkdir /ssh && \
    ssh-keygen -q -t ed25519 -N '' -f /ssh/ssh_host_ed25519_key && \
    ssh-keygen -q -t ecdsa -N '' -f /ssh/ssh_host_ecdsa_key && \
    awk '{print "sftp,[sftpgo]:2022 " $1 " " $2}' \
      /ssh/ssh_host_ed25519_key.pub > /ssh/known_hosts
RUN <<EOT
#!/bin/sh
set -eu

mkdir -p /fixtures/basic/baz /fixtures/empty-dir /fixtures/many \
  '/fixtures/path with spaces' '/fixtures/regexp [dir](v1)+.^$' \
  "/fixtures/quotes 'single'" '/fixtures/quotes "double"' \
  '/fixtures/日本語' /fixtures/nested/one/two/three

printf 'bar\n' > /fixtures/basic/bar.txt
printf 'baz\n' > /fixtures/basic/baz/baz.txt
printf '' > /fixtures/empty.bin
printf 'spaces\n' > '/fixtures/path with spaces/file name.txt'
printf 'regexp\n' > '/fixtures/regexp [dir](v1)+.^$/file [1](a)+.^$.txt'
printf 'single quotes\n' > "/fixtures/quotes 'single'/file 'quoted'.txt"
printf 'double quotes\n' > '/fixtures/quotes "double"/file "quoted".txt'
printf 'unicode\n' > '/fixtures/日本語/café-世界.txt'
printf '\000\001\002\377' > /fixtures/nested/one/two/three/data.bin

# Sparse files keep fixture generation cheap. Transfers still send every byte.
# A nonzero trailer catches truncated or zero-filled large downloads.
for size in 16 32 64 128 256 512; do
  file="/fixtures/large-${size}MiB.bin"
  truncate -s "$((size * 1024 * 1024 - 8))" "$file"
  printf '\001\002\377tail\n' >> "$file"
done

i=0
while [ "$i" -lt 5000 ]; do
  name=$(printf 'file-%04d.txt' "$i")
  printf 'fixture %s\n' "$i" > "/fixtures/many/$name"
  i=$((i + 1))
done

# Older FTP LIST entries only expose the date, so use midnight UTC.
find /fixtures -type f -exec touch -t 202401020000.00 {} +
EOT
