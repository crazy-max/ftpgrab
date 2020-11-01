# Installation with Docker

## About

FTPGrab provides automatically updated Docker :whale: images within several registries:

| Registry                                                                                         | Image                           |
|--------------------------------------------------------------------------------------------------|---------------------------------|
| [Docker Hub](https://hub.docker.com/r/crazymax/ftpgrab/)                             | `crazymax/ftpgrab`                 |
| [GitHub Container Registry](https://github.com/users/crazy-max/packages/container/package/ftpgrab)  | `ghcr.io/crazy-max/ftpgrab`        |

It is possible to always use the latest stable tag or to use another service that handles updating Docker images.

!!! note
    Want to be notified of new releases? Check out :bell: [Diun (Docker Image Update Notifier)](https://github.com/crazy-max/diun) project!

Following platforms for this image are available:

```shell
$ docker run --rm mplatform/mquery crazymax/ftpgrab:latest
Image: crazymax/ftpgrab:latest
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

```yaml
version: "3.5"

services:
  ftpgrab:
    image: crazymax/ftpgrab:latest
    container_name: ftpgrab
    volumes:
      - "./db:/db:rw"
      - "./download:/download:rw"
    environment:
      - "TZ=Europe/Paris"
      - "SCHEDULE=*/30 * * * *"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
      - "FTPGRAB_SERVER_FTP_HOST=test.rebex.net"
      - "FTPGRAB_SERVER_FTP_PORT=21"
      - "FTPGRAB_SERVER_FTP_USERNAME=demo"
      - "FTPGRAB_SERVER_FTP_PASSWORD=password"
      - "FTPGRAB_SERVER_FTP_SOURCES=/src1,/src2"
      - "FTPGRAB_DOWNLOAD_GID=1000"
      - "FTPGRAB_DOWNLOAD_UID=1000"
      - "FTPGRAB_DOWNLOAD_INCLUDE=^Mr\\.Robot\\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+"
      - "FTPGRAB_DOWNLOAD_EXCLUDE=\\.nfo$$"
      - "FTPGRAB_DOWNLOAD_SINCE=2019-02-01T18:50:05Z"
      - "FTPGRAB_DOWNLOAD_RETRY=5"
      - "FTPGRAB_NOTIF_MAIL_HOST=smtp.example.com"
      - "FTPGRAB_NOTIF_MAIL_PORT=25"
      - "FTPGRAB_NOTIF_MAIL_FROM=ftpgrab@example.com"
      - "FTPGRAB_NOTIF_MAIL_TO=webmaster@example.com"
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
    -e "FTPGRAB_SERVER_FTP_HOST=test.rebex.net" \
    -e "FTPGRAB_SERVER_FTP_PORT=21" \
    -e "FTPGRAB_SERVER_FTP_USERNAME=demo" \
    -e "FTPGRAB_SERVER_FTP_PASSWORD=password" \
    -e "FTPGRAB_SERVER_FTP_SOURCES=/src1,/src2" \
    -e "FTPGRAB_DOWNLOAD_GID=1000" \
    -e "FTPGRAB_DOWNLOAD_UID=1000" \
    -e "FTPGRAB_DOWNLOAD_INCLUDE=^Mr\.Robot\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+" \
    -e "FTPGRAB_DOWNLOAD_EXCLUDE=\.nfo\$" \
    -e "FTPGRAB_DOWNLOAD_SINCE=2019-02-01T18:50:05Z" \
    -e "FTPGRAB_DOWNLOAD_RETRY=5" \
    -e "FTPGRAB_NOTIF_MAIL_HOST=smtp.example.com" \
    -e "FTPGRAB_NOTIF_MAIL_PORT=25" \
    -e "FTPGRAB_NOTIF_MAIL_FROM=ftpgrab@example.com" \
    -e "FTPGRAB_NOTIF_MAIL_TO=webmaster@example.com" \
    -v "$(pwd)/db:/db:rw" \
    -v "$(pwd)/download:/download:rw" \
    crazymax/ftpgrab:latest
```

To upgrade your installation to the latest release:

```shell
$ docker-compose pull
$ docker-compose up -d
```

If you prefer to rely on the [`configuration file](../config/index.md#configuration-file) instead of
environment variables:

```yaml
# ./ftpgrab.yml

server:
  ftp:
    host: test.rebex.net
    port: 21
    username: demo
    password: password
    sources:
      - /src1
      - /src2

download:
  uid: 1000
  gid: 1000
  include:
    - ^Mr\.Robot\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+
  exclude:
    - \.nfo$
  since: 2019-02-01T18:50:05Z
  retry: 5

notif:
  mail:
    host: smtp.example.com
    port: 25
    from: ftpgrab@example.com
    to: webmaster@example.com
```

And here is your docker composition:

```yaml
version: "3.5"

services:
  ftpgrab:
    image: crazymax/ftpgrab:latest
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
