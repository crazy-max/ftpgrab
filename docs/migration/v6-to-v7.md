# FTPGrab v6 to v7

## Configuration transposed into environment variables

All configuration is now transposed into environment variables. Take a look at the
[documentation](../config/index.md#environment-variables) for more details.

`FTPGRAB_DB` env var has been renamed `FTPGRAB_DB_PATH` to follow environment variables transposition.

## All fields in configuration are now _camelCased_

In order to enable transposition into environmental variables, all fields in configuration are now _camelCased_:

* `server.ftp.disable_epsv` > `server.ftp.disableEPSV`
* `download.chmod_file` > `download.chmodFile`
* `notif.mail.insecure_skip_verify` > `notif.mail.insecureSkipVerify`
* ...

??? example "v6"
    ```yaml
    db:
      enable: true
      path: ftpgrab.db
    
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
        insecure_skip_verify: false
        log_trace: false
    
    download:
      uid: 1000
      gid: 1000
      chmod_file: 0644
      chmod_dir: 0755
      include:
        - ^Mr\.Robot\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+
      exclude:
        - \.nfo$
      since: 2019-02-01T18:50:05Z
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
      webhook:
        enable: true
        endpoint: http://webhook.foo.com/sd54qad89azd5a
        method: GET
        headers:
          Content-Type: application/json
          Authorization: Token123456
        timeout: 10s
    ```

??? example "v7"
    ```yaml
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
        disableEPSV: false
        tls: false
        insecureSkipVerify: false
        logTrace: false
    
    download:
      output: /download
      uid: 1000
      gid: 1000
      chmodFile: 0o644
      chmodDir: 0o755
      include:
        - ^Mr\.Robot\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+
      exclude:
        - \.nfo$
      since: 2019-02-01T18:50:05Z
      retry: 3
      hideSkipped: false
      createBaseDir: false
    
    notif:
      mail:
        host: smtp.example.com
        port: 587
        ssl: false
        insecureSkipVerify: false
        from: ftpgrab@example.com
        to: webmaster@example.com
      webhook:
        endpoint: http://webhook.foo.com/sd54qad89azd5a
        method: GET
        headers:
          content-type: application/json
          authorization: Token123456
        timeout: 10s
    ```

## Remove `type` field for server

The `type` field has been removed for server. The server type will now be choosed if it's defined.

!!! warning
    `ftp` and `sftp` are mutually exclusive

!!! example "v6"
    ```yaml
    server:
      type: ftp
      ftp:
        host: test.rebex.net
        port: 21
        sources:
          - /
    ```

!!! example "v7"
    ```yaml
    server:
      ftp:
        host: test.rebex.net
        port: 21
        sources:
          - /
    ```

## Changes for SFTP auth fields

`key` field has been renamed `keyFile` and can be use with the dedicated `keyPassphrase` field if a passphrase is required.

## Remove `enable` field for notifiers

The `enable` field has been removed for notifiers. If you don't want a notifier to be enabled, you must now remove
or comment its configuration.

!!! example "v6"
    ```yaml
    notif:
      mail:
        enable: true
        host: smtp.example.com
        port: 587
        ssl: false
        insecureSkipVerify: false
        from: ftpgrab@example.com
        to: webmaster@example.com
      webhook:
        enable: false
        endpoint: http://webhook.foo.com/sd54qad89azd5a
        method: GET
        timeout: 10s
    ```

!!! example "v7"
    ```yaml
    notif:
      mail:
        host: smtp.example.com
        port: 587
        ssl: false
        insecureSkipVerify: false
        from: ftpgrab@example.com
        to: webmaster@example.com
    ```
