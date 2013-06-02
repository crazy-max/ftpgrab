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
#  FTP Sync v1.3                                                                 #
#                                                                                #
#  A shell script to synchronize files between a remote FTP server and           #
#  your local server/computer.                                                   #
#  A file containing the MD5 hash of the name of each downloaded file will       #
#  prevent re-download a file even if it is not present in the destination       #
#  directory.                                                                    #
#  You can also apply a filter to search for files with a regular expression.    #
#  Because this script only need wget, it is ideal for those with a seedbox      #
#  or a shared seedbox to synchronize with a NAS (Synology Qnap D-Link) or a     #
#  local computer...                                                             #
#                                                                                #
#  Copyright (C) 2013 Cr@zy <webmaster@crazyws.fr>                               #
#                                                                                #
#  FTP Sync is free software; you can redistribute it and/or modify              #
#  it under the terms of the GNU Lesser General Public License as published by   #
#  the Free Software Foundation, either version 3 of the License, or             #
#  (at your option) any later version.                                           #
#                                                                                #
#  FTP Sync is distributed in the hope that it will be useful,                   #
#  but WITHOUT ANY WARRANTY; without even the implied warranty of                #
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the                  #
#  GNU Lesser General Public License for more details.                           #
#                                                                                #
#  You should have received a copy of the GNU Lesser General Public License      #
#  along with this program. If not, see http://www.gnu.org/licenses/.            #
#                                                                                #
#  Related post: http://goo.gl/OcJFA                                             #
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
DL_REGEX=""
DL_RETRY=3
DL_HIDE_SKIPPED=0
DL_HIDE_PROGRESS=1

# MD5
MD5_ENABLED=1
MD5_FILE="/etc/ftp-sync/ftp-sync.md5"

# Misc
DIR_LOGS="/etc/ftp-sync/logs"
EMAIL_LOG=""

# No edits necessary beyond this line

### FUNCTIONS ###

function isDownloaded() {
  local srcfile="$1"
  local srcfiledec=$(urlDecode "$srcfile")
  local srcfiletr=`echo -n "$srcfiledec" | sed -e "s#$FTP_SRC##g" | cut -c1-`
  local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
  local srcsize=$(getSize "$srcfile")

  # Check skip MD5
  if [ -z "$2" ]; then local skipmd5=0; else local skipmd5=$2; fi
  
  local destfile=`echo "$srcfiledec" | sed -e "s#$FTP_SRC#$DIR_DEST#g"`
  if [ -f "$destfile" ]
  then
    local destsize=`ls -la "$destfile" | awk '{ print $5}'`
    if [ "$srcsize" == "$destsize" ]
    then
      echo "1"
      if [ "$MD5_ACTIVATED" == "1" -a "$skipmd5" == "0" -a -z "`grep "^$srchash" "$MD5_FILE"`" ]
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
  local srcfiledec=$(urlDecode "$srcfile")
  local srcfiletr=`echo -n "$srcfiledec" | sed -e "s#$FTP_SRC# #g" | cut -c1-`
  local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
  local destfile="$2"
  local hidelog="$3"

  # Check download retry
  if [ -z "$4" ]; then local retry=0; else local retry=$4; fi

  # Create destfile path if does not exist
  local destpath="${destfile%/*}"
  if [ ! -d "$destpath" ]
  then
    mkdir -p "$destpath"
    changePerms "$destpath"
  fi

  # Begin download
  if [ -z "$LOG" ]; then echo "Start download to $destfile... Please wait..."; fi
  wget --progress=dot:mega --ftp-user="$FTP_USER" --ftp-password="$FTP_PASSWORD" -O "$destfile" "ftp://$FTP_HOST:$FTP_PORT$srcfile" 2>&1 | progressFilter
  
  local errordl="$?"
  local dlstatus=`isDownloaded "$srcfile" "1"`
  if [ $errordl == 0 -a ${dlstatus:0:1} -eq 1 ]
  then
    if [ -z "$LOG" ]; then echo "File successfully downloaded!"; fi
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
      if [ -z "$LOG" ]; then echo "ERROR $errordl${dlstatus:0:1}: Download failed... retry $retry/3"; fi
      downloadFile "$srcfile" "$destfile" "$hidelog" "$retry"
    else
      if [ -z "$LOG" ]; then echo "ERROR $errordl${dlstatus:0:1}: Download failed and too many retries..."; fi
    fi
  fi
}

