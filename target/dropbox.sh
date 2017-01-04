#!/bin/bash
set -e

# This script is a modification of http://blog.interlinked.org/tutorials/rsync_time_machine.html
DATE=`date --utc '+%Y-%m-%dT%H:%M:%S'`
SOURCE_PATH="$HOME/Dropbox"
BACKUP_PATH=${1-test}

set -e
rsync -avxP --stats                      \
    --delete-after --delete-excluded     \
    --exclude '.dropbox*'                \
    --link-dest=$backup_path/current     \
    $SOURCE_PATH/ $backup_path/$DATE/

rm -f $backup_path/current
ln -s $DATE $backup_path/current
