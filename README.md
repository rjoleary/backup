# Backup

This repo contains scripts for personal backups.

## Suggested settings for backup drive

- Disk format: LUKS + EXT4
  - `sudo apt-get install cryptsetup`
- Disk Name: Backup
- Times are in utc (time is recorded at the *start* of the backup)

## Dependencies

Dependencies depend on what you intend to backup.

- `sudo apt-get install rsync git mercurial`

## Arguments

`backup -f CONFIG SOURCES... TARGET`

- `-f FILENAME`: Specify configuration file

## Configuration

    {
      "sources": {
        "dropbox": "$HOME/Dropbox"
      },
      "target": {
        "/media/ryan
      }
      "recipes": {
        "rsync": {
          "type": "command",
          "args": [
            "rsync",
            "-avxP",
            "--stats",
            "--delete-after",
            "--delete-excluded",
            "--exclude",
            ".dropbox*"
          ]
        }
      }
    }

## Supported sources

### Shell

- `source`:


For example, this syncs the user's Dropbox directory to the destination:

### Command


### github


### bitbucket


## TODO

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
