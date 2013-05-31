#! /bin/sh
### BEGIN INIT INFO
# Provides:          ftp-sync
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: FTP files synchronization
### END INIT INFO

##################################################################################
#                                                                                #
#  FTP Sync v1.1                                                                 #
#                                                                                #
#  A shell script to synchronize files between a remote FTP server and           #
#  your local server/computer.                                                   #
#  A file containing the MD5 hash of the name of each downloaded file will       #
#  prevent re-download a file even if it is not present in the destination       #
#  directory.                                                                    #
#  You can also apply a filter to search for files with a regular expression.    #
#  Ideal for those with a seedbox or a shared seedbox...                         #
#  Tested on Debian and Ubuntu.                                                  #
#                                                                                #
#  Author: Cr@zy                                                                 #
#  Contact: http://www.crazyws.fr                                                #
#                                                                                #
#  This program is free software: you can redistribute it and/or modify it       #
#  under the terms of the GNU General Public License as published by the Free    #
#  Software Foundation, either version 3 of the License, or (at your option)     #
#  any later version.                                                            #
#                                                                                #
#  This program is distributed in the hope that it will be useful, but WITHOUT   #
#  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS #
#  FOR A PARTICULAR PURPOSE. See the GNU General Public License for more         #
#  details.                                                                      #
#                                                                                #
#  You should have received a copy of the GNU General Public License along       #
#  with this program.  If not, see http://www.gnu.org/licenses/.                 #
#                                                                                #
#  Usage: ./ftp-sync.sh DIR_DEST                                                 #
#                                                                                #
##################################################################################

# FTP
FTP_HOST="10.0.0.1"
FTP_PORT="21"
FTP_USER=""
FTP_PASSWORD=""
FTP_SRC="/"

# Download
DL_USER=""
DL_GROUP=""
DL_CHMOD=""
DL_PATTERN=""
DL_RETRY=3

# MD5
MD5_ENABLED=1
MD5_FILE="/etc/ftp-sync/ftp-sync.md5"

# Misc
DIR_LOGS="/etc/ftp-sync/logs"

# No edits necessary beyond this line

### FUNCTIONS ###

function isDownloaded() {
  local srcfile="$1"
  local srcfiletr=`echo -n "$srcfile" | sed -e "s#$DIR_SRC##g"`
  local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
  local srcsize=`ls -la "$srcfile" | awk '{ print $5}'`

  # Check skip MD5
  if [ -z "$2" ]; then local skipmd5=0; else local skipmd5=1; fi
  
  local destfile=`echo "$srcfile" | sed -e "s#$DIR_SRC#$DIR_DEST#g"`
  if [ -f "$destfile" ]
  then
    local destsize=`ls -la "$destfile" | awk '{ print $5}'`
    if [ "$srcsize" == "$destsize" ]
    then
      echo "1"
      if [ "$MD5_ACTIVATED" == "1" -a "$skipmd5" == "0" -a -z "`grep "$srchash" "$MD5_FILE"`" ]
      then
        echo "$srchash $srcfiletr" >> "$MD5_FILE"
      fi
      exit 1
    fi
    echo "2"
  elif [ "$MD5_ACTIVATED" == "1" -a "$skipmd5" == "0" ]
  then
    cat "$MD5_FILE" | while read line
    do
      md5sum=`echo -n "$line" | cut -d ' ' -f 1`
      if [ "$srchash" == "$md5sum" ]; then echo "3"; exit 1; fi
    done
  fi

  echo "0"
}

function isMd5Enabled() {
  if [ -z "$1" ]; then local skipmd5=0; else local skipmd5=1; fi
  if [ "$MD5_ENABLED" == "1" -a -f "$MD5_FILE" -a "$skipmd5" == "0" ]
  then
    echo "1"
    exit 1;
  fi
  echo "0"
}

