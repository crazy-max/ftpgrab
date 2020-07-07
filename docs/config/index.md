# Configuration

## Example

You can define a configuration file through the [`--config` flag](../usage/cli.md) with the following content:

??? example "ftpgrab.yml"
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
        insecure_skip_verify: false
        log_trace: false
    
    db:
      enable: true
      path: ftpgrab.db
    
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

## Reference

* server
    * [ftp](server/ftp.md)
    * [sftp](server/sftp.md)
* [db](db.md)
* [download](download.md)
* notif
    * [mail](notif/mail.md)
    * [slack](notif/slack.md)
    * [webhook](notif/webhook.md)
