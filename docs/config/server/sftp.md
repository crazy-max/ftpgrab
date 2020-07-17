# SFTP server configuration

!!! warning
    `ftp` and `sftp` are mutually exclusive

!!! example
    ```yaml
    server:
      sftp:
        host: 10.0.0.1
        port: 22
        username: foo
        password: bar
        key: /home/user/key.ppk
        sources:
          - /
        timeout: 30s
        maxPacketSize: 32768
    ```

## Reference

### `host`

SFTP host IP or domain.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        host: 127.0.0.1
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_HOST`

### `port`

SFTP port. (default `22`)

!!! example "Config file"
    ```yaml
    server:
      sftp:
        port: 22
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_PORT`

### `username`

SFTP username.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        username: foo
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_USERNAME`

### `password`

SFTP password or passphrase if `key` is used.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        password: bar
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_PASSWORD`

### `key`

Path to your private key to enable publickey authentication.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        key: /home/user/key.ppk
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_KEY`

### `sources`

List of sources paths to grab from SFTP server.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        sources:
          - /path1
          - /path2/folder
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_SOURCES`

### `timeout`

Timeout is the maximum amount of time for the TCP connection to establish. `0s` means no timeout. (default `30s`)

!!! example "Config file"
    ```yaml
    server:
      sftp:
        timeout: 30s
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_TIMEOUT`

### `maxPacketSize`

Sets the maximum size of the payload, measured in bytes. (default `32768`)

!!! example "Config file"
    ```yaml
    server:
      sftp:
        maxPacketSize: 32768
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_MAXPACKETSIZE`
