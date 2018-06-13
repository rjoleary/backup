# Encryption Setup

## Initial Creation

    sudo apt-get install cryptsetup

    # Alternatively, use dd if=/dev/zero if you do not want a sparse file.
    truncate backup.luks 512G
    sudo luksformat backup.luks
    sudo cryptsetup luksOpen backup.luks backup
    sudo mkfs.ext4 /dev/mapper/backup

## Opening

    sudo cryptsetup luksOpen backup.luks backup
    mkdir backup
    sudo mount /dev/mapper/backup backup

## Closing

    sudo umount backup
    sudo cryptsetup luksClose backup

## Status

    sudo cryptsetup status backup.luks

## Trim

Only necessecary when using a spare file or SSD. Perform weekly:

    sudo cryptsetup luksOpen backup.luks backup --allow-discards
    mkdir backup
    sudo mount /dev/mapper/backup backup
    sudo fstrim backup

## References

- Manpage is cryptsetup(8).
- https://serverfault.com/questions/696554/creating-a-grow-on-demand-encrypted-volume-with-luks
