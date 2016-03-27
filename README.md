# FTP Sync

A shell script to synchronize files between a remote FTP server and your local server/computer.<br />
A file containing the hash of the name of each downloaded file will prevent re-download a file even if it is not present in the destination directory.<br />
You can also apply a filter to search for files with a regular expression.<br />
Because this script only need `wget`, it is ideal for those with a seedbox or a shared seedbox to synchronize with a NAS (Synology Qnap D-Link) or a local computer...<br />
If you use the HASH_STORAGE called sqlite3, the process performance will be improved! (see [Configuration](#configuration) section)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Automatic sync with cron](#automatic-sync-with-cron)
- [Logs](#logs)
- [Troubleshooting](#troubleshooting)
  - [awk: line 1: syntax error at or near](#awk-line-1-syntax-error-at-or-near)
  - [Synology Network Attached Storage](#synology-network-attached-storage)
    - [bootstrap, ipkg](#bootstrap-ipkg)
    - [bash](#bash)
    - [coreutils](#coreutils)
    - [sqlite](#sqlite)
    - [nail](#nail)
    - [wget](#wget)
    - [crontab](#crontab)
- [Found a bug?](#found-a-bug)
- [License](#license)
- [More infos](#more-infos)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Requirements

* [awk](http://en.wikipedia.org/wiki/Awk) is required.
* [nawk](http://linux.die.net/man/1/nawk) is required.
* [gawk](http://www.gnu.org/software/gawk/) is required.
* [mail](http://linux.die.net/man/1/mail) is optional if you do not fill EMAIL_LOG.
* [wget](http://en.wikipedia.org/wiki/Wget) >= 1.12 is required.
* [md5sum](http://en.wikipedia.org/wiki/Md5sum) is required.
* [curl](http://en.wikipedia.org/wiki/CURL) >= 7 is optional if you do not fill DL_METHOD with `curl`.
* [sha1sum](https://en.wikipedia.org/wiki/Sha1sum) is optional if you do not fill HASH_TYPE with `sha1`.
* [sqlite3](http://linux.die.net/man/1/sqlite3) >= 3.4 is optional if you do not fill HASH_STORAGE with `sqlite3`.

## Installation

Execute the following commands to download the script :
```console
$ mkdir -p /etc/ftp-sync/ /var/log/ftp-sync/
$ cd /etc/init.d/
$ wget https://raw.github.com/crazy-max/ftp-sync/master/ftp-sync.sh -O ftp-sync --no-check-certificate
$ chmod +x ftp-sync
$ wget https://raw.github.com/crazy-max/ftp-sync/master/ftp-sync.conf -O /etc/ftp-sync/ftp-sync.conf --no-check-certificate
```

If you change the location of the config file, do not forget to change the path in the ftp-sync script file for the CONFIG_FILE var (default /etc/ftp-sync/ftp-sync.conf).

## Configuration

Before running the script, you must change some vars in the config file `/etc/ftp-sync/ftp-sync.conf` :

#### LOGS\_DIR (required ; default /var/log/ftp-sync)

Path to save ftp-sync logs.<br />
Example: `LOGS_DIR="/var/log/ftp-sync"`

#### PID\_FILE (required ; default /var/run/ftp-sync.pid)

Path to the file containing the current PID of the process.<br />
Example: `PID_FILE="/var/run/ftp-sync.pid"`

#### EMAIL\_LOG (optional)

Mail address where the logs are sent. Leave empty to disable sending mail.<br />
Example: `EMAIL_LOG="foo@foo.com"`

#### DEBUG (default 0)

Enable debug.<br />
Example: `DEBUG=1`

#### FTP\_HOST (required)

FTP host IP or domain.<br />
Example: `FTP_PORT="198.51.100.0"` or `FTP_PORT="ftp.foo.com"`

#### FTP\_PORT (required)

FTP port.<br />
Example: `FTP_PORT=21`

#### FTP\_USER (required)

FTP username.

#### FTP\_PASSWORD (required)

FTP password.

#### FTP\_SOURCES (required)

FTP sources paths to synchronize.<br />

Example for one path:
```
FTP_SOURCES="/downloads/"
```

Example for multi paths:
```
FTP_SOURCES="\
  /downloads/;\
  /other_path/;\
  /yet_another_path/;\
"
```

#### FTP\_SECURE (default 0)

Open a secure FTP connection (SSL/TLS). Only available for curl method.<br />
Example: `FTP_SECURE=1`

#### FTP\_CHECK\_CERT (default 0)

Check the server certificate against the available certificate authorities.<br />
Not used if `FTP_SECURE=0`.<br />
Example: `FTP_CHECK_CERT=1`

#### DL\_METHOD (default wget)

The download method. Can be `wget` or `curl`.<br />
Example: `DL_METHOD="wget"`

#### DL\_USER (optional)

Linux owner user of downloaded files.<br />
Example: `DL_USER="ftpuser"`

#### DL\_GROUP (optional)

Linux owner group of downloaded files.<br />
Example: `DL_GROUP="ftpgroup"`

#### DL\_CHMOD (optional)

Permissions of downloaded files.<br />
Example: `DL_CHMOD="644"`

#### DL\_REGEX (optional)

Apply a filter to search for files with a regular expression.<br />
Separate each regular expression with a semicolon.<br />
Leave empty to grab all files.<br />

Example for one regex:
```
DL_REGEX="Game.Of.Thrones.*.avi"
```

Example for multi regex:
```
DL_REGEX="\
  Game.Of.Thrones.*.avi;\
  Burn.Notice.*.avi;\
  The.Big.Bang.Theory.*VOSTFR.*720p.*WEB-DL.*.mkv;\
"
```

#### DL\_RETRY (default 3)

Number of retries in case of download failure.<br />
Example: `DL_RETRY=3`

#### DL\_RESUME (default 0)

Resume partially downloaded file.<br />
Example: `DL_RESUME=0`

#### DL\_HIDE\_SKIPPED (default 0)

Not display the downloads already made or valid in logs.<br />
Example: `DL_HIDE_SKIPPED=0`

#### DL\_HIDE\_PROGRESS (default 1)

Not display the progress dots during downloads.<br />
Example: `DL_HIDE_PROGRESS=1`

#### DL\_CREATE\_BASEDIR (default 0)

Create basename of a ftp source path in the destination folder.<br />
WARNING: Highly recommended if you have multiple ftp source paths to prevent overwriting!<br />
Does not work if `FTP_SOURCES="/"`.<br />

Example if `DL_CREATE_BASEDIR=1` and :
* The destination folder is `/tmp/seedbox/`
* `FTP_SOURCES="/downloads/;/other_path/"`
* `/downloads/` src path contains a file called `dl_file1`
* `/other_path/` src path contains a file called `other_file2`
<br />The destination structure will be :
```
[-] tmp
 | [-] seedbox
 |  | [-] downloads
 |  |     | dl_file1
 |  | [-] other_path
 |  |     | other_file2
 ```
 
Example if `DL_CREATE_BASEDIR=0` and :
* The destination folder is `/tmp/seedbox/`
* `FTP_SOURCES="/downloads/;/other_path/"`
* `/downloads/` src path contains a file called `dl_file1`
* `/other_path/` src path contains a file called `other_file2`
<br />The destination structure will be :
```
[-] tmp
 | [-] seedbox
 |  |  | dl_file1
 |  |  | other_file2
 ```

#### HASH\_ENABLED (default 1)

Enable audit file already downloaded.<br />
Example: `HASH_ENABLED=1`

#### HASH\_TYPE (default md5)

The hash type. Can be `md5` or `sha1`.<br />
For the `sha1` method, your need to install the required package: `apt-get install sha1sum`.<br />
Example: `HASH_TYPE="md5"`

#### HASH\_STORAGE (default text)

The hash storage process. Can be `text` or `sqlite3`.<br />
For the `sqlite3` method, your need to install the required package: `apt-get install sqlite3`.<br />
Example: `HASH_STORAGE="text"`

#### HASH\_DIR (default /etc/ftp-sync)

Path where hash checksums are stored.<br />
Example: `HASH_DIR="/etc/ftp-sync"`

## Usage

`$ /etc/init.d/ftp-sync <DIR_DEST>`

DIR_DEST is the directory where the files will be downloaded.
e.g. `$ /etc/init.d/ftp-sync /tmp/seedbox/`

## Automatic sync with cron

You can automatically synchronize FTP files by calling the script in a [crontab](http://en.wikipedia.org/wiki/Crontab).<br />
For example :

    0 4 * * * cd /etc/init.d/ && ./ftp-sync /tmp/seedbox/ >/dev/null 2>&1
	
This will synchronize your FTP files with the directory `/tmp/seedbox/` every day at 4 am.

## Logs

Each time the script is executed, a log file is created.<br />
Here is an example :

```console
FTP Sync v3.0 (2016/03/20 12:09:30)
--------------
Checking connection to ftp://198.51.100.0:21/complete/...
Successfully connected!
--------------
Script PID: 19383
Source: ftp://198.51.100.0:21/complete/
Destination: /tmp/seedbox/
Log file: /var/log/ftp-sync/20160320120930.log
FTP secure: 1
Download method: curl
Resume downloads: 1
Hash type: md5
Hash storage: sqlite3
Hash file: /etc/ftp-sync/ftp-sync.db
--------------
Finding files...
Regex: ^.*$
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

## Troubleshooting

### awk: line 1: syntax error at or near

If you have this kind of error with awk, enter this command to check your version of awk :
```console
$ awk -W version
GNU Awk 3.1.7
...
```

If you don't have GNU Awk (gawk), install it :
```console
$ apt-get install gawk
```

If you already have gawk installed on your system, check the location of awk and make a symbolic link to gawk :
```console
$ which awk
/usr/bin/awk
$ mv /usr/bin/awk /usr/bin/awk_
$ chmod -x /usr/bin/awk
$ which gawk
/usr/bin/gawk
$ ln -s /usr/bin/gawk /usr/bin/awk
```

### Synology Network Attached Storage

For Synology NAS, additional commands must be performed.

#### bootstrap, ipkg

First you must [install bootstrap, ipkg following the wiki of the official website](http://forum.synology.com/wiki/index.php/Overview_on_modifying_the_Synology_Server,_bootstrap,_ipkg_etc#How_to_install_ipkg).<br />
Next you can test ipkg and upgrade the repository.

```console
$ ipkg
$ ipkg update
$ ipkg upgrade
```

#### bash

The default shell installed on the Synology NAS is "ASH" and here we need [bash](http://en.wikipedia.org/wiki/Bash_%28Unix_shell%29).

```console
$ ipkg update
$ ipkg install bash
```

Now you have to create a symbolic link.

```console
$ ln -s /opt/bin/bash /usr/syno/bin/bash
```

#### coreutils

[coreutils](http://en.wikipedia.org/wiki/GNU_Core_Utilities) is a package containing many of the basic tools necessary for the script.

```console
$ ipkg update
$ ipkg install coreutils
```

Now you have to create a symbolic link to md5sum :

```console
$ ln -s /opt/bin/coreutils-md5sum /usr/syno/bin/md5sum
```

If you want to use sha1sum :

```console
$ ln -s /opt/bin/coreutils-sha1sum /usr/syno/bin/sha1sum
```

#### sqlite

This operation is optional if you have at least SQLite >= 3.8.<br />
You can check the current version by typing `sqlite3 --version`.

* SQLite version on Synology DSM 6 (build 7135) is **3.8.10.2** (`/usr/syno/bin/sqlite3`).
* SQLite version via ipkg is **3.8.1** (`/opt/bin/sqlite3`).

```console
$ ipkg update
$ ipkg install sqlite
```

Now you have to create a symbolic link to md5sum.

```console
$ ln -s /opt/bin/sqlite3 /usr/syno/bin/sqlite3
```

#### nail

nail is a command line email client. This means it can send emails via an email server, you need to have an email server for nail to use, e.g. could be your own hosted email server, or any email account such as yahoo, gmail, and millions of others.

```console
$ ipkg update
$ ipkg install nail
```

Here is an example to configure it with your gmail account.<br />
Open the nail config `/opt/etc/nail.rc` file with your favorite editor and add/edit the following parameters.

```console
set smtp-use-starttls
set ssl-verify=ignore
set smtp=smtp://smtp.gmail.com:587
set from=address@gmail.com
set smtp-auth=login
set smtp-auth-user=address@gmail.com
set smtp-auth-password=yourpassword
```

Now for the script, you have to create a symbolic link.

```console
$ ln -s /opt/bin/nail /usr/syno/bin/mail
```

#### wget

This operation is optional if you have at least Wget >= 1.12.<br />
You can check the current version by typing `wget --version`.

* Wget version on Synology DSM 5 is **1.10.1** (`/usr/syno/bin/wget`).
* Wget version on Synology DSM 6 (build 7135) is **1.15** (`/usr/syno/bin/wget`).
* Wget version via ipkg is **1.12** (`/opt/bin/wget`).

```console
$ ipkg update
$ ipkg install wget-ssl
```

Now you have to create a symbolic link.

```console
$ mv /usr/syno/bin/wget /usr/syno/bin/wget.old
$ ln -s /opt/bin/wget /usr/syno/bin/wget
```

#### crontab

```console
$ vi /etc/crontab
```

```console
0  4  *  *  *  root    cd /etc/init.d/ && bash ftp-sync /tmp/seedbox/ >/dev/null 2>&1
```

Then update crontab :

```console
$ /usr/syno/etc.defaults/rc.d/S04crond.sh stop
$ /usr/syno/etc.defaults/rc.d/S04crond.sh start
```

OR

```console
$ synoservice -restart crond
```

## Found a bug?

Please search for existing issues first and make sure to include all relevant information. Before reporting an issue :

* Tell me your ftp-sync version (eg. 3.0).
* Tell me your operating system and platform (eg. Debian 8 64bits).
* Tell me your wget version (eg. 1.16).
* Tell me your curl version (eg. 7.38.0).
* Tell me your md5sum / sha1sum version (eg. 8.23).
* Tell me your sqlite3 version (eg. 3.8.7.1).
* Reproduce the problem with `DEBUG=1`.
* Copy the content of the log file in `/var/log/ftp-sync` on [Pastebin](http://pastebin.com/)
* Copy/paste the Pastebin link to the issue.

## License

LGPL. See ``LICENSE`` for more details.

## More infos

http://www.crazyws.fr/dev/systeme/synchroniser-votre-seedbox-avec-votre-nas-ou-votre-ordinateur-6NGGE.html
