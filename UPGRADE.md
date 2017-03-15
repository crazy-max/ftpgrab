## 3.x > 4.x

To upgrade from 3.x to 4.x you have to move some files and rename the config and hash file to a custom name like `seedbox.conf` in the below example.

```
// Move to /opt
$ mv /etc/ftp-sync /opt
$ cd /opt/ftp-sync/

// Create required folders
$ mkdir conf
$ mkdir hash

// Move files
$ mv ftp-sync.conf conf/seedbox.conf
$ mv ftp-sync.txt conf/seedbox.txt

// Download the latest script and dist config
$ wget https://raw.github.com/ftp-sync/ftp-sync/master/ftp-sync.sh -O /etc/init.d/ftp-sync --no-check-certificate
$ chmod +x /etc/init.d/ftp-sync
$ wget https://raw.github.com/ftp-sync/ftp-sync/master/ftp-sync.conf -O /opt/ftp-sync/ftp-sync.conf --no-check-certificate

// Rename log files
cd /var/log/ftp-sync/
for FILENAME in *; do mv $FILENAME seedbox-$FILENAME; done
```

Next you will have to edit your config file `/opt/ftp-sync/conf/seedbox.conf` :

* Remove lines starting with `LOGS_DIR=`, `PID_FILE=` and `HASH_DIR=`
* Add a new line before `EMAIL_LOG=` with `DIR_DEST="/tmp/seedbox"` (replace `/tmp/seedbox` to your destination folder)
* Add a new line after `DL_RESUME=` with `DL_SHUFFLE=0`

If you have a cron, do not forget to replace the argument to the config file of your choice :

```
0 4 * * * cd /etc/init.d/ && ./ftp-sync seedbox.conf >/dev/null 2>&1
```
