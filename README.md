FTP Sync
========

A shell script to synchronize files between a remote FTP server and your local server/computer.<br />
A file containing the MD5 hash of the name of each downloaded file will prevent re-download a file even if it is not present in the destination directory.<br />
You can also apply a filter to search for files with a regular expression.<br />
Because this script only need ``wget``, it is ideal for those with a seedbox or a shared seedbox to synchronize with a NAS (Synology Qnap D-Link) or a local computer...

Requirements
------------

Commands :

* [awk](http://en.wikipedia.org/wiki/Awk) is required.
* [mail](http://linux.die.net/man/1/mail) is optional if you do not fill EMAIL_LOG.
* [md5sum](http://en.wikipedia.org/wiki/Md5sum) is required.
* [nawk](http://linux.die.net/man/1/nawk) is required.
* [wget](http://en.wikipedia.org/wiki/Wget) is required.

Installation
------------

Execute the following commands to download the script :
```console
$ cd /etc/init.d/
$ wget https://raw.github.com/crazy-max/ftp-sync/master/ftp-sync.sh -O ftp-sync
$ chmod +x ftp-sync
$ update-rc.d ftp-sync defaults
```

Before running the script, you must change some vars.

* **FTP_HOST** - FTP host IP or domain. (e.g. 10.0.0.1 or ftp.example.com)
* **FTP_PORT** - FTP port. (e.g. 21)
* **FTP_USER** - FTP username.
* **FTP_PASSWORD** - FTP password.
* **FTP_SRC** - FTP path to synchronize.
* **DL_USER** - Linux owner user of downloaded files. Optional.
* **DL_GROUP** - Linux owner group of downloaded files. Optional.
* **DL_CHMOD** - Permissions of downloaded files. Optional. (e.g. 644)
* **DL_REGEX** - Apply a filter to search for files with a regular expression. Separate each regular expression with a semicolon. Leave empty to grab all files. Optional. (e.g. Game.Of.Thrones*.avi;Burn.Notice.*.avi)
* **DL_RETRY** - Number of retries in case of failure of download. (default 3)
* **DL_HIDE_SKIPPED** - Not display the downloads already made ​​or valid in logs. (default 0)
* **DL_HIDE_PROGRESS** - Not display the progress dots during downloads. (default 1)
* **MD5_ENABLED** - Enable audit file already downloaded.
* **MD5_FILE** - The audit file containing the hash of each downloaded file (default /etc/ftp-sync/ftp-sync.md5).
* **DIR_LOGS** - Path to save ftp-sync logs. (default /etc/ftp-sync/logs)
* **EMAIL_LOG** - Mail address where the logs are sent. Leave empty to disable sending mail.
* **PID_FILE** - Path to the file containing the current PID of the process.

Usage
-----

``$ /etc/init.d/ftp-sync <DIR_DEST>``

DIR_DEST is the directory where the files will be downloaded.
e.g. ``$ /etc/init.d/ftp-sync /tmp/seedbox/``

Automatic sync with cron
------------------------

You can automatically synchronize FTP files by calling the script in a [crontab](http://en.wikipedia.org/wiki/Crontab).
For example :

    0 4 * * * cd /etc/init.d/ && ./ftp-sync /tmp/seedbox/ >/dev/null 2>&1
	
This will synchronize your FTP files with the directory ``/tmp/seedbox/`` every day at 4 am.

Logs
----

Each time the script is executed, a log file is created.
Here is an example :

```console
FTP Sync v1.3 (2013/06/02 19:00:16)
Script PID: 32017
Source: ftp://10.0.0.1:21/complete/
Destination: /tmp/seedbox/
Log file: /etc/ftp-sync/logs/ftp-sync-20130602190016.log
MD5 file: /etc/ftp-sync/ftp-sync.md5
--------------
Finding files...
Regex: ^.*$
--------------
Process file : Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi
Hash: 5cc4931d64bd5579e46041b7139bde9f
Size: 351 Mb
Status : Already downloaded and valid. Skip download...
Time spent: 00:00:00
--------------
Process file : Burn.Notice.S06E17E18.VOSTFR.HDTV.XviD.avi
Hash: 5cc4931d64bd5579e46041b7139bde9f
Size: 703 Mb
Status : Already downloaded and valid. Skip download...
Time spent: 00:00:00
--------------
```

The MD5 file looks like this :

```console
baf87b6719e9f5499627fc8691efbd3c Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi
92d1d13049bd148f89ffa1d501f153f5 Burn.Notice.S06E17E18.VOSTFR.HDTV.XviD.avi
```

License
-------

LGPL. See ``LICENSE`` for more details.

More infos
----------

http://www.crazyws.fr/dev/systeme/synchroniser-votre-seedbox-avec-votre-nas-ou-votre-ordinateur-6NGGE.html