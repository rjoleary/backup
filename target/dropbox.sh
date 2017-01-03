#!/bin/bash
set -e

# Wait until the backup drive is mounted.
until cat /proc/mounts | grep '/media/ryan/Backup'
do
    echo 'Mount the backup drive. Retrying in 5 mins...'
    sleep 5m
done

# This script is a modification of http://blog.interlinked.org/tutorials/rsync_time_machine.html
date=`date --utc '+%Y-%m-%dT%H:%M:%S'`
source_path="$HOME/Dropbox"
backup_path='/media/ryan/Backup/Dropbox'

set -e
rsync -avxP --stats                      \
    --delete-after --delete-excluded     \
    --exclude '.dropbox*'                \
    --link-dest=$backup_path/current     \
    $source_path/ $backup_path/$date/

rm -f $backup_path/current
ln -s $date $backup_path/current
