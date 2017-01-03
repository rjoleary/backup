#!/bin/bash
set -e

backup_path='/media/ryan/Backup'
# TODO: A better backup system would make each one of these lines part of their respective backup scripts.
#       This would allow me to quickly check which items are included in a backup.
[ -d $backup_path/Bitbucket ]     || mkdir $backup_path/Bitbucket
[ -d $backup_path/Contacts ]      || mkdir $backup_path/Contacts
[ -d $backup_path/Dropbox ]       || mkdir $backup_path/Dropbox
[ -d $backup_path/Email ]         || mkdir $backup_path/Email
[ -d $backup_path/Firefox ]       || mkdir $backup_path/Firefox
[ -d $backup_path/Gists ]         || mkdir $backup_path/Gists
[ -d $backup_path/GitHub ]        || mkdir $backup_path/GitHub
[ -d $backup_path/Google\ Drive ] || mkdir $backup_path/Google\ Drive
