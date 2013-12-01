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
$ wget https://raw.github.com/crazy-max/ftp-sync/master/ftp-sync.conf -O ftp-sync.conf --no-check-certificate
$ chmod +x ftp-sync
```

Before running the script, you must change some vars in the config file ``ftp-sync.conf``.

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
* **DL_HIDE_SKIPPED** - Not display the downloads already made ​​or valid in logs. (default 0)
* **MD5_ENABLED** - Enable audit file already downloaded.
* **MD5_FILE** - The audit file containing the hash of each downloaded file (default /etc/ftp-sync/ftp-sync.md5).
* **DIR_LOGS** - Path to save ftp-sync logs. (default /etc/ftp-sync/logs)
* **EMAIL_LOG** - Mail address where the logs are sent. Leave empty to disable sending mail.
* **PID_FILE** - Path to the file containing the current PID of the process.

If you change the location of the config file, do not forget to change the path in the ftp-sync script file for the CONFIG_FILE var (default /etc/init.d/ftp-sync.conf).

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

     0K ........ ........ ........ ........ ........ ........  1% 4,93M 37s
  3072K ........ ........ ........ ........ ........ ........  3% 10,4M 27s
  6144K ........ ........ ........ ........ ........ ........  4% 10,4M 23s
  9216K ........ ........ ........ ........ ........ ........  6% 10,4M 21s
 12288K ........ ........ ........ ........ ........ ........  8% 10,6M 20s
 15360K ........ ........ ........ ........ ........ ........  9% 10,5M 19s
 18432K ........ ........ ........ ........ ........ ........ 11% 10,6M 18s
 21504K ........ ........ ........ ........ ........ ........ 13% 10,6M 17s
 24576K ........ ........ ........ ........ ........ ........ 14% 9,96M 17s
 27648K ........ ........ ........ ........ ........ ........ 16% 10,7M 16s
 30720K ........ ........ ........ ........ ........ ........ 17% 9,97M 16s
 33792K ........ ........ ........ ........ ........ ........ 19% 10,0M 16s
 36864K ........ ........ ........ ........ ........ ........ 21% 10,2M 15s
 39936K ........ ........ ........ ........ ........ ........ 22% 9,97M 15s
 43008K ........ ........ ........ ........ ........ ........ 24% 9,83M 15s
 46080K ........ ........ ........ ........ ........ ........ 26% 9,86M 14s
 49152K ........ ........ ........ ........ ........ ........ 27% 9,87M 14s
 52224K ........ ........ ........ ........ ........ ........ 29% 9,59M 14s
 55296K ........ ........ ........ ........ ........ ........ 30% 9,94M 13s
 58368K ........ ........ ........ ........ ........ ........ 32% 9,78M 13s
 61440K ........ ........ ........ ........ ........ ........ 34% 9,74M 13s
 64512K ........ ........ ........ ........ ........ ........ 35% 9,79M 12s
 67584K ........ ........ ........ ........ ........ ........ 37% 9,82M 12s
 70656K ........ ........ ........ ........ ........ ........ 39% 10,0M 12s
 73728K ........ ........ ........ ........ ........ ........ 40% 9,71M 11s
 76800K ........ ........ ........ ........ ........ ........ 42% 10,0M 11s
 79872K ........ ........ ........ ........ ........ ........ 43% 9,82M 11s
 82944K ........ ........ ........ ........ ........ ........ 45% 9,82M 10s
 86016K ........ ........ ........ ........ ........ ........ 47% 9,76M 10s
 89088K ........ ........ ........ ........ ........ ........ 48% 9,56M 10s
 92160K ........ ........ ........ ........ ........ ........ 50% 9,63M 9s
 95232K ........ ........ ........ ........ ........ ........ 52% 9,28M 9s
 98304K ........ ........ ........ ........ ........ ........ 53% 9,44M 9s
101376K ........ ........ ........ ........ ........ ........ 55% 9,50M 9s
104448K ........ ........ ........ ........ ........ ........ 57% 9,87M 8s
107520K ........ ........ ........ ........ ........ ........ 58% 9,69M 8s
110592K ........ ........ ........ ........ ........ ........ 60% 9,69M 8s
113664K ........ ........ ........ ........ ........ ........ 61% 9,65M 7s
116736K ........ ........ ........ ........ ........ ........ 63% 9,35M 7s
119808K ........ ........ ........ ........ ........ ........ 65% 9,41M 7s
122880K ........ ........ ........ ........ ........ ........ 66% 9,84M 6s
125952K ........ ........ ........ ........ ........ ........ 68% 9,55M 6s
129024K ........ ........ ........ ........ ........ ........ 70% 9,76M 6s
132096K ........ ........ ........ ........ ........ ........ 71% 9,77M 5s
135168K ........ ........ ........ ........ ........ ........ 73% 9,71M 5s
138240K ........ ........ ........ ........ ........ ........ 74% 9,44M 5s
141312K ........ ........ ........ ........ ........ ........ 76% 9,74M 4s
144384K ........ ........ ........ ........ ........ ........ 78% 9,61M 4s
147456K ........ ........ ........ ........ ........ ........ 79% 9,48M 4s
150528K ........ ........ ........ ........ ........ ........ 81% 8,73M 4s
153600K ........ ........ ........ ........ ........ ........ 83% 9,41M 3s
156672K ........ ........ ........ ........ ........ ........ 84% 9,44M 3s
159744K ........ ........ ........ ........ ........ ........ 86% 9,66M 3s
162816K ........ ........ ........ ........ ........ ........ 87% 9,54M 2s
165888K ........ ........ ........ ........ ........ ........ 89% 9,50M 2s
168960K ........ ........ ........ ........ ........ ........ 91% 9,36M 2s
172032K ........ ........ ........ ........ ........ ........ 92% 9,57M 1s
175104K ........ ........ ........ ........ ........ ........ 94% 9,38M 1s
178176K ........ ........ ........ ........ ........ ........ 96% 9,41M 1s
181248K ........ ........ ........ ........ ........ ........ 97% 9,85M 0s
184320K ........ ........ ........ ........ ........ ........ 99% 9,48M 0s
187392K ........ ........ ..                                 100% 8,19M=19s

File successfully downloaded!
Time spent: 00:00:22
--------------
Change the ownership recursively of 'Destination' path to ftpuser:ftpgroup
Change the access permissions recursively of 'Destination' path to 755
--------------
Finished...
Total time spent: 00:00:23
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
