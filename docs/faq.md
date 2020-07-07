# FAQ

## Who is behind FTPGrab?

Hi, I'm [CrazyMax](https://crazymax.dev). This project is self-funded and developed using my decade of experience
building open source software.

By [supporting me](https://github.com/sponsors/crazy-max), you're not only sustaining this project,
but rather all of [my open source projects](https://github.com/crazy-max).

## How to grab from multiple sources?

You can add multiple sources in the `sources` field of the configuration file:

```yaml
ftp|sftp:
  ...
  sources:
    - /path1
    - /path2/folder
```

## What kind of CRON expression can I use for scheduling?

A CRON expression represents a set of times, using 6 space-separated fields.

* `*/30 * * * *` will launch a job every 30 minutes.
* `*/15 * * * * *` will launch a job every 15 seconds.

More examples can be found on the [official library documentation](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format).

## What Regexp semantic is used to filter inclusions/exclusions?

FTPGrab uses [Compile](https://golang.org/pkg/regexp/#Compile) to parse regular expressions. This means the regexp
returns a match that begins as early as possible in the input (leftmost) like Perl, Python, and other implementations
use. You can test your regular expression on [regex101.com](https://regex101.com/) and select Golang
flavor. Check this [quick example](https://regex101.com/r/jITi0D/1).

## What logs look like?

Here is a sample output:

```text
Tue, 29 Jan 2019 22:23:58 CET INF Starting FTPGrab 5.0.0
Tue, 29 Jan 2019 22:23:58 CET INF ########
Tue, 29 Jan 2019 22:23:58 CET INF Connecting to 198.51.100.0:21...
Tue, 29 Jan 2019 22:23:58 CET INF Grabbing from /complete/
Tue, 29 Jan 2019 22:23:59 CET INF --------
Tue, 29 Jan 2019 22:23:59 CET INF Checking /complete/Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi
Tue, 29 Jan 2019 22:23:59 CET INF Never downloaded
Tue, 29 Jan 2019 22:23:59 CET INF Downloading file (184.18MB) to /tmp/seedbox/Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi...
Tue, 29 Jan 2019 22:24:47 CET INF File successfully downloaded!
Tue, 29 Jan 2019 22:24:47 CET INF Time spent: 48 seconds
Tue, 29 Jan 2019 22:24:47 CET INF --------
Tue, 29 Jan 2019 22:24:47 CET INF Checking /complete/Burn.Notice.S06E17.VOSTFR.HDTV.XviD.avi
Tue, 29 Jan 2019 22:24:47 CET INF Never downloaded
Tue, 29 Jan 2019 22:24:47 CET INF Downloading file (186.27MB) to /tmp/seedbox/Burn.Notice.S06E17.VOSTFR.HDTV.XviD.avi...
Tue, 29 Jan 2019 22:25:40 CET INF File successfully downloaded!
Tue, 29 Jan 2019 22:25:40 CET INF Time spent: 50 seconds
Tue, 29 Jan 2019 22:25:40 CET INF --------
Tue, 29 Jan 2019 22:25:40 CET INF Checking /complete/test_special_chars/123.bin
Tue, 29 Jan 2019 22:25:40 CET INF Never downloaded
Tue, 29 Jan 2019 22:25:40 CET INF Downloading file (33.27kB) to /tmp/seedbox/test/test_special_chars/123.bin...
Tue, 29 Jan 2019 22:25:42 CET ERR Error downloading, retry 1/3 error="dial tcp 198.51.100.0:21: connect: connection refused"
Tue, 29 Jan 2019 22:25:42 CET INF --------
Tue, 29 Jan 2019 22:25:42 CET INF Checking /complete/test_special_chars/123.bin
Tue, 29 Jan 2019 22:25:42 CET INF Exists but size is different
Tue, 29 Jan 2019 22:25:42 CET INF Downloading file (33.27kB) to /tmp/seedbox/test/test_special_chars/123.bin...
Tue, 29 Jan 2019 22:25:44 CET ERR Error downloading, retry 2/3 error="dial tcp 198.51.100.0:21: connect: connection refused"
Tue, 29 Jan 2019 22:25:44 CET INF --------
Tue, 29 Jan 2019 22:25:44 CET INF Checking /complete/test_special_chars/123.bin
Tue, 29 Jan 2019 22:25:44 CET INF Exists but size is different
Tue, 29 Jan 2019 22:25:44 CET INF Downloading file (33.27kB) to /tmp/seedbox/test/test_special_chars/123.bin...
Tue, 29 Jan 2019 22:25:46 CET ERR Error downloading, retry 3/3 error="dial tcp 198.51.100.0:21: connect: connection refused"
Tue, 29 Jan 2019 22:25:46 CET ERR Cannot download file error="dial tcp 198.51.100.0:21: connect: connection refused"
Tue, 29 Jan 2019 22:25:46 CET INF Time spent: 6 seconds
Tue, 29 Jan 2019 22:25:46 CET INF --------
Tue, 29 Jan 2019 22:25:46 CET INF Checking /complete/exlcuded_file.txt
Tue, 29 Jan 2019 22:25:46 CET INF Not included
Tue, 29 Jan 2019 22:25:46 CET WRN Skipped: Not included
Tue, 29 Jan 2019 22:25:46 CET INF ########
Tue, 29 Jan 2019 22:25:51 CET INF Finished, total time spent: 1 minute 56 seconds
```

## How can I edit/remove some entries in the database?

FTPGrab currently uses the embedded key/value database [bbolt](https://github.com/etcd-io/bbolt).

You can use [boltBrowser](https://github.com/ShoshinNikita/boltBrowser) which is a GUI web-based explorer and editor
or [this CLI browser](https://github.com/br0xen/boltbrowser) to remove some entries.
