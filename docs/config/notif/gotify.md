# Gotify notifications

Send notifications using a [Gotify](https://gotify.net/) instance.

## Configuration

!!! example "File"
    ```yaml
    notif:
      gotify:
        endpoint: https://gotify.example.com
        token: application-token
        priority: 1
        timeout: 10s
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_NOTIF_GOTIFY_ENDPOINT`
    * `FTPGRAB_NOTIF_GOTIFY_TOKEN`
    * `FTPGRAB_NOTIF_GOTIFY_TOKENFILE`
    * `FTPGRAB_NOTIF_GOTIFY_PRIORITY`
    * `FTPGRAB_NOTIF_GOTIFY_TIMEOUT`
    * `FTPGRAB_NOTIF_GOTIFY_PROXY`
    * `FTPGRAB_NOTIF_GOTIFY_TLSSKIPVERIFY`
    * `FTPGRAB_NOTIF_GOTIFY_TLSCACERTFILES` (comma separated)

| Name             | Default | Description                                |
| ---------------- | ------- | ------------------------------------------ |
| `endpoint`       |         | Gotify base URL (required)                 |
| `token`          |         | Application token                          |
| `tokenFile`      |         | Secret file used when `token` is unset     |
| `priority`       | `1`     | Message priority                           |
| `timeout`        | `10s`   | Timeout for the HTTP request               |
| `proxy`          |         | HTTP proxy URL to use for requests         |
| `tlsSkipVerify`  | `false` | Skip TLS certificate verification          |
| `tlsCaCertFiles` |         | Custom CA PEM files (replace system roots) |
