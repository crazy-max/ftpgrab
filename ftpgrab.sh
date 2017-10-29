#! /bin/bash

###################################################################################
#                                                                                 #
#  FTPGrab v4.2.2                                                                 #
#                                                                                 #
#  Simple script to grab your files from a remote FTP server.                     #
#                                                                                 #
#  MIT License                                                                    #
#  Copyright (c) 2013-2017 Cr@zy                                                  #
#                                                                                 #
#  Permission is hereby granted, free of charge, to any person obtaining a copy   #
#  of this software and associated documentation files (the "Software"), to deal  #
#  in the Software without restriction, including without limitation the rights   #
#  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell      #
#  copies of the Software, and to permit persons to whom the Software is          #
#  furnished to do so, subject to the following conditions:                       #
#                                                                                 #
#  The above copyright notice and this permission notice shall be included in all #
#  copies or substantial portions of the Software.                                #
#                                                                                 #
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR     #
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,       #
#  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE    #
#  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER         #
#  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,  #
#  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE  #
#  SOFTWARE.                                                                      #
#                                                                                 #
#  Related post: http://goo.gl/OcJFA                                              #
#  Usage: ./ftpgrab.sh CONFIG_FILE                                                #
#                                                                                 #
###################################################################################

BASE_DIR="/opt/ftpgrab"
CONFIG_DIR="$BASE_DIR/conf"
HASH_DIR="$BASE_DIR/hash"
LOGS_DIR="/var/log/ftpgrab"
PID_DIR="/var/run/ftpgrab"

### FUNCTIONS ###

function ftpgrabIsDownloaded() {
  local _SRC_FILE_PROC="$1"
  local _SRC_FILE="$2"
  local _SRC_FILE_SHORT=""
  local _SRC_FILE_SHORT_2=""
  local _DEST_FILE=""
  local _SKIP_HASH=""

  if [ "$DL_METHOD" == "curl" ]
  then
    local _SRC_FILE_SHORT=$(echo -n "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")##" | cut -c1-)
    local _SRC_FILE_SHORT_2=$(echo -n "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")# #" | cut -c2-)
    local _DEST_FILE=$(echo "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")#$(ftpgrabEscapeSed "$DIR_DEST_REF")#")
  else
    local _SRC_FILE_SHORT=$(echo -n "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")##" | cut -c1-)
    local _SRC_FILE_SHORT_2=$(echo -n "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")# #" | cut -c2-)
    local _DEST_FILE=$(echo "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")#$(ftpgrabEscapeSed "$DIR_DEST_REF")#")
  fi

  local _SRC_HASH=$(echo -n "$_SRC_FILE_SHORT" | ${HASH_CMD} - | cut -d ' ' -f 1)
  local _SRC_SIZE=$(ftpgrabGetSize "$_SRC_FILE_PROC")
  if [ ! -z "$3" ]; then local _SKIP_HASH=$3; fi

  if [ -f "$_DEST_FILE" ]
  then
    local _DEST_SIZE=$(ls -la "$_DEST_FILE" | awk '{print $5}')
    if [ "$_SRC_SIZE" == "$_DEST_SIZE" ]
    then
      if [ "$HASH_ACTIVATED" == "1" ] && [ "$_SKIP_HASH" == "0" ]
      then
        if [ "$HASH_STORAGE" == "text" ] && [ -z $(grep "^$_SRC_HASH" "$HASH_FILE") ]
        then
          echo "$_SRC_HASH $_SRC_FILE_SHORT" >> "$HASH_FILE"
        elif [ "$HASH_STORAGE" == "sqlite3" ] && [ $(sqlite3 "$HASH_FILE" "SELECT EXISTS(SELECT 1 FROM data WHERE hash='$_SRC_HASH' LIMIT 1)") == 0 ]
        then
          sqlite3 "$HASH_FILE" "INSERT INTO data (hash,filename) VALUES (\"$_SRC_HASH\",\"$_SRC_FILE_SHORT\")";
        fi
      fi
      echo ${FILE_STATUS_SIZE_EQUAL}
      exit 1
    fi
    echo ${FILE_STATUS_SIZE_DIFF}
    exit 1
  elif [ "$HASH_ACTIVATED" == "1" ] && [ "$_SKIP_HASH" == "0" ]
  then
    if [ "$HASH_STORAGE" == "text" ] && [ ! -z $(grep "^$_SRC_HASH" "$HASH_FILE") ]
    then
      echo ${FILE_STATUS_HASH_EXISTS}
      exit 1
    elif [ "$HASH_STORAGE" == "sqlite3" ] && [ $(sqlite3 "$HASH_FILE" "SELECT EXISTS(SELECT 1 FROM data WHERE hash='$_SRC_HASH' LIMIT 1)") == 1 ]
    then
      echo ${FILE_STATUS_HASH_EXISTS}
      exit 1
    fi
  fi

  echo ${FILE_STATUS_NEVER_DL}
  exit 1
}

