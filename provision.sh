#!/bin/bash
set -e

# Provision a directory as a new backup target. This script is idempotent
# meaning it is also useful for re-provisioning a directory.
# Usage: ./provision BACKUP_PATH

BACKUP_PATH=${1-test}

if [ ! -d "$BACKUP_PATH" ]; then
    echo "\"$BACKUP_PATH\" is not a directory" >&2
    exit 1
fi

# TODO: A better backup system would make each one of these lines part of their respective backup scripts.
for FOLDER in bitbucket contacts dropbox email firefox gists github google_drive; do
    [ -d "$BACKUP_PATH/$FOLDER" ] || mkdir "$BACKUP_PATH/$FOLDER"
done
