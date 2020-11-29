# Webhook notifications

You can send webhook notifications with the following settings.

## Configuration

!!! example "File"
    ```yaml
    notif:
      webhook:
        endpoint: http://webhook.foo.com/sd54qad89azd5a
        method: GET
        headers:
          content-type: application/json
          authorization: Token123456
        timeout: 10s
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_NOTIF_WEBHOOK_ENDPOINT`
    * `FTPGRAB_NOTIF_WEBHOOK_METHOD`
    * `FTPGRAB_NOTIF_WEBHOOK_HEADERS_<KEY>`
    * `FTPGRAB_NOTIF_WEBHOOK_TIMEOUT`

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `endpoint`[^1]     |               | URL of the HTTP request |
| `method`[^1]       | `GET`         | HTTP method |
| `headers`          |               | Map of additional headers to be sent (key is case insensitive) |
| `timeout`          | `10s`         | Timeout specifies a time limit for the request to be made |

## Sample

The JSON response will look like this:

```json
{
  "ftpgrab_version": "5.2.0",
  "server_ip": "10.0.0.1",
  "dest_hostname": "my-computer",
  "journal": {
    "entries": [
      {
        "file": "/test/test_changed/1GB.bin",
        "status": "Not included",
        "level": "skip"
      },
      {
        "file": "/test/test_changed/56a42b12df8d27baa163536e7b10d3c7.png",
        "status": "Not included",
        "level": "skip"
      },
      {
        "file": "/test/test_special_chars/1024.rnd",
        "status": "Never downloaded",
        "level": "success",
        "text": "1.049MB successfully downloaded in 513 milliseconds"
      }
    ],
    "count": {
      "success": 1,
      "skip": 2
    },
    "duration": "12 seconds"
  }
}
```

[^1]: Value required
