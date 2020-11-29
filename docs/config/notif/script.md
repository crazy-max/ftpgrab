# Script notifications

You can call a script when a notification occured. Following environment variables will be passed:

```
FTPGRAB_VERSION=3.0.0
FTPGRAB_SERVER_IP=10.0.0.1
FTPGRAB_DEST_HOSTNAME=my-computer
FTPGRAB_JOURNAL_ENTRIES[0]_FILE=/test/test_changed/1GB.bin
FTPGRAB_JOURNAL_ENTRIES[0]_STATUS=Not included
FTPGRAB_JOURNAL_ENTRIES[0]_LEVEL=skip
FTPGRAB_JOURNAL_ENTRIES[0]_TEXT=
FTPGRAB_JOURNAL_ENTRIES[1]_FILE=/test/test_changed/56a42b12df8d27baa163536e7b10d3c7.png
FTPGRAB_JOURNAL_ENTRIES[1]_STATUS=Not included
FTPGRAB_JOURNAL_ENTRIES[1]_LEVEL=skip
FTPGRAB_JOURNAL_ENTRIES[1]_TEXT=
FTPGRAB_JOURNAL_ENTRIES[2]_FILE=/test/test_special_chars/1024.rnd
FTPGRAB_JOURNAL_ENTRIES[2]_STATUS=Never downloaded
FTPGRAB_JOURNAL_ENTRIES[2]_LEVEL=success
FTPGRAB_JOURNAL_ENTRIES[2]_TEXT=1.049MB successfully downloaded in 513 milliseconds
FTPGRAB_JOURNAL_COUNT_SUCCESS=1
FTPGRAB_JOURNAL_COUNT_SKIP=2
FTPGRAB_JOURNAL_COUNT_ERROR=0
FTPGRAB_JOURNAL_DURATION=12 seconds
```

## Configuration

!!! example "File"
    ```yaml
    notif:
      script:
        cmd: "myprogram"
        args:
          - "--anarg"
          - "another"
    ```

| Name                  | Default       | Description   |
|-----------------------|---------------|---------------|
| `cmd`[^1]             |               | Command or script to execute |
| `args`                |               | List of args to pass to `cmd` |
| `dir`                 |               | Specifies the working directory of the command |

!!! abstract "Environment variables"
    * `FTPGRAB_NOTIF_SCRIPT_CMD`
    * `FTPGRAB_NOTIF_SCRIPT_ARGS`
    * `FTPGRAB_NOTIF_SCRIPT_DIR`

[^1]: Value required
