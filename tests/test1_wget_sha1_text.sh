#!/bin/bash

source common.sh
TEST_NAME=$(echo $(basename $0) | sed -e 's/^[SK][0-9]*//' -e 's/\.sh$//')

CONFIG_FILE="/opt/ftpgrab/conf/$TEST_NAME.conf"
DIR_DEST="/tmp/$TEST_NAME"

# Clean
rm -f "$CONFIG_FILE"
rm -rf "$DIR_DEST"
rm -f "/opt/ftpgrab/hash/$TEST_NAME*"

# Edit config
cp -f "$DEFAULT_CONFIG_FILE" "$CONFIG_FILE"
sed "s#DIR_DEST=\"$DEFAULT_DIR_DEST\"#DIR_DEST=\"$DIR_DEST\"#" -i "$CONFIG_FILE"
#sed "s#DEBUG=0#DEBUG=1#" -i "$CONFIG_FILE"
sed "s#FTP_HOST=\"$DEFAULT_FTP_HOST\"#FTP_HOST=\"$SERVER1_IP\"#" -i "$CONFIG_FILE"
sed "s#FTP_USER=\"\"#FTP_USER=\"$SERVER1_USER\"#" -i "$CONFIG_FILE"
sed "s#FTP_PASSWORD=\"\"#FTP_PASSWORD=\"$SERVER1_PASSWORD\"#" -i "$CONFIG_FILE"
sed "s#HASH_TYPE=\"$DEFAULT_HASH_TYPE\"#HASH_TYPE=\"sha1\"#" -i "$CONFIG_FILE"

# Launch
/etc/init.d/ftpgrab "$TEST_NAME.conf"

# Relaunch to check hashes
/etc/init.d/ftpgrab "$TEST_NAME.conf"
