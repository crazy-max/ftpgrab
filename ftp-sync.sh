#! /bin/bash
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
#  FTP Sync v1.93                                                                #
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
#  Copyright (C) 2013-2014 Cr@zy <webmaster@crazyws.fr>                          #
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

CONFIG_FILE="/etc/ftp-sync/ftp-sync.conf"

# No edits necessary beyond this line

### FUNCTIONS ###

function ftpsyncIsDownloaded() {
  local srcfile="$1"
  local srcfiledec=$(ftpsyncUrlDecode "$srcfile")
  local srcfiletr=`echo -n "$srcfiledec" | sed -e "s#$FTP_SRC##g" | cut -c1-`
  local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
  local srcsize=$(ftpsyncGetSize "$srcfile")

  # Check skip MD5
  if [ -z "$2" ]; then local skipmd5=0; else local skipmd5=$2; fi
  
  local destfile=`echo "$srcfiledec" | sed -e "s#$FTP_SRC#$DIR_DEST#g"`
  if [ -f "$destfile" ]
  then
    local destsize=`ls -la "$destfile" | awk '{print $5}'`
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

function ftpsyncDownloadFile() {
  local srcfile="$1"
  local srcfiledec=$(ftpsyncUrlDecode "$srcfile")
  local srcfiletr=`echo -n "$srcfiledec" | sed -e "s#$FTP_SRC# #g" | cut -c1-`
  local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
  local destfile="$2"
  local hidelog="$3"
  local dlstatusfile="/tmp/ftpsync-$srchash.log"

  # Check download retry
  if [ -z "$4" ]; then local retry=0; else local retry=$4; fi

  # Create destfile path if does not exist
  local destpath="${destfile%/*}"
  if [ ! -d "$destpath" ]
  then
    mkdir -p "$destpath"
    ftpsyncChangePerms "$destpath"
  fi

  # Begin download
  if [ -z "$LOG" ]; then ftpsyncEcho "Start download to $destfile... Please wait..."; fi
  if [ -f "$dlstatusfile" ]; then rm "$dlstatusfile"; fi
  if [ "$DL_METHOD" == "curl" ]
  then
    curl --stderr "$dlstatusfile" --globoff -u "$FTP_USER:$FTP_PASSWORD" "ftp://$FTP_HOST:$FTP_PORT$srcfile" -o "$destfile"
    local errordl="$?"
    if [ -z "$LOG" -a "$DL_HIDE_PROGRESS" == "0" -a -f "$dlstatusfile" -a -s "$dlstatusfile" ]
    then
      ftpsyncEcho ""
      cat "$dlstatusfile" | sed s/\\r/\\n/g | head -n -2
      cat "$dlstatusfile" | sed s/\\r/\\n/g | head -n -2 >> "$LOG_FILE"
      ftpsyncEcho ""
    fi
  else
    wget --progress=dot:mega --ftp-user="$FTP_USER" --ftp-password="$FTP_PASSWORD" -O "$destfile" -a "$dlstatusfile" "ftp://$FTP_HOST:$FTP_PORT$srcfile"
    local errordl="$?"
    if [ -z "$LOG" -a "$DL_HIDE_PROGRESS" == "0" -a -f "$dlstatusfile" -a -s "$dlstatusfile" ]
    then
      ftpsyncEcho ""
      cat "$dlstatusfile" | sed s/\\r/\\n/g | sed '/\.\.\.\.\.\.\.\. /!d'
      cat "$dlstatusfile" | sed s/\\r/\\n/g | sed '/\.\.\.\.\.\.\.\. /!d' >> "$LOG_FILE"
      ftpsyncEcho ""
    fi
  fi
  if [ -f "$dlstatusfile" ]; then rm "$dlstatusfile"; fi
  
  local dlstatus=`ftpsyncIsDownloaded "$srcfile" "1"`
  if [ $errordl == 0 -a ${dlstatus:0:1} -eq 1 ]
  then
    if [ -z "$LOG" ]; then ftpsyncEcho "File successfully downloaded!"; fi
    ftpsyncChangePerms "$destfile"
    if [ "$MD5_ACTIVATED" == "1" -a -z "`grep "$srchash" "$MD5_FILE"`" ]
    then
      echo "$srchash $srcfiletr" >> "$MD5_FILE"
    fi
  else
    rm -rf "$destfile"
    if [ $retry -lt $DL_RETRY ]
    then
      retry=`expr $retry + 1`
      if [ -z "$LOG" ]; then ftpsyncEcho "ERROR $errordl${dlstatus:0:1}: Download failed... retry $retry/3"; fi
      ftpsyncDownloadFile "$srcfile" "$destfile" "$hidelog" "$retry"
    else
      if [ -z "$LOG" ]; then ftpsyncEcho "ERROR $errordl${dlstatus:0:1}: Download failed and too many retries..."; fi
    fi
  fi
}

function ftpsyncFindFiles() {
  local path="$1"
  local regex="$2"
  local address="ftp://$FTP_HOST:$FTP_PORT"
  local files=$(wget -q --ftp-user="$FTP_USER" --ftp-password="$FTP_PASSWORD" -O - "$address$path" | grep -o 'ftp:[^"]*')
  while read -r line
  do
    local lineClean=$(echo "$line" | sed "s#&\#32;#%20#g" | sed "s#$address# #g" | cut -c2-)
    local basename=$(basename "$lineClean")
    local file="$path$basename"
    local filedec=$(ftpsyncUrlDecode "$file")
    local filetr=`echo -n "$filedec" | sed -e "s#$FTP_SRC# #g" | cut -c2-`
    local vregex=`echo -n "$filetr" | sed -n "/$regex/p"`
    if [ "${lineClean#${lineClean%?}}" == "/" ]
    then
      ftpsyncFindFiles "$file/" "$regex"
    elif [ ! -z "$vregex" ]
    then
      echo "$file"
    fi
  done <<< "$files"
}

function ftpsyncProcess() {
  local regex="$1"
  ftpsyncEcho "Finding files..."
  ftpsyncEcho "Regex: $regex"
  ftpsyncEcho "--------------"
  ftpsyncFindFiles "$FTP_SRC" "$regex" | sort | while read srcfile
  do
    LOG=""
    local skipdl=0
    local srcfiledec=$(ftpsyncUrlDecode "$srcfile")
    local starttime=$(awk 'BEGIN{srand();print srand()}')
    local srcfiletr=`echo -n "$srcfiledec" | sed -e "s#$FTP_SRC##g" | cut -c1-`
    local destfile=`echo "$srcfiledec" | sed -e "s#$FTP_SRC#$DIR_DEST#g"`

    if [ ${destfile:${#destfile} - 1} == "/" ]
    then
      mkdir -p "$destfile"
    else
      # Start process on a file
      ftpsyncAddLog "Process file : $srcfiletr"
      local srchash=`echo -n "$srcfiletr" | md5sum - | cut -d ' ' -f 1`
      ftpsyncAddLog "Hash: $srchash"
      ftpsyncAddLog "Size: $(ftpsyncGetHumanSize "$srcfile")"

      # Check validity
      local dlstatus=`ftpsyncIsDownloaded "$srcfile"`

      if [ ${dlstatus:0:1} -eq 0 ]
      then
        ftpsyncAddLog "Status : Never downloaded..."
      elif [ ${dlstatus:0:1} -eq 1 ]
      then
        skipdl=1
        ftpsyncAddLog "Status : Already downloaded and valid. Skip download..."
      elif [ ${dlstatus:0:1} -eq 2 ]
      then
        ftpsyncAddLog "Status : Exists but sizes are different..."
      elif [ ${dlstatus:0:1} -eq 3 ]
      then
        skipdl=1
        ftpsyncAddLog "Status : MD5 sum exists. Skip download..."
      fi

      # Check if download skipped and want to hide it in log file
      if [ "$skipdl" == "0" ] || [ "$DL_HIDE_SKIPPED" == "0" ]; then ftpsyncEcho "$LOG"; LOG=""; fi

      if [ "$skipdl" == "0" ]
      then
        ftpsyncDownloadFile "$srcfile" "$destfile" "$hidelog"
      fi

      # Time spent
      local endtime=$(awk 'BEGIN{srand();print srand()}')
      if [ -z "$LOG" ]; then ftpsyncEcho "Time spent: `ftpsyncFormatSeconds $(($endtime - $starttime))`"; fi
      if [ -z "$LOG" ]; then ftpsyncEcho "--------------"; fi
    fi
  done
}

function ftpsyncKill() {
  local cpid="$1"
  pids="$cpid"
  if [ -d "/proc/$cpid" -a -f "/proc/$cpid/cmdline" ]
  then
    local cmdline=`cat "/proc/$cpid/cmdline"`
    kill -9 $cpid
    sleep 2
    local oPidsFile=`find /proc -type f -name "cmdline" | grep '/proc/[1-9][0-9]*/cmdline'`
    echo "$oPidsFile" | sort | while read oPidFile
    do
      if [ -f "$oPidFile" ]
      then
        local oCmdLine=`cat "$oPidFile" 2>/dev/null`
        if [ "$cmdline" == "$oCmdLine" ]
        then
          local oPid=$(echo "$oPidFile" | cut -d '/' -f 3)
          if [ $oPid != $$ ]; then kill -9 $oPid 2>/dev/null; fi
        fi
      fi
    done
  fi
}

function ftpsyncUrlDecode() {
  echo "$1" | sed -e "s/%\([0-9A-F][0-9A-F]\)/\\\\\x\1/g" | xargs -0 echo -e
}

function ftpsyncGetSize() {
  echo $(wget -S --spider --ftp-user="$FTP_USER" --ftp-password="$FTP_PASSWORD" -O - "ftp://$FTP_HOST:$FTP_PORT$1" >&1 2>&1 | grep '^213' | awk '{print $2}')
}

function ftpsyncGetHumanSize() {
  echo $(ftpsyncGetSize "$1") | awk '{ sum=$1 ; hum[1024**3]="Gb";hum[1024**2]="Mb";hum[1024]="Kb"; for (x=1024**3; x>=1024; x/=1024){ if (sum>=x) { printf "%.2f %s\n",sum/x,hum[x];break } }}'
}

function ftpsyncChangePerms() {
  local path="$1"
  if [ "$DL_USER" != "" ]; then chown $DL_USER:$DL_GROUP "$path"; fi
  if [ "$DL_CHMOD" != "" ]; then chmod $DL_CHMOD "$path"; fi
}

function ftpsyncFormatSeconds() {
  local s=${1}
  ((h=s/3600))
  ((m=s%3600/60))
  ((s=s%60))
  if [ "${#h}" == 1 ]; then h="0"$h; fi
  if [ "${#m}" == 1 ]; then m="0"$m; fi
  if [ "${#s}" == 1 ]; then s="0"$s; fi
  echo "$h:$m:$s"
}

function ftpsyncRebuildPath() {
  local path="$1"
  local len=${#path}-1
  if [ "${path:len}" != "/" ]; then path="$path/"; fi
  if [ "${path:0:1}" != "/" ]; then path="/$path"; fi
  echo "$path"
}

function ftpsyncAddLog() {
  local text="$1"
  if [ ! -z "$LOG" ]; then LOG=$LOG"\n"; fi
  LOG=$LOG"$text"
}

function ftpsyncEcho() {
  echo -e "$1" | tee -a "$LOG_FILE"
}

### BEGIN ###

SCRIPT_NAME=$(basename "$0")

# Read config file
if [ ! -f "$CONFIG_FILE" ]
then
  echo "ERROR: Config file $CONFIG_FILE not found..."
  exit 1
else
  source "$CONFIG_FILE"
fi

# Destination folder
DIR_DEST="$1"
if [ -z "$DIR_DEST" ]
then
  echo "Usage: ./$0 DIR_DEST"
  exit 1
fi

# Check download method
FTP_SRC=`ftpsyncRebuildPath "$FTP_SRC"`
if [ -z "$DL_METHOD" ] || [ "$DL_METHOD" != "wget" -a "$DL_METHOD" != "curl" ]
then
  DL_METHOD="wget"
fi;

# Log file
if [ ! -d "$DIR_LOGS" ]; then mkdir -p "$DIR_LOGS"; fi
LOG_FILE="$DIR_LOGS/ftp-sync-`date +%Y%m%d%H%M%S`.log"
touch "$LOG_FILE"

ftpsyncEcho "FTP Sync v1.93 (`date +"%Y/%m/%d %H:%M:%S"`)"
ftpsyncEcho "--------------"

# Check required packages
if [ ! -x `which awk` ]; then ftpsyncEcho "ERROR: You need awk for this script (try apt-get install awk)"; exit 1; fi
if [ ! -x `which nawk` ]; then ftpsyncEcho "ERROR: You need nawk for this script (try apt-get install nawk)"; exit 1; fi
if [ ! -x `which gawk` ]; then ftpsyncEcho "ERROR: You need nawk for this script (try apt-get install gawk)"; exit 1; fi
if [ ! -x `which md5sum` ]; then ftpsyncEcho "ERROR: You need md5sum for this script (try apt-get install md5sum)"; exit 1; fi
if [ ! -x `which wget` ]; then ftpsyncEcho "ERROR: You need wget for this script (try apt-get install wget)"; exit 1; fi

# Check directories
FTP_SRC=`ftpsyncRebuildPath "$FTP_SRC"`
if [ ! -d "$DIR_DEST" ]; then mkdir -p "$DIR_DEST"; fi; DIR_DEST=`ftpsyncRebuildPath "$DIR_DEST"`

# Check MD5 file
if [ "$MD5_ENABLED" == "1" -a ! -z "$MD5_FILE" ]
then
  md5filepath="${MD5_FILE%/*}"
  if [ ! -d "$md5filepath" ]; then mkdir -p "$md5filepath"; fi
  if [ ! -f "$MD5_FILE" ]; then touch "$MD5_FILE"; fi
fi
if [ "$MD5_ENABLED" == "1" -a -f "$MD5_FILE" ]; then MD5_ACTIVATED=1; else MD5_ACTIVATED=0; fi

# Check ftpsyncProcess already running
currentPid=$$
if [ -f "$PID_FILE" ]
then
  oldPid=`cat "$PID_FILE"`
  if [ -d "/proc/$oldPid" ]
  then
    ftpsyncEcho "ERROR: ftp-sync already running..."
    read -t 10 -p "Do you want to kill the current process? [Y/n] : " choice
    choice=${choice:-timeout}
    echo -n "Do you want to kill the current process? [Y/n] : $choice" >> "$LOG_FILE"
    case "$choice" in
      y|Y)
        ftpsyncKill "$oldPid";;
      n|N)
        exit 1;;
      timeout)
        echo "n"
        exit 1;;
    esac
    ftpsyncEcho "--------------"
  fi
fi
echo $currentPid > "$PID_FILE"

# Check connection
ftpsyncEcho "Checking connection to ftp://$FTP_HOST:$FTP_PORT$FTP_SRC..."
wget --spider -q --tries=1 --timeout=5 --ftp-user="$FTP_USER" --ftp-password="$FTP_PASSWORD" -O - "ftp://$FTP_HOST:$FTP_PORT$FTP_SRC"
connectionExitCode="$?"

if [ $connectionExitCode != "0" ]
then
  # More infos: http://www.gnu.org/software/wget/manual/html_node/Exit-Status.html
  case "$connectionExitCode" in
    1)
      ftpsyncEcho "ERROR: Generic error code...";;
    2)
      ftpsyncEcho "ERROR: Parse error (for instance, when parsing command-line options, the '.wgetrc' or '.netrc')...";;
    3)
      ftpsyncEcho "ERROR: File I/O error...";;
    4)
      ftpsyncEcho "ERROR: Network failure...";;
    5)
      ftpsyncEcho "ERROR: SSL verification failure...";;
    6)
      ftpsyncEcho "ERROR: Username/password authentication failure...";;
    7)
      ftpsyncEcho "ERROR: Protocol errors...";;
    8)
      ftpsyncEcho "ERROR: Server issued an error response...";;
  esac
  exit 1