function ftpgrabDownloadFile() {
  local _SRC_FILE_PROC="$1"
  local _SRC_FILE="$2"
  local _SRC_FILE_SHORT=""
  local _SRC_FILE_SHORT_2=""
  local _DEST_FILE=""
  local _RESUME_CMD=""

  if [ "$DL_METHOD" == "curl" ]
  then
    local _SRC_FILE_SHORT=`echo -n "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")##" | cut -c1-`
    local _SRC_FILE_SHORT_2=`echo -n "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")# #" | cut -c2-`
    local _DEST_FILE=`echo "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")#$(ftpgrabEscapeSed "$DIR_DEST_REF")#"`
  else
    local _SRC_FILE_SHORT=`echo -n "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")##" | cut -c1-`
    local _SRC_FILE_SHORT_2=`echo -n "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")# #" | cut -c2-`
    local _DEST_FILE=`echo "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")#$(ftpgrabEscapeSed "$DIR_DEST_REF")#"`
  fi

  local _SRC_HASH=`echo -n "$_SRC_FILE_SHORT" | ${HASH_CMD} - | cut -d ' ' -f 1`
  if [ ! -z "$3" ]; then local _DEST_FILE=$3; fi
  local _RESUME="$4"
  local _DL_RETRY_COUNT="$5"
  local _DL_STATUS_FILE="/tmp/ftpgrab-$_SRC_HASH.log"

  # Check download resume
  if [ "$_RESUME" == "1" ]
  then
    if [ "$DL_METHOD" == "curl" ]
    then
      local _RESUME_CMD=" --continue-at -"
    else
      local _RESUME_CMD=" --continue"
    fi
  fi

  # Create destfile path if does not exist
  local _DEST_PATH="${_DEST_FILE%/*}"
  if [ ! -d "$_DEST_PATH" ]
  then
    mkdir -p "$_DEST_PATH"
    ftpgrabChangePerms "$_DEST_PATH"
  fi

  # Begin download
  local _ERROR_DL=0
  if [ -z "$LOG" ]; then ftpgrabEcho "Start download to $_DEST_FILE... Please wait..."; fi
  if [ -f "$_DL_STATUS_FILE" ]; then rm "$_DL_STATUS_FILE"; fi
  if [ "$DL_METHOD" == "curl" ]
  then
    ftpgrabDebug "Download command: curl ${FTP_CURL_HIDECREDS}${_RESUME_CMD} \"ftp://${FTP_HOST}:${FTP_PORT}${_SRC_FILE_PROC}\" -o \"${_DEST_FILE}\""
    curl --stderr "$_DL_STATUS_FILE" ${FTP_CURL}${_RESUME_CMD} "ftp://${FTP_HOST}:${FTP_PORT}${_SRC_FILE_PROC}" -o "$_DEST_FILE"
    local _ERROR_DL="$?"
    if [ -z "$LOG" ] && [ ${DL_HIDE_PROGRESS} -eq 0 -a -f "$_DL_STATUS_FILE" -a -s "$_DL_STATUS_FILE" ]
    then
      ftpgrabEcho ""
      < "$_DL_STATUS_FILE" | sed '/^$/d' | head -n -2
      < "$_DL_STATUS_FILE" | sed '/^$/d' | head -n -2 >> "$LOG_FILE"
      ftpgrabEcho ""
    fi
  else
    ftpgrabDebug "Download command: wget --progress=dot:mega ${FTP_WGET_HIDECREDS}${_RESUME_CMD} -O \"${_DEST_FILE}\" \"ftp://{$FTP_HOST}:${FTP_PORT}${_SRC_FILE_PROC}\""
    wget --progress=dot:mega ${FTP_WGET}${_RESUME_CMD} -O "${_DEST_FILE}" -a "${_DL_STATUS_FILE}" "ftp://${FTP_HOST}:${FTP_PORT}${_SRC_FILE_PROC}"
    local _ERROR_DL="$?"
    if [ -z "$LOG" ] && [ ${DL_HIDE_PROGRESS} -eq 0 -a -f "$_DL_STATUS_FILE" -a -s "$_DL_STATUS_FILE" ]
    then
      ftpgrabEcho ""
      cat "$_DL_STATUS_FILE" | sed s/\\r/\\n/g | sed '/\.\.\.\.\.\.\.\. /!d'
      cat "$_DL_STATUS_FILE" | sed s/\\r/\\n/g | sed '/\.\.\.\.\.\.\.\. /!d' >> "$LOG_FILE"
      ftpgrabEcho ""
    fi
  fi
  if [ -f "$_DL_STATUS_FILE" ]; then rm "$_DL_STATUS_FILE"; fi

  local _DL_STATUS=$(ftpgrabIsDownloaded "$_SRC_FILE_PROC" "$_SRC_FILE" "1")
  if [ ${_ERROR_DL} -eq 0 -a ${_DL_STATUS:0:1} -eq ${FILE_STATUS_SIZE_EQUAL} ]
  then
    if [ -z "$LOG" ]; then ftpgrabEcho "File successfully downloaded!"; fi
    ftpgrabChangePerms "$_DEST_FILE"
    if [ "$HASH_ACTIVATED" == "1" ]
    then
      if [ "$HASH_STORAGE" == "text" ] && [ -z "`grep "^$_SRC_HASH" "$HASH_FILE"`" ]
      then
        echo "$_SRC_HASH $_SRC_FILE_SHORT" >> "$HASH_FILE"
      elif [ "$HASH_STORAGE" == "sqlite3" ] && [ $(sqlite3 "$HASH_FILE" "SELECT EXISTS(SELECT 1 FROM data WHERE hash='$_SRC_HASH' LIMIT 1)") == 0 ]
      then
        sqlite3 "$HASH_FILE" "INSERT INTO data (hash,filename) VALUES (\"$_SRC_HASH\",\"$_SRC_FILE_SHORT\")";
      fi
    fi
  else
    rm -rf "$_DEST_FILE"
    if [ ${_DL_RETRY_COUNT} -lt ${DL_RETRY} ]
    then
      local _DL_RETRY_COUNT=$((${_DL_RETRY_COUNT} + 1))
      if [ -z "$LOG" ]; then ftpgrabEcho "ERROR $_ERROR_DL${_DL_STATUS:0:1}: Download failed... retry $_DL_RETRY_COUNT/3"; fi
      ftpgrabDownloadFile "$_SRC_FILE_PROC" "$_SRC_FILE" "$_DEST_FILE" "$_RESUME" "$_DL_RETRY_COUNT"
    else
      if [ -z "$LOG" ]; then ftpgrabEcho "ERROR $_ERROR_DL${_DL_STATUS:0:1}: Download failed and too many retries..."; fi
    fi
  fi
}

