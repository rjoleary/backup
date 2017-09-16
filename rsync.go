package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type rsyncSource struct{}

type rsyncConfig struct {
	directory, args string
}

func init() {
	sources["rsync"] = rsyncSource{}
}

func (rsyncSource) backup(backupPath string, config json.RawMessage) error {
	cfg := rsyncConfig{}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}

	// This script is an adaptation of:
	//   http://blog.interlinked.org/tutorials/rsync_time_machine.html
	var (
		date        = time.Now().UTC().Format("2006-01-02T15:04:05")
		srcPath     = os.ExpandEnv("$HOME/Dropbox") // TODO: make not Dropbox-specific
		datedPath   = filepath.Join(backupPath, date)
		currentLink = filepath.Join(backupPath, "current")
	)

	cmd := execCommand(
		"rsync", "-avxP", "--stats",
		"--delete-after", "--delete-excluded",
		"--exclude", ".dropbox*", // TODO: make not Dropbox-specific
		"--link-dest="+currentLink,
		srcPath, datedPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// TODO:
	//rm -f $backup_path/current
	//ln -s $DATE $backup_path/current

	return nil
}
