# FTP Sync

A shell script to synchronize files between a remote FTP server and your local server/computer.<br />
A file containing the MD5 hash of the name of each downloaded file will prevent re-download a file even if it is not present in the destination directory.<br />
You can also apply a filter to search for files with a regular expression.<br />
Because this script only need ``wget``, it is ideal for those with a seedbox or a shared seedbox to synchronize with a NAS (Synology Qnap D-Link) or a local computer...

## Requirements

Commands :

* [awk](http://en.wikipedia.org/wiki/Awk) is required.
* [nawk](http://linux.die.net/man/1/nawk) is required.
* [gawk](http://www.gnu.org/software/gawk/) is required.
* [mail](http://linux.die.net/man/1/mail) is optional if you do not fill EMAIL_LOG.
* [md5sum](http://en.wikipedia.org/wiki/Md5sum) is required.
* [wget](http://en.wikipedia.org/wiki/Wget) >= 1.12 is required.

## Installation

Execute the following commands to download the script :
```console
$ cd /etc/init.d/
$ wget https://raw.github.com/crazy-max/ftp-sync/master/ftp-sync.sh -O ftp-sync --no-check-certificate
$ chmod +x ftp-sync
$ mkdir /etc/ftp-sync/
$ wget https://raw.github.com/crazy-max/ftp-sync/master/ftp-sync.conf -O /etc/ftp-sync/ftp-sync.conf --no-check-certificate
```

Before running the script, you must change some vars in the config file ``/etc/ftp-sync/ftp-sync.conf``.

* **FTP_HOST** - FTP host IP or domain. (e.g. 10.0.0.1 or ftp.example.com)
* **FTP_PORT** - FTP port. (e.g. 21)
* **FTP_USER** - FTP username.
* **FTP_PASSWORD** - FTP password.
* **FTP_SRC** - FTP path to synchronize.
* **DL_USER** - Linux owner user of downloaded files. Optional.
* **DL_GROUP** - Linux owner group of downloaded files. Optional.
* **DL_CHMOD** - Permissions of downloaded files. Optional. (e.g. 644)
* **DL_REGEX** - Apply a filter to search for files with a regular expression. Separate each regular expression with a semicolon. Leave empty to grab all files. Optional. For example: `Game.Of.Thrones.*.avi;Burn.Notice.*.avi;The.Big.Bang.Theory.*VOSTFR.*720p.*WEB-DL.*.mkv`
* **DL_RETRY** - Number of retries in case of failure of download. (default 3)
* **DL_METHOD** - The download method. Can be wget or curl. (default wget)
* **DL_HIDE_SKIPPED** - Not display the downloads already made or valid in logs. (default 0)
* **DL_HIDE_PROGRESS** - Not display the progress dots during downloads. (default 1)
* **MD5_ENABLED** - Enable audit file already downloaded.
* **MD5_FILE** - The audit file containing the hash of each downloaded file (default /etc/ftp-sync/ftp-sync.md5).
* **DIR_LOGS** - Path to save ftp-sync logs. (default /etc/ftp-sync/logs)
* **EMAIL_LOG** - Mail address where the logs are sent. Leave empty to disable sending mail.
* **PID_FILE** - Path to the file containing the current PID of the process.

If you change the location of the config file, do not forget to change the path in the ftp-sync script file for the CONFIG_FILE var (default /etc/ftp-sync/ftp-sync.conf).

## Usage

``$ /etc/init.d/ftp-sync <DIR_DEST>``

DIR_DEST is the directory where the files will be downloaded.
e.g. ``$ /etc/init.d/ftp-sync /tmp/seedbox/``

## Automatic sync with cron

You can automatically synchronize FTP files by calling the script in a [crontab](http://en.wikipedia.org/wiki/Crontab).
For example :

    0 4 * * * cd /etc/init.d/ && ./ftp-sync /tmp/seedbox/ >/dev/null 2>&1
	
This will synchronize your FTP files with the directory ``/tmp/seedbox/`` every day at 4 am.

## Logs

Each time the script is executed, a log file is created.
Here is an example :

```console
FTP Sync v1.91 (2013/12/01 01:52:14)
--------------
Checking connection to ftp://10.0.0.1:21/complete/...
Successfully connected!
--------------
Script PID: 20164
Source: ftp://10.0.0.1:21/complete/
Destination: /tmp/seedbox/
Log file: /etc/ftp-sync/logs/ftp-sync-20131201015214.log
Download method: wget
MD5 file: /etc/ftp-sync/ftp-sync.md5
--------------
Finding files...
Regex: ^.*$
--------------
Process file : Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi
Hash: baf87b6719e9f5499627fc8691efbd3c
Size: 184.18 Mb
Status : Never downloaded...
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

The MD5 file looks like this :

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

First you must [install bootstrap, ipkg following the wiki of the official website](http://forum.synology.com/wiki/index.php/Overview_on_modifying_the_Synology_Server,_bootstrap,_ipkg_etc#How_to_install_ipkg).
Next you can test ipkg and upgrade the repository.

```console
$ ipkg
$ ipkg update
$ ipkg upgrade
```

#### coreutils

[coreutils](http://en.wikipedia.org/wiki/GNU_Core_Utilities) is a package containing many of the basic tools necessary for the script.

```console
$ ipkg update
$ ipkg install coreutils
```

#### nail

nail is a command line email client. This means it can send emails via an email server, you need to have an email server for nail to use, e.g. could be your own hosted email server, or any email account such as yahoo, gmail, and millions of others.

```console
$ ipkg update
$ ipkg install nail
```

Here is an example to configure it with your gmail account.
Open the nail config ``/opt/etc/nail.rc`` file with your favorite editor and add/edit the following parameters.

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
$ ln -s /opt/bin/nail /opt/bin/mail
```

#### wget

The current version of wget on Synology is **GNU Wget 1.10.1** (/usr/syno/bin/wget).
You have to install at least wget 1.12 via ipkg.

```console
$ ipkg update
$ ipkg remove wget-ssl
$ ipkg install wget
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
0       4       *       *       *       root    cd /etc/init.d/ && bash ftp-sync /tmp/seedbox/ >/dev/null 2>&1
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

## License

LGPL. See ``LICENSE`` for more details.

## More infos

http://www.crazyws.fr/dev/systeme/synchroniser-votre-seedbox-avec-votre-nas-ou-votre-ordinateur-6NGGE.html