function ftpgrabProcess() {
  local _PATH="$1"
  local _REGEX="$2"
  local _ADDRESS="ftp://$FTP_HOST:$FTP_PORT"
  local _FILES=""
  local _LINECLEAN=""
  local _BASENAME=""
  local _SRC_FILE=""
  local _SRC_FILE_PROC=""
  local _SRC_FILE_SHORT=""
  local _SRC_FILE_SHORT_2=""
  local _DEST_FILE=""
  local _VREGEX=""

  if [ "$DL_METHOD" == "curl" ]
  then
    local _FILES=$(curl --silent --list-only ${FTP_CURL} "$_ADDRESS$_PATH")
  else
    local _FILES=$(wget -q ${FTP_WGET} -O - "$_ADDRESS$_PATH" | grep -o 'ftp:[^"]*')
  fi
  if [ "$DL_SHUFFLE" == "1" ]
  then
    local _FILES=$(echo -e "$_FILES" | shuf)
  fi
  while read -r _LINE
  do
    if [ "$DL_METHOD" == "curl" ]
    then
      if [ "$_LINE" == "." -o "$_LINE" == ".." ]
      then
        continue
      fi
      local _LINECLEAN="$_LINE"
      ftpgrabDebug "checkfolder: curl --silent --list-only $FTP_CURL_HIDECREDS \"$_ADDRESS$(ftpgrabUrlEncode "$_PATH$_LINE")/\""
      curl --silent --list-only ${FTP_CURL} "$_ADDRESS$(ftpgrabUrlEncode "$_PATH$_LINE")/" >/dev/null
      if [ "$?" == "0" ]
      then
        local _LINECLEAN="$_LINE/"
      fi
      local _BASENAME=$(basename "$_LINECLEAN")
      local _SRC_FILE="$_PATH$_BASENAME"
      local _SRC_FILE_PROC="$(ftpgrabUrlEncode "$_PATH$_BASENAME")"
      local _SRC_FILE_SHORT=`echo -n "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")##" | cut -c1-`
      local _SRC_FILE_SHORT_2=`echo -n "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")# #" | cut -c2-`
      local _DEST_FILE=`echo "$_SRC_FILE" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")#$(ftpgrabEscapeSed "$DIR_DEST_REF")#"`
      local _VREGEX=$(echo -n "$_SRC_FILE_SHORT_2" | sed -n "/$_REGEX/p")
    else
      local _LINECLEAN=$(echo "$_LINE" | sed "s#&\#32;#%20#g" | sed "s#$_ADDRESS# #g" | cut -c2-)
      local _BASENAME=$(basename "$_LINECLEAN")
      local _SRC_FILE="$_PATH$_BASENAME"
      local _SRC_FILE_PROC="$_SRC_FILE"
      local _SRC_FILE_SHORT=`echo -n "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")##" | cut -c1-`
      local _SRC_FILE_SHORT_2=`echo -n "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")# #" | cut -c2-`
      local _DEST_FILE=`echo "$(ftpgrabUrlDecode "$_SRC_FILE")" | sed -e "s#$(ftpgrabEscapeSed "$FTP_SRC")#$(ftpgrabEscapeSed "$DIR_DEST_REF")#"`
      local _VREGEX=`echo -n "$_SRC_FILE_SHORT_2" | sed -n "/$_REGEX/p"`
    fi
    ftpgrabDebug "lineClean: $_LINECLEAN"
    ftpgrabDebug "basename: $_BASENAME"
    ftpgrabDebug "srcfile: $_SRC_FILE"
    ftpgrabDebug "srcfileproc: $_SRC_FILE_PROC"
    ftpgrabDebug "srcfileshort: $_SRC_FILE_SHORT"
    ftpgrabDebug "srcfileshort2: $_SRC_FILE_SHORT_2"
    ftpgrabDebug "srchash: $(echo -n "$_SRC_FILE_SHORT" | ${HASH_CMD} - | cut -d ' ' -f 1)"
    ftpgrabDebug "srcsize: $(ftpgrabGetSize "$_SRC_FILE_PROC")"
    ftpgrabDebug "destfile: $_DEST_FILE"
    if [ -f "$_DEST_FILE" ]; then
      ftpgrabDebug "destsize: $(ls -la "$_DEST_FILE" | awk '{print $5}')"
    else
      ftpgrabDebug "destsize: N/A"
    fi
    ftpgrabDebug "vregex: $_VREGEX"
    if [[ "$_LINECLEAN" == */ ]]
    then
      ftpgrabProcess "$_SRC_FILE/" "$_REGEX"
    elif [ ! -z "$_VREGEX" ]
    then
      LOG=""
      local _SKIP_DL=0
      local _RESUME=0
      local _START_TIME=$(awk "BEGIN{srand();print srand()}")
      if [ ${_DEST_FILE: -1} == "/" ]
      then
        mkdir -p "$_DEST_FILE"
      else
        # Start process on a file
        ftpgrabAddLog "Process file: $_SRC_FILE_SHORT"
        local _SRC_HASH=`echo -n "$_SRC_FILE_SHORT" | ${HASH_CMD} - | cut -d ' ' -f 1`
        ftpgrabAddLog "Hash: $_SRC_HASH"
        ftpgrabAddLog "Size: $(ftpgrabGetHumanSize "$_SRC_FILE_PROC")"

        # Check validity
        local _DL_STATUS=$(ftpgrabIsDownloaded "$_SRC_FILE_PROC" "$_SRC_FILE")

        if [ ${_DL_STATUS:0:1} -eq ${FILE_STATUS_NEVER_DL} ]
        then
          ftpgrabAddLog "Status: Never downloaded..."
        elif [ ${_DL_STATUS:0:1} -eq ${FILE_STATUS_SIZE_EQUAL} ]
        then
          local _SKIP_DL=1
          ftpgrabAddLog "Status: Already downloaded and valid. Skip download..."
        elif [ ${_DL_STATUS:0:1} -eq ${FILE_STATUS_SIZE_DIFF} ]
        then
          if [ "$DL_RESUME" == "1" ]; then local _RESUME=1; fi
          ftpgrabAddLog "Status: Exists but sizes are different..."
        elif [ ${_DL_STATUS:0:1} -eq ${FILE_STATUS_HASH_EXISTS} ]
        then
          local _SKIP_DL=1
          ftpgrabAddLog "Status: Hash sum exists. Skip download..."
        fi

        # Check if download skipped and want to hide it in log file
        if [ "$_SKIP_DL" == "0" ] || [ "$DL_HIDE_SKIPPED" == "0" ]; then ftpgrabEcho "$LOG"; LOG=""; fi

        if [ "$_SKIP_DL" == "0" ]
        then
          ftpgrabDownloadFile "$_SRC_FILE_PROC" "$_SRC_FILE" "$_DEST_FILE" "$_RESUME" "0"
        fi

        # Time spent
        local _END_TIME=$(awk 'BEGIN{srand();print srand()}')
        if [ -z "$LOG" ]; then ftpgrabEcho "Time spent: $(ftpgrabFormatSeconds $((_END_TIME - _START_TIME)))"; fi
        if [ -z "$LOG" ]; then ftpgrabEcho "--------------"; fi
      fi
    fi
  done <<< "$_FILES"
}

