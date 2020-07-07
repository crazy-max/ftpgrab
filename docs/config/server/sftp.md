# SFTP server configuration

## Overview

You have to choose the server type `sftp` in the configuration and use the below corresponding fields.

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        host:
        port: 22
        username:
        password:
        key:
        sources:
          - /
        timeout: 30s
        max_packet_size: 32768
    ```

## Reference

### `host`

SFTP host IP or domain.

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        host: 127.0.0.1
    ```

### `port`

SFTP port. (default `22`)

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        port: 22
    ```

### `user`

SFTP username.

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        user: foo
    ```

### `password`

SFTP password or passphrase if `key` is used.

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        password: bar
    ```

### `key`

Path to your private key to enable publickey authentication.

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        key: /home/user/key.ppk
    ```

### `sources`

List of sources paths to grab from SFTP server.

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        sources:
          - /path1
          - /path2/folder
    ```

### `timeout`

Timeout is the maximum amount of time for the TCP connection to establish. `0s` means no timeout. (default `30s`)

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        timeout: 30s
    ```

### `max_packet_size`

Sets the maximum size of the payload, measured in bytes. (default `32768`)

!!! example
    ```yaml
    server:
      type: sftp
      sftp:
        max_packet_size: 32768
    ```
