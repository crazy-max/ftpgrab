# Command Line

## Usage

```shell
$ ftpgrab [options]
```

## Options

```
$ ftpgrab --help
Usage: ftpgrab

Grab your files periodically from a remote FTP or SFTP server easily. More info:
https://github.com/crazy-max/ftpgrab

Flags:
  --help                Show context-sensitive help.
  --version
  --config=STRING       FTPGrab configuration file ($CONFIG).
  --schedule=STRING     CRON expression format ($SCHEDULE).
  --timezone="UTC"      Timezone assigned to FTPGrab ($TZ).
  --log-level="info"    Set log level ($LOG_LEVEL).
  --log-json            Enable JSON logging output ($LOG_JSON).
  --log-file=STRING     Add logging to a specific file ($LOG_FILE).
```

## Environment variables

Following environment variables can be used in place:

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `CONFIG`           |               | FTPGrab configuration file |
| `SCHEDULE`         |               | CRON expression format |
| `TZ`               | `UTC`         | Timezone assigned |
| `LOG_LEVEL`        | `info`        | Log level output |
| `LOG_JSON`         | `false`       | Enable JSON logging output |
| `LOG_CALLER`       | `false`       | Enable to add `file:line` of the caller |
| `LOG_FILE`         |               | Add logging to a specific file |