function ftpgrabStart() {
  # Check FTP_SRC
  FTP_SRC=$(ftpgrabRebuildPath "$(echo $1 | xargs)")
  ftpgrabDebug "FTP_SRC: $FTP_SRC"

  # Check DIR_DEST
  DIR_DEST_REF=$(ftpgrabRebuildPath "$DIR_DEST")
  #if [ "$FTP_SOURCES_CNT" -gt "1" ] && [ "$DL_CREATE_MULTI_BASEDIR" == "1"]; then
  if [ "$FTP_SRC" != "/" ] && [ "$DL_CREATE_BASEDIR" == "1" ]; then
    DIR_DEST_REF="$DIR_DEST_REF$(basename "$FTP_SRC")/"
  fi
  if [ ! -d "$DIR_DEST_REF" ]; then
    mkdir -p "$DIR_DEST_REF"
  fi
  ftpgrabDebug "DIR_DEST: $DIR_DEST"
  ftpgrabDebug "DIR_DEST_REF: $DIR_DEST_REF"

  ftpgrabEcho "Source: ftp://$FTP_HOST:$FTP_PORT$FTP_SRC"
  ftpgrabEcho "Destination: $DIR_DEST_REF"

  # Check connection
  ftpgrabEcho "Checking connection to ftp://$FTP_HOST:$FTP_PORT$FTP_SRC..."
  if [ "$DL_METHOD" == "curl" ]
  then
    ftpgrabDebug "checkConnection: curl --silent --retry 1 --retry-delay 5 $FTP_CURL "ftp://${FTP_HOST}:${FTP_PORT}${FTP_SRC}""
    curl --silent --retry 1 --retry-delay 5 ${FTP_CURL} "ftp://$FTP_HOST:$FTP_PORT$FTP_SRC" >/dev/null
    local _CON_EXIT_CODE="$?"
    if [ ${_CON_EXIT_CODE} != "0" ]
    then
      ftpgrabEcho "ERROR: Curl error $_CON_EXIT_CODE"
      ftpgrabEcho "More infos: https://curl.haxx.se/libcurl/c/libcurl-errors.html"
      exit 1
    fi
  else
    ftpgrabDebug "checkConnection: wget --spider -q --tries=1 --timeout=5 $FTP_WGET -O - "ftp://${FTP_HOST}:${FTP_PORT}${FTP_SRC}""
    wget --spider -q --tries=1 --timeout=5 ${FTP_WGET} -O - "ftp://$FTP_HOST:$FTP_PORT$FTP_SRC"
    local _CON_EXIT_CODE="$?"
    if [ ${_CON_EXIT_CODE} != "0" ]
    then
      ftpgrabEcho "ERROR: Wget error $_CON_EXIT_CODE"
      ftpgrabEcho "More infos: http://www.gnu.org/software/wget/manual/html_node/Exit-Status.html"
      exit 1
    fi
  fi

  ftpgrabEcho "Successfully connected!"
  ftpgrabEcho "--------------"

  # Process
  if [ -z "$DL_REGEX" ]; then DL_REGEX="^.*$;"; fi
  IFS=';' read -ra REGEX <<< "$DL_REGEX"
  for _P in "${REGEX[@]}"; do
    ftpgrabProcess "$FTP_SRC" "$(echo ${_P} | xargs)"
  done
}

