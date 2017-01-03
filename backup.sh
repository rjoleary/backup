#!/usr/bin/env sh

# Wait until the backup drive is mounted.
until cat /proc/mounts | grep '/media/ryan/Backup'
do
    sleep 5m
done