function downloadFile() {
  local srcfile="$1"
  local srcfiletr=`echo -n "$srcfile" | sed -e "s#$DIR_SRC##g"`
  local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
  local destfile="$2"

  # Check download retry
  if [ -z "$3" ]; then local retry=0; else local retry=$3; fi

  # Create destfile path if does not exist
  local destpath="${destfile%/*}"
  if [ ! -d $destpath ]
  then
    mkdir -p "$destpath"
    changePerms "$destpath"
  fi

  # Begin download
  echo "Start download to $destfile... Please wait..."
  if [ -x `which pv` ]; then cp "$srcfile" "$destfile"; fi
  local cpstatus="$?"

  local dlstatus=`isDownloaded "$srcfile" "1"`
  if [ "$cpstatus" == "0" -a ${dlstatus:0:1} -eq 1 ]
  then
    echo "File successfully downloaded!"
    changePerms "$destfile"
    if [ "$MD5_ACTIVATED" == "1" -a -z "`grep "$srchash" "$MD5_FILE"`" ]
    then
      echo "$srchash $srcfiletr" >> "$MD5_FILE"
    fi
  else
    rm -rf "$destfile"
    if [ $retry -lt $DL_RETRY ]
    then
      retry=`expr $retry + 1`
      echo "ERROR: Download failed... retry $retry/3"
      downloadFile "$srcfile" "$destfile" "$retry"
    else
      echo "ERROR: Download failed and too many retries..."
    fi
  fi
}

function mountFtp() {
  local mountpoint="/mnt/ftp-sync"
  if grep -qs "$mountpoint" /proc/mounts; then umountFtp; fi
  if [ ! -d $mountpoint ]; then mkdir -p $mountpoint; fi

  echo "Connecting to $FTP_HOST:$FTP_PORT..."
  local status=$(curlftpfs "$FTP_USER:$FTP_PASSWORD@$FTP_HOST:$FTP_PORT" "$mountpoint" -o nonempty 2>&1)
  if [ "$status" != "" ]; then echo "ERROR: $status"; exit 1; fi

  DIR_SRC="$mountpoint$FTP_SRC"
}

function umountFtp() {
  local mountpoint="/mnt/ftp-sync"
  umount "$mountpoint"
  rmdir "$mountpoint"
}

function process() {
  local pattern="$1"
  echo "Finding files..."
  echo "Regexp: $pattern"
  echo "--------------"
  find "$DIR_SRC" -name "$pattern" -type f | sort | while read srcfile
  do
    local starttime=$(awk 'BEGIN{srand();print srand()}')
    local srcfiletr=`echo -n "$srcfile" | sed -e "s#$DIR_SRC##g"`

    # Start process on a file
    echo "Process file : $srcfiletr"
    local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
    echo "Hash: $srchash"

    # File size
    local srcsize=`ls -lah "$srcfile" | awk '{ print $5}'`
    echo "Size: $srcsize"

    # Check validity
    local dlstatus=`isDownloaded "$srcfile"`
    if [ ${dlstatus:0:1} -eq 0 ]
    then
      echo "Status : Never downloaded..."
    elif [ ${dlstatus:0:1} -eq 1 ]
    then
      echo "Status : Already downloaded and valid. Skip download..."
    elif [ ${dlstatus:0:1} -eq 2 ]
    then
      echo "Status : Exists but sizes are different..."
    elif [ ${dlstatus:0:1} -eq 3 ]
    then
      echo "Status : MD5 sum exists. Skip download..."
    fi

    if [ ${dlstatus:0:1} -ne 1 -a ${dlstatus:0:1} -ne 3 ]
    then
      local destfile=`echo "$srcfile" | sed -e "s#$DIR_SRC#$DIR_DEST#g"`
      downloadFile "$srcfile" "$destfile"
    fi

    # Time spent
    local endtime=$(awk 'BEGIN{srand();print srand()}')
    echo "Time spent: `formatSeconds $(($endtime - $starttime))`"
    echo "--------------"
  done
}

function changePerms() {
  local path="$1"
  if [ "$DL_USER" != "" ]; then chown $DL_USER:$DL_GROUP "$path"; fi
  if [ "$DL_CHMOD" != "" ]; then chmod $DL_CHMOD "$path"; fi
}

function formatSeconds() {
  local s=${1}
  ((h=s/3600))
  ((m=s%3600/60))
  ((s=s%60))
  if [ "${#h}" == 1 ]; then h="0"$h; fi
  if [ "${#m}" == 1 ]; then m="0"$m; fi
  if [ "${#s}" == 1 ]; then s="0"$s; fi
  echo "$h:$m:$s"
}

