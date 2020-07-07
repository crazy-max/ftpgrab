# Installation with Docker

## About

FTPGrab provides automatically updated Docker :whale: images within [Docker Hub](https://hub.docker.com/r/ftpgrab/ftpgrab).
It is possible to always use the latest stable tag or to use another service that handles updating Docker images.

!!! note
    Want to be notified of new releases? Check out :bell: [Diun (Docker Image Update Notifier)](https://github.com/crazy-max/diun) project!

Following platforms for this image are available:

```shell
$ docker run --rm mplatform/mquery ftpgrab/ftpgrab:latest
Image: ftpgrab/ftpgrab:latest
 * Manifest List: Yes
 * Supported platforms:
   - linux/amd64
   - linux/arm/v6
   - linux/arm/v7
   - linux/arm64
   - linux/386
   - linux/ppc64le
   - linux/s390x
```

This reference setup guides users through the setup based on `docker-compose`, but the installation of `docker-compose`
is out of scope of this documentation. To install `docker-compose` itself, follow the official
[install instructions](https://docs.docker.com/compose/install/).

## Volumes

| Path               | Description   |
|--------------------|---------------|
| `/db`              | Folder containing bbolt database file |
| `/download`        | Downloaded files folder |

## Usage

First create a [`ftpgrab.yml` configuration](../config/index.md) file like this one:

```yaml
server:
  type: ftp
  ftp:
    host: test.rebex.net
    port: 21
    username: demo
    password: password
    sources:
      - /
    timeout: 5s
    disable_epsv: false
    tls: false

db:
  enable: true

download:
  retry: 3
  hide_skipped: false
  create_basedir: false

notif:
  mail:
    enable: true
    host: smtp.example.com
    port: 587
    ssl: false
    insecure_skip_verify: false
    username: webmaster@example.com
    password: apassword
    from: ftpgrab@example.com
    to: webmaster@example.com
```

And your docker composition:

```yaml
version: "3.5"

services:
  ftpgrab:
    image: ftpgrab/ftpgrab:latest
    container_name: ftpgrab
    volumes:
      - "./db:/db:rw"
      - "./download:/download:rw"
      - "./ftpgrab.yml:/ftpgrab.yml:ro"
    environment:
      - "TZ=Europe/Paris"
      - "SCHEDULE=*/30 * * * *"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
    restart: always
```

Edit this example with your preferences and run the following commands to bring up FTPGrab:

```shell
$ docker-compose up -d
$ docker-compose logs -f
```

Or use the following command:

```shell
$ docker run -d --name ftpgrab \
    -e "TZ=Europe/Paris" \
    -e "SCHEDULE=*/30 * * * *" \
    -e "LOG_LEVEL=info" \
    -e "LOG_JSON=false" \
    -v "$(pwd)/db:/db:rw" \
    -v "$(pwd)/download:/download:rw" \
    -v "$(pwd)/ftpgrab.yml:/ftpgrab.yml:ro" \
    ftpgrab/ftpgrab:latest
```

To upgrade your installation to the latest release:

```shell
$ docker-compose pull
$ docker-compose up -d
```
