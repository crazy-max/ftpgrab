# Command Line

## Usage

```shell
ftpgrab [options]
```

## Options

```
$ ftpgrab --help
Usage: ftpgrab

Grab your files periodically from a remote FTP or SFTP server easily. More info:
https://github.com/crazy-max/ftpgrab

Flags:
  -h, --help                Show context-sensitive help.
      --version
      --config=STRING       FTPGrab configuration file ($CONFIG).
      --schedule=STRING     CRON expression format ($SCHEDULE).
      --log-level="info"    Set log level ($LOG_LEVEL).
      --log-json            Enable JSON logging output ($LOG_JSON).
      --log-timestamp       Adds the current local time as UNIX timestamp to the
                            logger context ($LOG_TIMESTAMP).
      --log-caller          Add file:line of the caller to log output
                            ($LOG_CALLER).
      --log-file=STRING     Add logging to a specific file ($LOG_FILE).
```

## Environment variables

Following environment variables can be used in place:

| Name               | Default       | Description   |
|--------------------|---------------|---------------|
| `CONFIG`           |               | FTPGrab configuration file |
| `SCHEDULE`         |               | [CRON expression](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format) to schedule FTPGrab |
| `LOG_LEVEL`        | `info`        | Log level output |
| `LOG_JSON`         | `false`       | Enable JSON logging output |
| `LOG_TIMESTAMP`    | `true`        | Adds the current local time as UNIX timestamp to the logger context |
| `LOG_CALLER`       | `false`       | Enable to add `file:line` of the caller |
| `LOG_FILE`         |               | Add logging to a specific file |
