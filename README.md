# Backup

This repo contains Go scripts for personal backups.

## Suggested settings for backup drive

- Disk format: LUKS + EXT4
  - `sudo apt-get install cryptsetup`
- Disk Name: Backup
- Times are in utc (time is recorded at the *start* of the backup)

## Dependencies

Dependencies depend on what you intend to backup.

- `sudo apt-get install rsync git mercurial`

## Supported sources

Code repositories

- `github`: Backup public and private repos using token authentication.
  - Metadata (stars, ...)
  - Git repositories
- `bitbucket`: Backup public and private repos using password authentication.
  - Metadata
  - Git repositories
  - Mercurial repositories
- `gitlab`: TODO
- `Google Drive`: TODO
- `rsync`: Quickly sync files from source to destination.

## TODO

- encrypt settings file for storing passwords
