# FTP Sync [![Donate Paypal](https://img.shields.io/badge/donate-paypal-blue.svg)](https://www.paypal.me/crazyws)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [About](#about)
- [Get Started](#get-started)
- [Logs](#logs)
- [License](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## About

A shell script to synchronize files between a remote FTP server and your local server/computer.<br />

A file containing the hash of the name of each downloaded file will prevent re-download a file even if it is not present in the destination directory.<br />

You can also apply a filter to search for files with a regular expression.<br />

Because this script only need `wget`, it is ideal for those with a seedbox or a shared seedbox to synchronize with a NAS (Synology Qnap D-Link) or a local computer...<br />

With the sqlite3 [HASH_STORAGE](../../wiki/Configuration#hash_storage), the process performance will be improved!.

Before reporting an issue, please read the [Troubleshooting page](../../wiki/Troubleshooting).<br />
To be notified of new releases you can Star / Watch the project.

## Get Started

* [Installation](../../wiki/Installation) instructions.
* [Configuration](../../wiki/Configuration) instructions.
* [Usage](../../wiki/Usage) instructions.

## Logs

Each time the script is executed, a log file is created.<br />
Here is an example :

```console
FTP Sync v3.1 (2016/03/27 19:59:13)
--------------
Script PID: 19383
Log file: /var/log/ftp-sync/20160320120930.log
FTP sources count: 1
FTP secure: 1
Download method: curl
Resume downloads: 1
Hash type: md5
Hash storage: sqlite3
Hash file: /etc/ftp-sync/ftp-sync.db
--------------
Source: ftp://198.51.100.0:21/complete/
Destination: /tmp/seedbox/
Checking connection to ftp://198.51.100.0:21/complete/...
Successfully connected!
--------------
Process file: Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi
Hash: baf87b6719e9f5499627fc8691efbd3c
Size: 184.18 Mb
Status: Never downloaded...
Start download to /tmp/seedbox/Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi... Please wait...
File successfully downloaded!
Time spent: 00:00:48
--------------
Change the ownership recursively of 'Destination' path to ftpuser:ftpgroup
Change the access permissions recursively of 'Destination' path to 755
--------------
Finished...
Total time spent: 00:00:49
```

The hash file looks like this :

```console
baf87b6719e9f5499627fc8691efbd3c Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi
```

## License

MIT. See ``LICENSE`` for more details.