function ftpgrabKill() {
  local _CPID="$1"
  if [ -d "/proc/$_CPID" ] && [ -f "/proc/$_CPID/cmdline" ]
  then
    local _CMD_LINE=$(cat "/proc/$_CPID/cmdline")
    kill -9 ${_CPID}
    sleep 2
    local _O_PIDS_FILE=$(find /proc -type f -name "cmdline" | grep '/proc/[1-9][0-9]*/cmdline')
    echo "$_O_PIDS_FILE" | sort | while read _O_PID_FILE
    do
      if [ -f "$_O_PID_FILE" ]
      then
        local _O_CMD_LINE=$(cat "$_O_PID_FILE" 2>/dev/null)
        if [ "$_CMD_LINE" == "$_O_CMD_LINE" ]
        then
          local _O_PID=$(echo "$_O_PID_FILE" | cut -d '/' -f 3)
          if [ ${_O_PID} != $$ ]; then kill -9 ${_O_PID} 2>/dev/null; fi
        fi
      fi
    done
  fi
}

function ftpgrabUrlEncode() {
  echo "$1" | sed 's/%/%25/g; s/ /%20/g; s/ /%09/g; s/!/%21/g; s/"/%22/g; s/#/%23/g; s/\$/%24/g; s/\&/%26/g; s/'\''/%27/g; s/(/%28/g; s/)/%29/g; s/\*/%2a/g; s/+/%2b/g; s/,/%2c/g; s/-/%2d/g; s/:/%3a/g; s/;/%3b/g; s//%3e/g; s/?/%3f/g; s/@/%40/g; s/\[/%5b/g; s/\\/%5c/g; s/\]/%5d/g; s/\^/%5e/g; s/_/%5f/g; s/`/%60/g; s/{/%7b/g; s/|/%7c/g; s/}/%7d/g; s/~/%7e/g; s/      /%09/g;'
}