function rebuildPath() {
  local path="$1"
  local len=${#path}-1
  if [ "${path:len}" != "/" ]; then path="$path/"; fi
  if [ "${path:0:1}" != "/" ]; then path="/$path"; fi
  echo "$path"
}

function watchTail() {
  cur_pid=$$
  tail_args=`echo "tail -f $LOG" | cut -c1-79`
  pid=`ps -e -o pid,ppid,args | grep ${cur_pid} | grep "${tail_args}"| grep -v grep | nawk '{print $1}'`

  if [ "$pid" = "" ]
  then
    if [ -z "$PS1" ]; then exit 0; else return 0; fi
  fi

  ppid=2
  while [ "$ppid" != "1" ]
  do
     pids=`ps -e -o pid,ppid,args | grep "${tail_args}"| grep ${pid} | grep -v grep | nawk '{print $1"-"$2}'`
     if [ "$pids" == "" ]; then break; fi
     ppid=`echo ${pids} | nawk -F- '{print $2}'`
     if ((ppid==1))
     then
       sleep 3
       kill -9 $pid
     fi
  done
}

### BEGIN ###

# Destination folder
DIR_DEST="$1"
if [ -z "$DIR_DEST" ]
then
  echo "Usage: $0 DIR_DEST"
  exit 1
fi

# Log file
if [ ! -d "$DIR_LOGS" ]; then mkdir -p "$DIR_LOGS"; fi
LOG="$DIR_LOGS/ftp-sync-`date +%Y%m%d%H%M%S`.log"
touch "$LOG"

# Output to log file and 
exec 1>"$LOG" 2>&1

# Starting to print log file on screen
term=`tty`
if [ -z "`echo $term | grep "/dev/"`" ]
then
  term=""
  tail -f "$LOG"
else
  tail -f "$LOG">$term & 
fi

# Starting watch in background and process
watchTail &

echo "FTP Sync v1.1 (`date +"%Y/%m/%d %H:%M:%S"`)"

# Check required packages
if [ ! -x `which awk` ]; then echo "ERROR: You need awk for this script (try apt-get install awk)"; exit 1; fi
if [ ! -x `which md5sum` ]; then echo "ERROR: You need md5sum for this script (try apt-get install md5sum)"; exit 1; fi
if [ ! -x `which curlftpfs` ]; then echo "ERROR: You need curlftpfs for this script (try apt-get install curlftpfs)"; exit 1; fi

# Mount FTP
mountFtp
if [ "$?" == "1" ]; then exit 1; fi

# Check directories
if [ ! -d "$DIR_SRC" ]; then echo "ERROR: $DIR_SRC is not a directory"; exit 1; else DIR_SRC=`rebuildPath "$DIR_SRC"`; fi
if [ ! -d "$DIR_DEST" ]; then mkdir -p "$DIR_DEST"; fi; DIR_DEST=`rebuildPath "$DIR_DEST"`

# Check MD5 file
if [ "$MD5_ENABLED" == "1" -a ! -z "$MD5_FILE" ]
then
  md5filepath="${MD5_FILE%/*}"
  if [ ! -d "$md5filepath" ]; then mkdir -p "$md5filepath"; fi
  if [ ! -f "$MD5_FILE" ]; then touch "$MD5_FILE"; fi
fi
if [ "$MD5_ENABLED" == "1" -a -f "$MD5_FILE" ]; then MD5_ACTIVATED=1; else MD5_ACTIVATED=0; fi

echo "Script PID: $$"
echo "Source: ftp://$FTP_HOST:$FTP_PORT$FTP_SRC"
echo "Destination: $DIR_DEST"
echo "Log file: $LOG"

if [ "$MD5_ACTIVATED" == "1" ]; then echo "MD5 file: $MD5_FILE"; fi
echo "--------------"

# Start process
starttime=$(awk 'BEGIN{srand();print srand()}')

if [ -z "$DL_PATTERN" ]; then DL_PATTERN="*;"; fi
IFS=';' read -ra PATTERN <<< "$DL_PATTERN"
for p in "${PATTERN[@]}"; do
  process "$p"
done

echo "Finished..."
endtime=$(awk 'BEGIN{srand();print srand()}')
echo "Total time spent: `formatSeconds $(($endtime - $starttime))`"

# Umount FTP
umountFtp

exit 0