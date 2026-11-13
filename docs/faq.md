# FAQ

## Timezone

By default, all interpretation and scheduling is done with your local timezone (`TZ` environment variable).

Cron schedule may also override the timezone to be interpreted in by providing an additional space-separated field
at the beginning of the cron spec, of the form `CRON_TZ=<timezone>`:

```shell
ftpgrab --schedule "CRON_TZ=Asia/Tokyo */30 * * * *"
```

## What kind of CRON expression can I use for scheduling?

A CRON expression represents a set of times, using 6 space-separated fields.

* `*/30 * * * *` will launch a job every 30 minutes.
* `*/15 * * * * *` will launch a job every 15 seconds.

More examples can be found on the [official library documentation](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format).

## How to grab from multiple sources?

You can add multiple sources in the `sources` field of the configuration file:

```yaml
ftp|sftp:
  ...
  sources:
    - /path1
    - /path2/folder
```

## What Regexp semantic is used to filter inclusions/exclusions?

FTPGrab uses [Compile](https://golang.org/pkg/regexp/#Compile) to parse regular expressions. This means the regexp
returns a match that begins as early as possible in the input (leftmost) like Perl, Python, and other implementations
use. You can test your regular expression on [regex101.com](https://regex101.com/) and select Golang
flavor. Check this [quick example](https://regex101.com/r/jITi0D/1).

## How can I avoid downloading incomplete uploads?

FTPGrab does not detect whether a remote file is still being written. Configure
the uploader to use a temporary suffix, such as `.part` or `.writing`, while
uploading a file, and rename it to its final name only after the upload completes.
Exclude those temporary names with
[`download.exclude`](config/download.md#exclude):

```yaml
download:
  exclude:
    - '\.(part|writing)$'
```

For example, `report.csv.part` is skipped. After the uploader finishes and renames
it to `report.csv`, the file becomes eligible for download on the next run.

This approach requires the uploader to follow the naming convention. An unchanged
size or modification time over a short interval does not guarantee completion,
since an upload may be paused. The [`tempFirst`](config/download.md#tempfirst)
option only controls local temporary files; it does not check whether a remote
upload has finished.

## How can I edit/remove some entries in the database?

FTPGrab currently uses the embedded key/value database [bbolt](https://github.com/etcd-io/bbolt).

You can use [boltBrowser](https://github.com/ShoshinNikita/boltBrowser) which is a GUI web-based explorer and editor
or [this CLI browser](https://github.com/br0xen/boltbrowser) to remove some entries.
