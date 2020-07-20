# Run as service on Debian based distro

## Using systemd

!!! warning
    Make sure to follow the instructions to [install from binary](binary.md) before.

To create a new service, paste this content in `/etc/systemd/system/ftpgrab.service`:

```
[Unit]
Description=FTPGrab
Documentation={{ config.site_url }}
After=syslog.target
After=network.target

[Service]
RestartSec=2s
Type=simple
User=ftpgrab
Group=ftpgrab
ExecStart=/usr/local/bin/ftpgrab --config /etc/ftpgrab/ftpgrab.yml --schedule "*/30 * * * *" --log-level info
Restart=always
Environment=FTPGRAB_DB=/var/lib/ftpgrab/ftpgrab.db

[Install]
WantedBy=multi-user.target
```

Change the user, group, and other required startup values following your needs.

Enable and start FTPGrab at boot:

```shell
$ sudo systemctl enable ftpgrab
$ sudo systemctl start ftpgrab
```

To view logs:

```shell
$ journalctl -fu ftpgrab.service
```