else
  ftpsyncEcho "Successfully connected!"
  ftpsyncEcho "--------------"
fi

ftpsyncEcho "Script PID: $currentPid"
ftpsyncEcho "Source: ftp://$FTP_HOST:$FTP_PORT$FTP_SRC"
ftpsyncEcho "Destination: $DIR_DEST"
ftpsyncEcho "Log file: $LOG_FILE"
ftpsyncEcho "Download method: $DL_METHOD"

if [ "$MD5_ACTIVATED" == "1" ]; then ftpsyncEcho "MD5 file: $MD5_FILE"; fi
ftpsyncEcho "--------------"

# Start ftpsyncProcess
starttime=$(awk 'BEGIN{srand();print srand()}')

if [ -z "$DL_REGEX" ]; then DL_REGEX="^.*$;"; fi
IFS=';' read -ra REGEX <<< "$DL_REGEX"
for p in "${REGEX[@]}"; do
  ftpsyncProcess "$p"
done

# Change perms
if [ "$DL_USER" != "" ]
then
  ftpsyncEcho "Change the ownership recursively of 'Destination' path to $DL_USER:$DL_GROUP"
  chown -R $DL_USER:$DL_GROUP "$DIR_DEST"
fi
if [ "$DL_CHMOD" != "" ]
then
  ftpsyncEcho "Change the access permissions recursively of 'Destination' path to $DL_CHMOD"
  chmod -R $DL_CHMOD "$DIR_DEST"
fi
if [ "$DL_USER" != "" ] || [ "$DL_CHMOD" != "" ]
then
  ftpsyncEcho "--------------"
fi

ftpsyncEcho "Finished..."
endtime=$(awk 'BEGIN{srand();print srand()}')
ftpsyncEcho "Total time spent: `ftpsyncFormatSeconds $(($endtime - $starttime))`"

rm "$PID_FILE"

# Send logs
if [ ! -z "$EMAIL_LOG" ]; then cat "$LOG_FILE" | mail -s "ftp-sync on $(hostname)" $EMAIL_LOG; fi

exit 0
