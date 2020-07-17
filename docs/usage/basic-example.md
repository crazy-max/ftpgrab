# Basic example

In this section we quickly go over a basic way to run FTPGrab.

## Setup

!!! warning
    Make sure to follow the instructions to [install from binary](../install/binary.md) before.

First create a [`ftpgrab.yml` configuration](../config/index.md) file like this one:

```yaml
# ./ftpgrab.yml

db:
  path: ftpgrab.db

server:
  ftp:
    host: test.rebex.net
    port: 21
    username: demo
    password: password
    sources:
      - /
    timeout: 5s

download:
  output: /download
  retry: 3
  hideSkipped: false
  createBaseDir: false

notif:
  mail:
    host: smtp.example.com
    port: 25
    username: foo
    password: bar
    from: ftpgrab@example.com
    to: webmaster@example.com
```

That's it. Now you can launch FTPGrab with the following command:

```shell
$ ftpgrab --config ./ftpgrab.yml
```
