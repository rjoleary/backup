#!/bin/sh

set -eu

BACKUP=$(dirname -- "$0")
TIMESTAMP=$(date '+%Y%m%dT%H%M%S')
SRC=${SRC-*}
DEST=$(readlink --canonicalize ${DEST-/media/ryan/Backup})


########## Github ##########
if [ "$SRC" = 'github' -o "$SRC" = '*' ]; then
    echo "########## GITHUB ##########"
    mkdir -p -- "$DEST/github"
    go run -- "$BACKUP/cmd/scmbackup.go" \
        -update github \
        -index "$DEST/github/index_$TIMESTAMP.json"
    ln -sf -- "$DEST/github/index_$TIMESTAMP.json" "$DEST/github/index.json"
fi


########## Bitbucket ##########
if [ "$SRC" = 'bitbucket' -o "$SRC" = '*' ]; then
    echo "########## BITBUCKET ##########"
    mkdir -p -- "$DEST/bitbucket"
    go run -- "$BACKUP/cmd/scmbackup.go" \
        -update bitbucket \
        -index "$DEST/bitbucket/index_$TIMESTAMP.json"
    ln -sf -- "$DEST/bitbucket/index_$TIMESTAMP.json" "$DEST/bitbucket/index.json"
fi
