# Download configuration

!!! example
    ```yaml
    download:
      output: /download
      uid: 1000
      gid: 1000
      chmodFile: 0644
      chmodDir: 0755
      include:
        - ^Mr\.Robot\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+
      exclude:
        - \.nfo$
      since: 2019-02-01T18:50:05Z
      retry: 3
      hideSkipped: false
      createBaseDir: false
    ```

## `output`

Output destination folder of downloaded files. Env var `FTPGRAB_DOWNLOAD_OUTPUT` overrides this value.

!!! example "Config file"
    ```yaml
    download:
      output: /download
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_OUTPUT`

## `uid`

Owner user applied to downloaded files. (default to caller)

!!! example "Config file"
    ```yaml
    download:
      uid: 1000
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_UID`

## `gid`

Owner group applied to downloaded files. (default to caller)

!!! example "Config file"
    ```yaml
    download:
      gid: 1000
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_GID`

## `chmodFile`

Permissions applied to files. (default: `0644`)

!!! example "Config file"
    ```yaml
    download:
      chmodFile: 0644
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_CHMODFILE`

## `chmodDir`

Permissions applied to folders. (default: `0755`)

!!! example "Config file"
    ```yaml
    download:
      chmodDir: 0755
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_CHMODDIR`

## `include`

List of regular expressions to include files.

!!! example "Config file"
    ```yaml
    download:
      include:
        - ^Mr\.Robot\.S04.+(VOSTFR|SUBFRENCH).+(720p).+(HDTV|WEB-DL|WEBRip).+
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_INCLUDE`

## `exclude`

List of regular expressions to exclude files.

!!! example "Config file"
    ```yaml
    download:
      exclude:
        - \.nfo$
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_EXCLUDE`

## `since`

Only download files created since the specified date in RFC3339 format.

!!! example "Config file"
    ```yaml
    download:
      since: 2019-02-01T18:50:05Z
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_SINCE`

## `retry`

Number of retries in case of download failure. (default: `3`)

!!! example "Config file"
    ```yaml
    download:
      retry: 3
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_RETRY`

## `hideSkipped`

Not display skipped downloads. (default: `false`)

!!! example "Config file"
    ```yaml
    download:
      hideSkipped: false
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_HIDESKIPPED`

## `createBaseDir`

Create basename of a FTP source path in the destination folder. This is highly recommended if you have multiple FTP
source paths to prevent overwriting. (default: `false`)

!!! warning
    Does not apply if `sources` is `/` only.

!!! example "Config file"
    ```yaml
    download:
      createBaseDir: false
    ```

!!! abstract "Environment variables"
    * `FTPGRAB_DOWNLOAD_CREATEBASEDIR`
