# Download configuration

!!! example
    ```yaml
    download:
      output: /download
      uid:
      gid:
      chmod_file: 0644
      chmod_dir: 0755
      include:
      exclude:
      since: 0001-01-01T00:00:00Z
      retry: 3
      hide_skipped: false
      create_basedir: false
    ```

## `output`

Output destination folder of downloaded files. Env var `FTPGRAB_DOWNLOAD_OUTPUT` overrides this value.

!!! example
    ```yaml
    download:
      output: /download
    ```

## `uid`

Owner user applied to downloaded files. (default to caller)

!!! example
    ```yaml
    download:
      uid: 1000
    ```

## `gid`

Owner group applied to downloaded files. (default to caller)

!!! example
    ```yaml
    download:
      gid: 1000
    ```

## `chmod_file`

Permissions applied to files. (default: `0644`)

!!! example
    ```yaml
    download:
      chmod_file: 0644
    ```

## `chmod_dir`

Permissions applied to folders. (default: `0755`)

!!! example
    ```yaml
    download:
      chmod_dir: 0755
    ```

## `include`

List of regular expressions to include files.

!!! example
    ```yaml
    download:
      include:
        - ^Mr\.Robot\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+
    ```

## `exclude`

List of regular expressions to exclude files.

!!! example
    ```yaml
    download:
      exclude:
        - \.nfo$
    ```

## `since`

Only download files created since the specified date in RFC3339 format.

!!! example
    ```yaml
    download:
      since: 2019-02-01T18:50:05Z
    ```

## `retry`

Number of retries in case of download failure. (default: `3`)

!!! example
    ```yaml
    download:
      retry: 3
    ```

## `hide_skipped`

Not display skipped downloads. (default: `false`)

!!! example
    ```yaml
    download:
      hide_skipped: false
    ```

## `create_basedir`

Create basename of a FTP source path in the destination folder. This is highly recommended if you have multiple FTP
source paths to prevent overwriting. (default: `false`)

!!! warning
    Does not apply if `sources` is `/` only.

!!! example
    ```yaml
    download:
      create_basedir: false
    ```