function ftpgrabUrlDecode() {
  echo "$1" | sed -e "s/%\([0-9A-F][0-9A-F]\)/\\\\\x\1/g" | xargs -0 echo -e
}

function ftpgrabGetSize() {
  if [ "$DL_METHOD" == "curl" ]
  then
    echo $(curl --silent --head ${FTP_CURL} "ftp://$FTP_HOST:$FTP_PORT$1" | grep Content-Length | awk '{print $2}' | tr -d '\r')
  else
    echo $(wget -S --spider ${FTP_WGET} -O - "ftp://$FTP_HOST:$FTP_PORT$1" >&1 2>&1 | grep '^213' | awk '{print $2}')
  fi
}

function ftpgrabGetHumanSize() {
  echo $(ftpgrabGetSize "$1") | awk '{ sum=$1; if (sum < 1024) { printf "%s %s\n",sum,"b"; } else { hum[1024**3]="Gb";hum[1024**2]="Mb";hum[1024]="Kb"; for (x=1024**3; x>=1024; x/=1024){ if (sum>=x) { printf "%.2f %s\n",sum/x,hum[x];break } } }}'
}

function ftpgrabEscapeSed() {
  echo "$1" | sed -e 's/\\/\\\\/g' -e 's/\//\\\//g' -e 's/&/\\\&/g'
}

function ftpgrabChangePerms() {
  local _PATH="$1"
  if [ "$DL_USER" != "" ]; then chown ${DL_USER}.${DL_GROUP} "$_PATH"; fi
  if [ "$DL_CHMOD" != "" ]; then chmod ${DL_CHMOD} "$_PATH"; fi
}

function ftpgrabFormatSeconds() {
  local _S=${1}
  ((_H=_S/3600))
  ((_M=_S%3600/60))
  ((_S=_S%60))
  if [ "${#_H}" == 1 ]; then _H="0"${_H}; fi
  if [ "${#_M}" == 1 ]; then _M="0"${_M}; fi
  if [ "${#_S}" == 1 ]; then _S="0"${_S}; fi
  echo "$_H:$_M:$_S"
}

