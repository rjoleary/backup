#!/bin/bash
set -e

# Wait until the backup drive is mounted.
until cat /proc/mounts | grep '/media/ryan/Backup'
do
    echo 'Mount the backup drive. Retrying in 5 mins...'
    sleep 5m
done

# Wait until Firefox is closed.
while pidof -s 'firefox'
do
    echo 'Firefox profile cannot be copied while Firefox is running, so close Firefox.'
    exit 1
done

# This script is a modification of http://blog.interlinked.org/tutorials/rsync_time_machine.html
# Also, info for backup and restore of the Firefox profile come from:
#     https://support.mozilla.org/en-US/kb/back-and-restore-information-firefox-profiles
set -e
date=`date --utc '+%Y-%m-%dT%H:%M:%S'`
source_path="$HOME/.mozilla/firefox/*.default"
backup_path='/media/ryan/Backup/Firefox'

rsync -avxP --stats                      \
    --delete-after --delete-excluded     \
    --link-dest=$backup_path/current     \
    $source_path/ $backup_path/$date/

rm -f $backup_path/current
ln -s $date $backup_path/current
