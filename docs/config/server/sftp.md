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

### `usernameFile`

Use content of secret file as SFTP username if `username` not defined.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        usernameFile: /run/secrets/username
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_USERNAMEFILE`

### `password`

!!! warning
    `password` and `keyFile` are mutually exclusive

SFTP password.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        password: bar
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_PASSWORD`

### `passwordFile`

Use content of secret file as SFTP password if `password` not defined.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        passwordFile: /run/secrets/password
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_PASSWORDFILE`

### `keyFile`

!!! warning
    `keyFile` and `password` are mutually exclusive

Path to your private key to enable publickey authentication.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        keyFile: /home/user/key.ppk
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_KEYFILE`

### `keyPassphrase`

SFTP key passphrase if `keyFile` is defined.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        keyPassphrase: bar
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_SERVER_SFTP_PASSWORD`

### `keyPassphraseFile`

Use content of secret file as SFTP key passphrase if `keyPassphrase` not defined.

!!! example "Config file"
    ```yaml
    server:
      sftp:
        keyPassphraseFile: /run/secrets/passphrase
    ```

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
