# Backup

Some friends were interested in how I do backups. Nothing special. Just a bunch
of ad-hoc scripts.

Disk format: LUKS + EXT4
Disk Name: Backup
sudo apt-get install cryptsetup
times are in utc (time is recorded at the *start* of the backup)

## Dependencies

- rsync
- jq

## TODO

- encrypt settings file for storing passwords
- sources for:
  - Github
  - Gitlab
  - Google Drive
