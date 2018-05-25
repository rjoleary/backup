#!/bin/sh

set -eu

BACKUP=$(dirname -- "$0")
TIMESTAMP=$(date '+%Y%m%dT%H%M%S')
SRC=${SRC-*}
DEST=$(readlink --canonicalize ${DEST-/media/$USER/Backup})


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


rsync_snapshot() {
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
        --link-dest="$2/current"                                            \
        --                                                                  \
        "$1"               `# Trailing slash makes a difference`            \
        "$2/$TIMESTAMP"

    # Update current link for the next --link-dest backup.
    ln -sf -- "$2/$TIMESTAMP" "$2/current"
}


########## Dropbox ##########
if [ "$SRC" = 'dropbox' -o "$SRC" = '*' ]; then
    mkdir -p -- "$DEST/dropbox"
    # TODO: check that dropbox status is up to date before starting a backup
    rsync_snapshot "$HOME/Dropbox/" "$DEST/dropbox"
fi


########## Keys ##########
if [ "$SRC" = 'keys' -o "$SRC" = '*' ]; then
    mkdir -p -- "$DEST/keys/ssh"
    rsync_snapshot "$HOME/.ssh/" "$DEST/keys/ssh"
fi


########## Firefox ##########
if [ "$SRC" = 'firefox' -o "$SRC" = '*' ]; then
    mkdir -p -- "$DEST/firefox"
    if [ -n "$(pidof firefox)" ]; then
        echo "ERROR: firefox is running"
        exit 1
    fi

    for FF_PROFILE in "$HOME"/.mozilla/firefox/*.default; do
        rsync_snapshot "$FF_PROFILE" "$DEST/firefox"
    done
fi


########## Github ##########
if [ "$SRC" = 'github' -o "$SRC" = '*' ]; then
    mkdir -p -- "$DEST/github"
    go run -- "$BACKUP/cmd/scmbackup.go" \
        -dest "$DEST/github"
fi


########## Bitbucket ##########
if [ "$SRC" = 'bitbucket' -o "$SRC" = '*' ]; then
    mkdir -p -- "$DEST/bitbucket"
    go run -- "$BACKUP/cmd/scmbackup.go" \
        -dest "$DEST/bitbucket"
fi
