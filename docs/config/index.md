# Configuration

## Overview

There are two different ways to define configuration in FTPGrab:

* In a [configuration file](#configuration-file)
* As [environment variables](#environment-variables)

These ways are evaluated in the order listed above.

If no value was provided for a given option, a default value applies. Moreover, if an option has sub-options, and any of these sub-options is not specified, a default value will apply as well.

For example, the `FTPGRAB_DB` environment variable is enough by itself to enable the database, even though sub-options like `FTPGRAB_DB_PATH` exist. Once positioned, this option sets (and resets) all the default values of the sub-options of `FTPGRAB_DB`.

## Configuration file

At startup, FTPGrab searches for a file named `ftpgrab.yml` (or `ftpgrab.yaml`) in:

* `/etc/ftpgrab/`
* `$XDG_CONFIG_HOME/`
* `$HOME/.config/`
* `.` _(the working directory)_

You can override this using the [`--config` flag or `CONFIG` env var](../usage/cli.md).

??? example "ftpgrab.yml"
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
      chmodFile: 0644
      chmodDir: 0755
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

## Environment variables

All configuration from file can be transposed into environment variables. As an example, the following configuration:

??? example "ftpgrab.yml"
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
          - /src1
          - /src2
        timeout: 5s
        disableEPSV: false
        tls: false
        insecureSkipVerify: false
        logTrace: false
    
    download:
      output: /downloads
      uid: 1000
      gid: 1000
      chmodFile: 0644
      chmodDir: 0755
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
    ```

Can be transposed to:

??? example "environment variables"
    ```
    FTPGRAB_DB_PATH=ftpgrab.db
    
    FTPGRAB_SERVER_FTP_HOST=test.rebex.net
    FTPGRAB_SERVER_FTP_PORT=21
    FTPGRAB_SERVER_FTP_USERNAME=demo
    FTPGRAB_SERVER_FTP_PASSWORD=password
    FTPGRAB_SERVER_FTP_SOURCES=/src1,/src2
    FTPGRAB_SERVER_FTP_TIMEOUT=5s
    FTPGRAB_SERVER_FTP_DISABLEEPSV=false
    FTPGRAB_SERVER_FTP_TLS=false
    FTPGRAB_SERVER_FTP_INSECURESKIPVERIFY=false
    FTPGRAB_SERVER_FTP_LOGTRACE=false
    
    FTPGRAB_DOWNLOAD_OUTPUT=/downloads
    FTPGRAB_DOWNLOAD_GID=1000
    FTPGRAB_DOWNLOAD_UID=1000
    FTPGRAB_DOWNLOAD_CHMODDIR=493
    FTPGRAB_DOWNLOAD_CHMODFILE=420
    FTPGRAB_DOWNLOAD_INCLUDE=^Mr\.Robot\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+
    FTPGRAB_DOWNLOAD_EXCLUDE=\.nfo$
    FTPGRAB_DOWNLOAD_SINCE=2019-02-01T18:50:05Z
    FTPGRAB_DOWNLOAD_RETRY=3
    FTPGRAB_DOWNLOAD_HIDESKIPPED=false
    FTPGRAB_DOWNLOAD_CREATEBASEDIR=false
    
    FTPGRAB_NOTIF_MAIL_HOST=smtp.example.com
    FTPGRAB_NOTIF_MAIL_PORT=587
    FTPGRAB_NOTIF_MAIL_SSL=false
    FTPGRAB_NOTIF_MAIL_INSECURESKIPVERIFY=false
    FTPGRAB_NOTIF_MAIL_FROM=ftpgrab@example.com
    FTPGRAB_NOTIF_MAIL_TO=webmaster@example.com
    ```

## Reference

* [db](db.md)
* server
  * [ftp](server/ftp.md)
  * [sftp](server/sftp.md)
* [download](download.md)
* notif
  * [mail](notif/mail.md)
  * [slack](notif/slack.md)
  * [webhook](notif/webhook.md)
