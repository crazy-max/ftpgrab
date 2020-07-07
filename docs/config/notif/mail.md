# Mail notifications

Notifications can be sent through SMTP.

## Configuration

!!! example
    ```yaml
    notif:
      mail:
        enable: true
        host: localhost
        port: 25
        ssl: false
        insecure_skip_verify: false
        from: ftpgrab@example.com
        to: webmaster@example.com
    ```

| Name                   | Default       | Description   |
|------------------------|---------------|---------------|
| `enable`[^1]           | `false`       | Enable email reports |
| `host`[^1]             | `localhost`   | SMTP server host |
| `port`[^1]             | `25`          | SMTP server port |
| `ssl`                  | `false`       | SSL defines whether an SSL connection is used. Should be false in most cases since the auth mechanism should use STARTTLS |
| `insecure_skip_verify` | `false`       | Controls whether a client verifies the server's certificate chain and hostname |
| `username`             |               | SMTP username |
| `password`             |               | SMTP password |
| `from`[^1]             |               | Sender email address |
| `to`[^1]               |               | Recipient email address |

## Sample

![](../../assets/notif/mail.png)

[^1]: Value required