function ftpgrabRebuildPath() {
  local _PATH="$1"
  local _LEN=${#_PATH}-1
  if [ "${_PATH:_LEN}" != "/" ]; then _PATH="$_PATH/"; fi
  if [ "${_PATH:0:1}" != "/" ]; then _PATH="/$_PATH"; fi
  echo "$_PATH"
}

function ftpgrabAddLog() {
  local _TEXT="$1"
  if [ ! -z "$LOG" ]; then LOG=${LOG}"\n"; fi
  LOG=${LOG}"$_TEXT"
}

function ftpgrabEcho() {
  if [ -f "$LOG_FILE" ]; then
    echo -e "$1" | tee -a "$LOG_FILE"
  else
    echo -e "$1"
  fi
}

function ftpgrabDebug() {
  if [ "$DEBUG" == "1" ]; then
    ftpgrabEcho "#DEBUG $1"
  fi
}

### BEGIN ###

#SCRIPT_NAME=$(basename "$0")
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH

# Default config
DIR_DEST="/tmp/seedbox"
EMAIL_LOG=""
DEBUG=0
FTP_HOST="198.51.100.0"
FTP_PORT="21"
FTP_USER=""
FTP_PASSWORD=""
FTP_SOURCES="/"
FTP_SECURE=0
FTP_CHECK_CERT=0
DL_METHOD="wget"
DL_USER=""
DL_GROUP=""
DL_CHMOD=""
DL_REGEX=""
DL_RETRY=3
DL_RESUME=0
DL_SHUFFLE=0
DL_HIDE_SKIPPED=0
DL_HIDE_PROGRESS=1
DL_CREATE_BASEDIR=0
HASH_ENABLED=1
HASH_TYPE="md5"
HASH_STORAGE="text"

LOG=""
DIR_DEST_REF=$(ftpgrabRebuildPath "$DIR_DEST")

# Check config file
CONFIG_FILE="$CONFIG_DIR/$1"
if [ ! -f "$CONFIG_FILE" ]
then
  echo "ERROR: Config file $CONFIG_FILE not found"
  exit 1
fi

# Read config file
source "$CONFIG_FILE"
BASENAME_FILE=$(basename "$CONFIG_FILE" | cut -d. -f1)

# File status
FILE_STATUS_NEVER_DL=1
FILE_STATUS_SIZE_EQUAL=2
FILE_STATUS_SIZE_DIFF=3
FILE_STATUS_HASH_EXISTS=4

# Destination folder
mkdir -p "$DIR_DEST"
if [ ! -d "$DIR_DEST" ]; then ftpgrabEcho "ERROR: Cannot create dir $DIR_DEST with $(whoami) user"; exit 1; fi
if [ ! -w "$DIR_DEST" ]; then ftpgrabEcho "ERROR: Dir $DIR_DEST is not writable by $(whoami)"; exit 1; fi

# Log folder
LOG_FILE="$LOGS_DIR/$BASENAME_FILE-$(date +%Y%m%d%H%M%S).log"
if [ ! -w "$LOGS_DIR" ]; then echo "ERROR: Dir $LOGS_DIR is not writable by $(whoami)"; exit 1; fi
touch "$LOG_FILE"

# PID folder
mkdir -p "$PID_DIR"
if [ ! -d "$PID_DIR" ]; then ftpgrabEcho "ERROR: Cannot create dir $PID_DIR with $(whoami) user"; exit 1; fi
if [ ! -w "$PID_DIR" ]; then ftpgrabEcho "ERROR: Dir $PID_DIR is not writable by $(whoami)"; exit 1; fi
PID_FILE="$PID_DIR/$BASENAME_FILE.pid"

# Hash folder
mkdir -p "$HASH_DIR"
if [ ! -d "$HASH_DIR" ]; then ftpgrabEcho "ERROR: Cannot create dir $HASH_DIR with $(whoami) user"; exit 1; fi
if [ ! -w "$HASH_DIR" ]; then ftpgrabEcho "ERROR: Dir $HASH_DIR is not writable by $(whoami)"; exit 1; fi

ftpgrabEcho "FTPGrab v4.2.2 ($BASENAME_FILE - $(date +"%Y/%m/%d %H:%M:%S"))"
ftpgrabEcho "--------------"

# Check required packages
if ! type awk > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need awk for this script (try apt-get install awk)"; exit 1; fi
if ! type nawk > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need nawk for this script (try apt-get install nawk)"; exit 1; fi
if ! type gawk > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need gawk for this script (try apt-get install gawk)"; exit 1; fi

# Check conditionnaly required packages
if [[ "$DL_SHUFFLE" == "1" ]] && ! type shuf > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need shuf for this script (try apt-get install shuf)"; exit 1; fi

# Check download method
if [ "$DL_METHOD" == "wget" ] || [ "$DL_METHOD" != "curl" ]
then
  if ! type wget > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need wget for this script (try apt-get install wget)"; exit 1; fi
  DL_METHOD="wget"
elif [ "$HASH_TYPE" == "curl" ]
then
  if ! type curl > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need curl for this script (try apt-get install curl)"; exit 1; fi
  DL_METHOD="curl"
fi

# Check hash type
HASH_CMD=""
if [ "$HASH_TYPE" == "md5" ] || [ "$HASH_TYPE" != "sha1" ]
then
  if ! type md5sum > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need md5sum for this script (try apt-get install md5sum)"; exit 1; fi
  HASH_CMD="md5sum"
elif [ "$HASH_TYPE" == "sha1" ]
then
  if ! type sha1sum > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need sha1sum for this script (try apt-get install sha1sum)"; exit 1; fi
  HASH_CMD="sha1sum"
fi

# Check hash method
if [ "$HASH_STORAGE" == "text" ] || [ "$HASH_STORAGE" != "sqlite3" ]
then
  HASH_STORAGE="text"
  HASH_FILE="$HASH_DIR/$BASENAME_FILE.txt"
elif [ "$HASH_STORAGE" == "sqlite3" ]
then
  if ! type sqlite3 > /dev/null 2>&1; then ftpgrabEcho "ERROR: You need sqlite3 for this script (try apt-get install sqlite3)"; exit 1; fi
  HASH_FILE="$HASH_DIR/$BASENAME_FILE.db"
fi

# Basic command
FTP_CURL="--globoff -u $FTP_USER:$FTP_PASSWORD"
FTP_CURL_HIDECREDS="--globoff -u *****:*****"
FTP_WGET="--ftp-user=$FTP_USER --ftp-password=$FTP_PASSWORD"
FTP_WGET_HIDECREDS="--ftp-user=***** --ftp-password=*****"

# FTP security
if [ "$FTP_SECURE" == "1" ]
then
  FTP_CURL="$FTP_CURL --ftp-ssl"
  FTP_CURL_HIDECREDS="$FTP_CURL_HIDECREDS --ftp-ssl"
  if [ "$FTP_CHECK_CERT" == "0" ]; then
    FTP_CURL="$FTP_CURL --insecure"
    FTP_CURL_HIDECREDS="$FTP_CURL_HIDECREDS --insecure"
  fi
fi

# Check hash file
if [ "$HASH_ENABLED" == "1" -a ! -z "$HASH_FILE" ]
then
  hashfilepath="${HASH_FILE%/*}"
  if [ ! -d "$hashfilepath" ]; then mkdir -p "$hashfilepath"; fi
  if [ ! -f "$HASH_FILE" ]; then touch "$HASH_FILE"; fi
fi
if [ "$HASH_ENABLED" == "1" -a -f "$HASH_FILE" ]; then HASH_ACTIVATED=1; else HASH_ACTIVATED=0; fi

# Init sqlite database
if [ "$HASH_STORAGE" == "sqlite3" -a ! -s "$HASH_FILE" ]; then
  echo "CREATE TABLE data (id INTEGER PRIMARY KEY,hash TEXT,filename TEXT);" > "$HASH_DIR/$BASENAME_FILE.struct"
  sqlite3 "$HASH_FILE" < "$HASH_DIR/$BASENAME_FILE.struct";
  rm -f "$HASH_DIR/$BASENAME_FILE.struct"
fi

# Check ftpgrabProcess already running
currentPid=$$
if [ -f "$PID_FILE" ]
then
  oldPid=$(cat "$PID_FILE")
  if [ -d "/proc/$oldPid" ]
  then
    ftpgrabEcho "ERROR: ftpgrab ($BASENAME_FILE) already running..."
    read -t 10 -p "Do you want to kill the current process? [Y/n] : " choice
    choice=${choice:-timeout}
    echo -n "Do you want to kill the current process? [Y/n] : $choice" >> "$LOG_FILE"
    case "$choice" in
      y|Y)
        ftpgrabKill "$oldPid";;
      n|N)
        exit 1;;
      timeout)
        echo "n"
        exit 1;;
    esac
  fi
