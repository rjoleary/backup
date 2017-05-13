package main

import (
	"path/filepath"
    "os"
    "time"
)

func init() {
	sources["dropbox"] = backupDropbox
}

func backupDropbox(backupPath string) {
	// This script is an adaptation of:
	//   http://blog.interlinked.org/tutorials/rsync_time_machine.html
	var (
        date        = time.Now().UTC().Format("2006-01-02T15:04:05")
		srcPath     = os.ExpandEnv("$HOME/Dropbox")
		datedPath   = filepath.Join(backupPath, date)
		currentLink = filepath.Join(backupPath, "current")
	)

    cmd := execCommand(
		"rsync", "-avxP", "--stats",
		"--delete-after", "--delete-excluded",
		"--exclude", ".dropbox*",
		"--link-dest="+currentLink,
		srcPath, datedPath)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()

    // TODO:
	//rm -f $backup_path/current
	//ln -s $DATE $backup_path/current
}
