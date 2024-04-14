# Backup

This repo contains tools for my personal backups. Use at your own risk.

## Download Dependendencies

On MacOSX:

```shell
$ brew install git go rsync vim
```

## Running

To start the backup:

```shell
$ go run .
```

You will be prompted for a password. The config file is encrypted and saved at
`~/backuprc.json.enc`.

The available subcommands are:

* `backup` (default subcommand): Will perform the backup.
* `edit`: Will open Vim to edit the backuprc config file.
* `change-password`: Will change the password on the backuprc config file.

## Architecture

Here is a brief overview of the components in the architecture:

1. The **backup config** is encrypted and stored at `~/.backuprc.json.enc`. The
   config contains settings for one or more Listers and Fetchers.
2. **Listers** query services to determine what should be backed up. It outputs
   one or more Fetchers.
3. **Fetchers** download content into the Staging Area.
4. The **Staging Area** is a mounted encrypted filesystem. Once all the
   downloads are complete, it is unmounted and uploaded to an Archiver.
5. **Archivers** copy the backup image to one or more remote machines.

```
   +------------------------------------+
   | 1. Backup Config                   |
   +------------------------------------+
      |             |
      |             v
      |    +------------------------------------+
      |    | 2. Listers                         |
      |    | +-----------+ +--------+ +-------+ |
      |    | | BitBucket | | GitHub | | etc.. | |
      |    | +-----------+ +--------+ +-------+ |
      |    +------------------------------------+
      |             |
      v             v
   +-----------------------------------+
   | 3. Fetchers                       |
   | +-------+ +-----------+ +-------+ |
   | | Local | | git clone | | etc.. | |
   | +-------+ +-----------+ +-------+ |
   +-----------------------------------+
                    |
                    v
           +-----------------+
           | 4. Staging Area |
           +-----------------+
                    |
                    v
 +----------------------------------------+
 | 5. Upload to archive                   |
 | +-----+ +------------------+ +-------+ |
 | | GCS | | USB Mass Storage | | etc.. | |
 | +-----+ +------------------+ +-------+ |
 +----------------------------------------+
```

## Updating Access Token

### Github

To update the GitHub Access Token.

1. Visit https://github.com/settings/tokens?type=beta and press "Generate new token".
2. Use these settings:
    a) Token name `backup`
    b) Expiration `30 days`
    c) Resource owner `rjoleary`
    d) Repository access `All Repositories`
    e) Contents `Access: Read-only`
    f) Metadata `Access: Read-only`
3. Press "Generate Token" and copy the token.
4. Paste the token into `go run . edit`.

### Git

Run `ssh-add ~/.ssh/<your git key>` before running the backup.

## GCS

```shell
$ brew install google-cloud-sdk
$ gcloud auth login
$ gcloud auth application-default
```

## TODO

Top priority:

- Remote backups
- Cron
- Image file must end in "sparseimage"
- Need directory for .ssh

High priority:

- Backup firefox profile
- Sleep until firefox is closed
- Wait for dropbox to be synced before starting backup
- Per-machine repo for .ssh files

Low priority:

- gitlab repos
- github/bitbucket metadata (stars, ...)
- Make sure times are in UTC