function findFiles() {
  local path="$1"
  local regex="$2"
  local address="ftp://$FTP_HOST:$FTP_PORT"
  local files=$(wget -q --ftp-user="$FTP_USER" --ftp-password="$FTP_PASSWORD" -O - "$address$path" | grep -o 'ftp:[^"]*')
  while read -r line
  do
    local file=$(echo "$line" | sed "s#&\#32;#%20#g" | sed "s#$address# #g" | cut -c2-)
    local filedec=$(urlDecode "$file")
    local filetr=`echo -n "$filedec" | sed -e "s#$FTP_SRC# #g" | cut -c2-`
    local vregex=`echo -n "$filetr" | sed -n "/$regex/p"`
    if [ "${file#${file%?}}" == "/" ]
    then
      findFiles "$file" "$regex"
    elif [ ! -z "$vregex" ]
    then
      echo "$file"
    fi
  done <<< "$files"
}

function process() {
  local regex="$1"
  echo "Finding files..."
  echo "Regex: $regex"
  echo "--------------"
  findFiles "$FTP_SRC" "$regex" | sort | while read srcfile
  do
    LOG=""
    local skipdl=0
    local srcfiledec=$(urlDecode "$srcfile")
    local starttime=$(awk 'BEGIN{srand();print srand()}')
    local srcfiletr=`echo -n "$srcfiledec" | sed -e "s#$FTP_SRC##g" | cut -c1-`

    # Start process on a file
    addLog "Process file : $srcfiletr"
    local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
    addLog "Hash: $srchash"
    addLog "Size: $(getHumanSize "$srcfile")"

    # Check validity
    local dlstatus=`isDownloaded "$srcfile"`

    if [ ${dlstatus:0:1} -eq 0 ]
    then
      addLog "Status : Never downloaded..."
    elif [ ${dlstatus:0:1} -eq 1 ]
    then
      skipdl=1
      addLog "Status : Already downloaded and valid. Skip download..."
    elif [ ${dlstatus:0:1} -eq 2 ]
    then
      addLog "Status : Exists but sizes are different..."
    elif [ ${dlstatus:0:1} -eq 3 ]
    then
      skipdl=1
      addLog "Status : MD5 sum exists. Skip download..."
    fi

    # Check if download skipped and want to hide it in log file
    if [ "$skipdl" == "0" ] || [ "$DL_HIDE_SKIPPED" == "0" ]; then echo -e "$LOG"; LOG=""; fi

    if [ "$skipdl" == "0" ]
    then
      local destfile=`echo "$srcfiledec" | sed -e "s#$FTP_SRC#$DIR_DEST#g"`
      downloadFile "$srcfile" "$destfile" "$hidelog"
    fi

    # Time spent
    local endtime=$(awk 'BEGIN{srand();print srand()}')
    if [ -z "$LOG" ]; then echo "Time spent: `formatSeconds $(($endtime - $starttime))`"; fi
    if [ -z "$LOG" ]; then echo "--------------"; fi
  done
}

function progressFilter() {
  if [ "$DL_HIDE_PROGRESS" == "0" ]
  then
    local flag=2 c count cr=$'\r' nl=$'\n'
    while IFS='' read -d '' -rn 1 c
    do
      if [ $flag == 1 ]
      then
        printf '%c' "$c"
        if [[ "$c" =~ (s$) ]]
        then
          flag=0
          echo ""
        fi
      elif [ $flag != 0 ]
      then
        if [[ $c != $cr && $c != $nl ]]
        then
          count=0
        else
          ((count++))
          if ((count > 1))
          then
            flag=1
          fi
        fi
      fi
    done
  fi
}

