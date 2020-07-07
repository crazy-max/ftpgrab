# Basic example

In this section we quickly go over a basic docker-compose file to run FTPGrab.

## Setup

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

That's it. Now you can launch FTPGrab with the following command:

```shell
$ docker-compose up -d
```
