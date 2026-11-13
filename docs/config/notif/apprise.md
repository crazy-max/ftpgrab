# Apprise notifications

Send notifications using an Apprise API instance.

## Configuration

!!! example "File"
    ```yaml
    notif:
      apprise:
        endpoint: http://apprise:8000
        token: ftpgrab
        tags:
          - reports
        timeout: 10s
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_NOTIF_APPRISE_ENDPOINT`
    * `FTPGRAB_NOTIF_APPRISE_TOKEN`
    * `FTPGRAB_NOTIF_APPRISE_TOKENFILE`
    * `FTPGRAB_NOTIF_APPRISE_TAGS` (comma separated)
    * `FTPGRAB_NOTIF_APPRISE_URLS` (comma separated)
    * `FTPGRAB_NOTIF_APPRISE_TIMEOUT`
    * `FTPGRAB_NOTIF_APPRISE_PROXY`
    * `FTPGRAB_NOTIF_APPRISE_TLSSKIPVERIFY`
    * `FTPGRAB_NOTIF_APPRISE_TLSCACERTFILES` (comma separated)

| Name             | Default | Description                                |
| ---------------- | ------- | ------------------------------------------ |
| `endpoint`       |         | Apprise API base URL (required)            |
| `token`          |         | Config key for saved destinations          |
| `tokenFile`      |         | Secret file used when `token` is unset     |
| `tags`           |         | Tags selecting saved destinations          |
| `urls`           |         | Direct notification URLs                   |
| `timeout`        | `10s`   | Timeout for the HTTP request               |
| `proxy`          |         | HTTP proxy URL to use for requests         |
| `tlsSkipVerify`  | `false` | Skip TLS certificate verification          |
| `tlsCaCertFiles` |         | Custom CA PEM files (replace system roots) |