function watchTail() {
  local cur_pid=$$
  local tail_args=`echo "tail -f $LOG_FILE" | cut -c1-79`
  local pid=`ps -e -o pid,ppid,args | grep ${cur_pid} | grep "${tail_args}"| grep -v grep | nawk '{print $1}'`

  if [ "$pid" = "" ]
  then
    if [ -z "$PS1" ]; then exit 0; else return 0; fi
  fi

  local ppid=2
  while [ "$ppid" != "1" ]
  do
     local pids=`ps -e -o pid,ppid,args | grep "${tail_args}"| grep ${pid} | grep -v grep | nawk '{print $1"-"$2}'`
     if [ "$pids" == "" ]; then break; fi
     local ppid=`echo ${pids} | nawk -F- '{print $2}'`
     if ((ppid==1))
     then
       if [ ! -z "$EMAIL_LOG" ]; then cat "$LOG_FILE" | mail -s "ftp-sync on $(hostname)" $EMAIL_LOG; fi
       sleep 3
       kill -9 $pid
     fi
  done
}

function urlDecode() {
  echo "$1" | sed -e "s/%\([0-9A-F][0-9A-F]\)/\\\\\x\1/g" | xargs -0 echo -e
}

function getSize() {
  echo $(wget -S --spider --ftp-user="$FTP_USER" --ftp-password="$FTP_PASSWORD" -O - "ftp://$FTP_HOST:$FTP_PORT$1" >&1 2>&1 | grep '^213' | awk '{print $2}')
}

function getHumanSize() {
  echo $(getSize "$1") | awk '{ sum=$1 ; hum[1024**3]="Gb";hum[1024**2]="Mb";hum[1024]="Kb"; for (x=1024**3; x>=1024; x/=1024){ if (sum>=x) { printf "%.2f %s\n",sum/x,hum[x];break } }}'
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

function addLog() {
  local text="$1"
  if [ ! -z "$LOG" ]; then LOG=$LOG"\n"; fi
  LOG=$LOG"$text"
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
LOG_FILE="$DIR_LOGS/ftp-sync-`date +%Y%m%d%H%M%S`.log"
touch "$LOG_FILE"

# Output to log file and 
exec 1>"$LOG_FILE" 2>&1

# Starting to print log file on screen
term=`tty`
if [ -z "`echo $term | grep "/dev/"`" ]
then
  term=""
  tail -f "$LOG_FILE"
else
  tail -f "$LOG_FILE">$term & 
fi

# Starting watch in background and process
watchTail &

echo "FTP Sync v1.3 (`date +"%Y/%m/%d %H:%M:%S"`)"

# Check required packages
if [ ! -x `which awk` ]; then echo "ERROR: You need awk for this script (try apt-get install awk)"; exit 1; fi
if [ ! -x `which md5sum` ]; then echo "ERROR: You need md5sum for this script (try apt-get install md5sum)"; exit 1; fi
if [ ! -x `which nawk` ]; then echo "ERROR: You need nawk for this script (try apt-get install nawk)"; exit 1; fi
if [ ! -x `which wget` ]; then echo "ERROR: You need wget for this script (try apt-get install wget)"; exit 1; fi

# Check directories
FTP_SRC=`rebuildPath "$FTP_SRC"`
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
echo "Log file: $LOG_FILE"

if [ "$MD5_ACTIVATED" == "1" ]; then echo "MD5 file: $MD5_FILE"; fi
echo "--------------"

# Start process
starttime=$(awk 'BEGIN{srand();print srand()}')

if [ -z "$DL_REGEX" ]; then DL_REGEX="^.*$;"; fi
IFS=';' read -ra REGEX <<< "$DL_REGEX"
for p in "${REGEX[@]}"; do
  process "$p"
done

echo "Finished..."
endtime=$(awk 'BEGIN{srand();print srand()}')
echo "Total time spent: `formatSeconds $(($endtime - $starttime))`"

exit 0