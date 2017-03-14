<p align="center"><a href="https://github.com/crazy-max/ftp-sync" target="_blank"><img width="100" src="https://raw.githubusercontent.com/wiki/crazy-max/ftp-sync/img/logo-128.png"></a></p>

<p align="center">
  <a href="https://github.com/crazy-max/ftp-sync/releases/latest"><img src="https://img.shields.io/github/release/crazy-max/ftp-sync.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=7NFD44VBNE3VL"><img src="https://img.shields.io/badge/donate-paypal-blue.svg?style=flat-square" alt="Donate Paypal"></a>
  <a href="https://flattr.com/submit/auto?user_id=crazymax&url=https://github.com/crazy-max/ftp-sync"><img src="https://img.shields.io/badge/flattr-this-green.svg?style=flat-square" alt="Flattr this!"></a>
</p>

## About

**FTP Sync** is a shell script to synchronize files between a remote FTP server and your local server / computer.<br />

A file containing the hash of the name of each downloaded file will prevent re-download a file even if it is not present in the destination directory.<br />

You can also apply a filter to search for files with a regular expression.<br />

Because this script only need `wget`, it is ideal for those with a seedbox or a shared seedbox to synchronize with a NAS (Synology Qnap D-Link) or a local computer...<br />

With the sqlite3 [HASH_STORAGE](../../wiki/Configuration#hash_storage), the process performance will be improved!.

Before reporting an issue, please read the [Troubleshooting page](../../wiki/Troubleshooting).<br />
Do not forget to star :star2: the project if you like it :heart_eyes:

## Documentation

* [Requirements](../../wiki/Requirements)
* [Installation](../../wiki/Installation)
* [Upgrade to 4.x](../../wiki/Upgrade-to-4.x)
* [Configuration](../../wiki/Configuration)
* [Usage](../../wiki/Usage)
* [Troubleshooting](../../wiki/Troubleshooting)

## Logs

Each time the script is executed, a log file is created prefiexd by the config used.<br />
Here is an example :

```console
FTP Sync v4.0 (seedbox - 2017/03/14 01:41:49)
--------------
Config: seedbox
Script PID: 19383
Log file: /var/log/ftp-sync/seedbox-20170314014149.log
FTP sources count: 1
FTP secure: 1
Download method: curl
Resume downloads: 1
Hash type: md5
Hash storage: sqlite3
Hash file: /opt/ftp-sync/hash/seedbox.db
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

I'm a single developer and if you find this project useful, please consider making a donation.<br />
Any funds donated will be used to help further development on this project! :gift_heart:

<p>
  <a href="https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=7NFD44VBNE3VL">
    <img src="../../wiki/img/paypal.png" alt="Donate Paypal">
  </a>
  <a href="https://flattr.com/submit/auto?user_id=crazymax&url=https://github.com/crazy-max/ftp-sync">
    <img src="../../wiki/img/flattr.png" alt="Flattr this!">
  </a>
</p>

## License

MIT. See `LICENSE` for more details.<br />
Icon credit to [Nick Roach](http://www.elegantthemes.com/).
