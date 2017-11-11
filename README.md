<p align="center"><a href="https://ftpgrab.github.io" target="_blank"><img width="100" src="https://ftpgrab.github.io/img/logo.png"></a></p>

<p align="center">
  <a href="https://travis-ci.org/ftpgrab/ftpgrab"><img src="https://img.shields.io/travis/ftpgrab/ftpgrab/master.svg?style=flat-square" alt="Build Status"></a>
  <a href="https://www.codacy.com/app/crazy-max/ftpgrab"><img src="https://img.shields.io/codacy/grade/354bfb181fc5482dac1e8f31e8e29af5.svg?style=flat-square" alt="Code Quality"></a>
  <a href="https://github.com/ftpgrab/ftpgrab/releases/latest"><img src="https://img.shields.io/github/release/ftpgrab/ftpgrab.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=7NFD44VBNE3VL"><img src="https://img.shields.io/badge/donate-paypal-7057ff.svg?style=flat-square" alt="Donate Paypal"></a>
</p>

## About

**FTPGrab** (formerly *FTP Sync*) is a shell script to grab your files from a remote FTP server to your NAS / server / computer.<br />

A file containing the hash of the name of each downloaded file will prevent re-download a file even if it is not present in the destination directory.<br />

You can also apply a filter to search for files with a regular expression.<br />

Because this script only need `wget`, it is ideal for those with a seedbox or a shared seedbox to synchronize with a NAS (Synology Qnap D-Link) or a local computer...<br />

With the sqlite3 [HASH_STORAGE](https://ftpgrab.github.io/doc/configuration/#hash_storage), the process performance will be improved!.

You can install FTPGrab using Docker !<br />
An [official docker image](https://hub.docker.com/r/crazymax/ftpgrab/) 🐳 is available for FTPGrab. For more info, have a look on the [docker repository](https://github.com/ftpgrab/docker).

Before reporting an issue, please read the [Troubleshooting page](https://ftpgrab.github.io/doc/troubleshooting).<br />

## Documentation

* [Get started](https://ftpgrab.github.io/doc/get-started)
* [Configuration](https://ftpgrab.github.io/doc/configuration)
* [Troubleshooting](https://ftpgrab.github.io/doc/troubleshooting)
* [Changelog](https://ftpgrab.github.io/doc/changelog)
* [Upgrade notes](https://ftpgrab.github.io/doc/upgrade-notes)
* [Reporting an issue](https://ftpgrab.github.io/doc/reporting-issue)

## Logs

Each time the script is executed, a log file is created prefiexd by the config used.<br />
Here is an example :

```console
FTPGrab v4.1 (seedbox - 2017/03/15 01:41:49)
--------------
Config: seedbox
Script PID: 19383
Log file: /var/log/ftpgrab/seedbox-20170315014149.log
FTP sources count: 1
FTP secure: 1
Download method: curl
Resume downloads: 1
Hash type: md5
Hash storage: sqlite3
Hash file: /opt/ftpgrab/hash/seedbox.db
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

The hash file looks like this if you used `text` as `HASH_STORAGE` :

```console
baf87b6719e9f5499627fc8691efbd3c Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi
```

## How can i help ?

**FTPGrab** is free and open source and always will be.<br />
We welcome all kinds of contributions :raised_hands:!<br />
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:<br />
Any funds donated will be used to help further development on this project! :gift_heart:

[![Donate Paypal](https://ftpgrab.github.io/img/paypal.png)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=7NFD44VBNE3VL)

## License

MIT. See `LICENSE` for more details.<br />
Icon credit to [Nick Roach](http://www.elegantthemes.com/).