fi
echo ${currentPid} > "$PID_FILE"

# Start
starttime=$(awk "BEGIN{srand();print srand()}")

if [ -z "$FTP_SOURCES" ]; then FTP_SOURCES="^.*$;"; fi
IFS=';' read -ra FTP_SRC <<< "$FTP_SOURCES"
FTP_SOURCES_CNT=${#FTP_SRC[@]}

ftpgrabEcho "Config: $BASENAME_FILE"
ftpgrabEcho "Script PID: $currentPid"
ftpgrabEcho "Log file: $LOG_FILE"
ftpgrabEcho "FTP sources count: $FTP_SOURCES_CNT"
ftpgrabEcho "FTP secure: $FTP_SECURE"
ftpgrabEcho "Download method: $DL_METHOD"
if [ ! -z "$DL_REGEX" ]; then ftpgrabEcho "Regex: $DL_REGEX"; fi
ftpgrabEcho "Resume downloads: $DL_RESUME"
ftpgrabEcho "Shuffle file/folder list: $DL_SHUFFLE"
ftpgrabEcho "Hash type: $HASH_TYPE"
ftpgrabEcho "Hash storage: $HASH_STORAGE"
if [ "$HASH_ACTIVATED" == "1" ]; then ftpgrabEcho "Hash file: $HASH_FILE"; fi
ftpgrabEcho "--------------"

for s in "${FTP_SRC[@]}"; do
  ftpgrabStart "$s"
done

# Change perms
if [ "$DL_USER" != "" ]
then
  ftpgrabEcho "Change the ownership recursively of 'Destination' path to $DL_USER:$DL_GROUP"
  chown -R ${DL_USER}:${DL_GROUP} "$DIR_DEST"
fi
if [ "$DL_CHMOD" != "" ]
then
  ftpgrabEcho "Change the access permissions recursively of 'Destination' path to $DL_CHMOD"
  chmod -R ${DL_CHMOD} "$DIR_DEST"
fi
if [ "$DL_USER" != "" ] || [ "$DL_CHMOD" != "" ]
then
  ftpgrabEcho "--------------"
fi

ftpgrabEcho "Finished..."
endtime=$(awk 'BEGIN{srand();print srand()}')
ftpgrabEcho "Total time spent: $(ftpgrabFormatSeconds $((endtime - starttime)))"

rm -f "$PID_FILE"

# Send logs
if [ ! -z "$EMAIL_LOG" ]; then
  (
    echo "To: ${EMAIL_LOG}";
    echo "Subject: FTPGrab on $(hostname)";
    cat "$LOG_FILE"
  ) | sendmail -t
fi

exit 0
