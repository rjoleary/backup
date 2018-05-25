#!/bin/sh

set -eu

BACKUP=$(dirname -- "$0")
TIMESTAMP=$(date '+%Y%m%dT%H%M%S')
SRC=${SRC-*}
DEST=$(readlink --canonicalize ${DEST-/media/ryan/Backup})


# Check that $DEST is mounted.
if [ ! -d "$DEST" ]; then
    echo "Waiting 120s for '$DEST' to be mounted"
    for i in $(seq 120 -1 1); do
        printf "\r${i}s   "
        sleep 1
        if [ -d "$DEST" ]; then
            break
        fi
    done
    if [ ! -d "$DEST" ]; then
        echo "Error: '$DEST' is not mounted"
        exit 1
    fi
fi


########## Dropbox ##########
if [ "$SRC" = 'dropbox' -o "$SRC" = '*' ]; then
    mkdir -p -- "$DEST/dropbox"
    rsync                                                                   \
        --archive          `# Recursive and preserve almost everything`     \
        --verbose          `# List files being copied`                      \
        --progress         `# Print progress bar`                           \
        --one-file-system  `# Don't cross filesystem boundaries`            \
        --partial          `# Keep partially transferred files`             \
        --stats            `# Print stats afterwards`                       \
        --delete-after     `# Delete after all files have been transferred` \
        --delete-excluded  `# Exclude some files from being deleted`        \
        --exclude .dropbox `# Exclude dropbox metadata`                     \
        --link-dest="$DEST/dropbox/current"                                 \
        --                                                                  \
        "$HOME/Dropbox/"   `# Trailing slash makes a difference`            \
        "$DEST/dropbox/$TIMESTAMP"

    # Update current link for the next --link-dest backup.
    ln -sf -- "$DEST/dropbox/$TIMESTAMP" "$DEST/dropbox/current"
fi


########## Github ##########
if [ "$SRC" = 'github' -o "$SRC" = '*' ]; then
    mkdir -p -- "$DEST/github"
    go run -- "$BACKUP/cmd/scmbackup.go" -dest="$DEST/github"
fi


########## Bitbucket ##########
if [ "$SRC" = 'bitbucket' -o "$SRC" = '*' ]; then
    mkdir -p -- "$DEST/bitbucket"
    go run -- "$BACKUP/cmd/scmbackup.go" -dest="$DEST/bitbucket"
fi
