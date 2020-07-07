# FTP server configuration

## Overview

You have to choose the server type `ftp` in the configuration and use the below corresponding fields.

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        host:
        port: 21
        username:
        password:
        sources:
          - /
        timeout: 5s
        disable_epsv: false
        tls: false
        insecure_skip_verify: false
        log_trace: false
    ```

## Reference

### `host`

FTP host IP or domain.

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        host: 127.0.0.1
    ```

### `port`

FTP port. (default `21`)

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        port: 21
    ```

### `username`

FTP username.

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        username: foo
    ```

### `password`

FTP password.

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        password: bar
    ```

### `sources`

List of sources paths to grab from FTP server.

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        sources:
          - /path1
          - /path2/folder
    ```

### `timeout`

Timeout for opening connections, sending control commands, and each read/write of data transfers. (default `5s`)

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        timeout: 5s
    ```

### `disable_epsv`

Disables EPSV in favour of PASV. This is useful in cases where EPSV connections neither complete nor downgrade to
PASV successfully by themselves, resulting in hung connections. (default `false`)

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        disable_epsv: false
    ```

### `tls`

Use implicit FTP over TLS. (default `false`)

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        tls: false
    ```

### `insecure_skip_verify`

Controls whether a client verifies the server’s certificate chain and host name. (default `false`)

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        insecure_skip_verify: false
    ```

### `log_trace`

Enable low-level FTP log. Works only if global log level is debug. (default `false`)

!!! example
    ```yaml
    server:
      type: ftp
      ftp:
        log_trace: false
    ```
