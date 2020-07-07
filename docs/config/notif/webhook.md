# Webhook notifications

You can send webhook notifications with the following settings.

## Configuration

!!! example
    ```yaml
    notif:
      webhook:
        enable: true
        endpoint: http://webhook.foo.com/sd54qad89azd5a
        method: GET
        headers:
          content-type: application/json
          authorization: Token123456
        timeout: 10s
    ```

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `enable`[^1]       | `false`       | Enable webhook notification |
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
        "status_type": "skip",
        "status_text": "Not included"
      },
      {
        "file": "/test/test_changed/56a42b12df8d27baa163536e7b10d3c7.png",
        "status_type": "skip",
        "status_text": "Not included"
      },
      {
        "file": "/test/test_special_chars/1024.rnd",
        "status_type": "success",
        "status_text": "1.049MB successfully downloaded in 513 milliseconds"
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
