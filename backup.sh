#!/bin/bash
set -e

# Usage: ./backup.sh BACKUP_PATH [TARGETS ...]
# Ex: ./backup.sh /media/ryan/backup bitbucket github dropbox

BACKUP_PATH=${1-test}

# Wait until the backup drive is mounted.
while [ ! -d "$BACKUP_PATH" ]; do
    sleep 2s # TODO: use notify
done

for TARGET in "${@:2}"; do
    if [[ "$TARGET" =~ [^a-zA-Z] ]]; then
        echo "\"$TARGET\" does not match ^[a-zA-Z]$"
    else
        "$(dirname $0)/target/$TARGET.sh" "$BACKUP_PATH/$TARGET"
    fi
done
