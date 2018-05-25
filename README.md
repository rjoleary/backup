# Backup

This repo contains scripts for personal backups.

## Suggested settings for backup drive

- Disk format: LUKS + EXT4
  - `sudo apt-get install cryptsetup`
- Disk Name: Backup
- Times are in utc (time is recorded at the *start* of the backup)

## Dependencies

Depending on which script you run:

- `rsync`, `git`, `hg` (mercurial)

## Scripts

- `backup.sh` does the actual backup.
- `update_index.sh` updates the list of SCM repositories downloaded from
  Github/Bitbucket/Gitlab APIs.

The following environment variables can be overriden:

- `SRC`: What to backup. See the sources section below. (default: `*`)
- `DEST`: Where to backup. (default: "/media/$USER/Backup")
- `BACKUP`: The directory of this
- `TIMESTAMP`: The timestamp used for snapshots. (default: `$(date '+%Y%m%dT%H%M%S')`)

For best results, run `ssh-add ~/.ssh/<your github key>` before running the
backup; otherwise, you will be prompted for your passphrase multiple times.

## Sources

- `dropbox`
  - Source: `$HOME/Dropbox`
  - Dest:   `$DEST/dropbox/$TIMESTAMP/`
- `keys`
  - Source: `$HOME/.ssh`
  - Dest:   `$DEST/dropbox/$TIMESTAMP/`
- `firefox`
  - Source: `$HOME/.mozilla/firefox/*.default/`
  - Dest:   `$DEST/firefox/$TIMESTAMP/*.default/`
- `github`
  - Source: Clone/pull repositories listed in the index `$DEST/github/index.json`.
  - Dest:   `$DEST/github/**/`
- `bitbucket`
  - Source: Clone/pull repositories listed in the index `$DEST/bitbucket/index.json`.
  - Dest:   `$DEST/bitbucket/**/`

Note: git repositories are headless mirrors with garbage collection turned off.

## TODO

High priority:

- Make per-machine repos for keys and firefox profiles
- Remote backups
- Cron
- Google Drive
- Canaries and restore

Low priority:

- gitlab repos
- github/bitbucket metadata (stars, ...)
- hg repos
- Explain how hard links work
- Wait for dropbox to be synced before starting backup
- Full home directory backup
- Make sure times are in UTC
