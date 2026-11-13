# Discord notifications

Send notifications to a Discord channel using an incoming webhook.

## Configuration

!!! example "File"
    ```yaml
    notif:
      discord:
        webhookURL: https://discord.com/api/webhooks/1234567890/token
        mentions:
          - "<@123>"
        renderEmbeds: true
        renderFields: true
        timeout: 10s
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_NOTIF_DISCORD_WEBHOOKURL`
    * `FTPGRAB_NOTIF_DISCORD_WEBHOOKURLFILE`
    * `FTPGRAB_NOTIF_DISCORD_MENTIONS` (comma separated)
    * `FTPGRAB_NOTIF_DISCORD_RENDEREMBEDS`
    * `FTPGRAB_NOTIF_DISCORD_RENDERFIELDS`
    * `FTPGRAB_NOTIF_DISCORD_TIMEOUT`
    * `FTPGRAB_NOTIF_DISCORD_PROXY`

| Name             | Default   | Description                                        |
| ---------------- | --------- | -------------------------------------------------- |
| `webhookURL`     |           | Discord incoming webhook URL                       |
| `webhookURLFile` |           | Secret file used when `webhookURL` is unset        |
| `mentions`       |           | Users or roles to notify                           |
| `renderEmbeds`   | `true`    | Render a message embed                             |
| `renderFields`   | `true`    | Include server and destination fields in the embed |
| `timeout`        | `10s`     | Timeout for each HTTP request                      |
| `proxy`          |           | HTTP proxy URL to use for requests                 |
