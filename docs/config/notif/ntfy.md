# ntfy notifications

Send notifications using an [ntfy](https://ntfy.sh/) instance.

## Configuration

!!! example "File"
    ```yaml
    notif:
      ntfy:
        endpoint: https://ntfy.sh
        topic: ftpgrab-reports
        priority: 3
        tags:
          - package
        timeout: 10s
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_NOTIF_NTFY_ENDPOINT`
    * `FTPGRAB_NOTIF_NTFY_TOKEN`
    * `FTPGRAB_NOTIF_NTFY_TOKENFILE`
    * `FTPGRAB_NOTIF_NTFY_TOPIC`
    * `FTPGRAB_NOTIF_NTFY_PRIORITY`
    * `FTPGRAB_NOTIF_NTFY_TAGS` (comma separated)
    * `FTPGRAB_NOTIF_NTFY_ICON`
    * `FTPGRAB_NOTIF_NTFY_CLICK`
    * `FTPGRAB_NOTIF_NTFY_TIMEOUT`
    * `FTPGRAB_NOTIF_NTFY_PROXY`
    * `FTPGRAB_NOTIF_NTFY_TLSSKIPVERIFY`
    * `FTPGRAB_NOTIF_NTFY_TLSCACERTFILES` (comma separated)

| Name             | Default           | Description                                  |
| ---------------- | ----------------- | -------------------------------------------- |
| `endpoint`       | `https://ntfy.sh` | ntfy base URL                                |
| `token`          |                   | Access token                                 |
| `tokenFile`      |                   | Secret file used when `token` is unset       |
| `topic`          |                   | Topic to publish to (required)               |
| `priority`       | `3`               | Message priority (0-5)                       |
| `tags`           | `["package"]`     | Notification tags                            |
| `icon`           | FTPGrab logo      | Notification icon URL                        |
| `click`          |                   | URL to open when the notification is clicked |
| `timeout`        | `10s`             | Timeout for the HTTP request                 |
| `proxy`          |                   | HTTP proxy URL to use for requests           |
| `tlsSkipVerify`  | `false`           | Skip TLS certificate verification            |
| `tlsCaCertFiles` |                   | Custom CA PEM files (replace system roots)   |
