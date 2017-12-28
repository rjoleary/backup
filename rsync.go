package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type rsyncSource struct{}

type rsyncConfig struct {
	Directory string
	Args      []string
}

func init() {
	sources["rsync"] = rsyncSource{}
}

func (rsyncSource) backup(backupPath string, config json.RawMessage) error {
	cfg := rsyncConfig{}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}

	var (
		date        = time.Now().UTC().Format("2006-01-02T15:04:05")
		srcPath     = os.ExpandEnv(cfg.Directory)
		datedPath   = filepath.Join(backupPath, date)
		currentLink = filepath.Join(backupPath, "current")
	)

	args := append(cfg.Args, "--link-dest="+currentLink, srcPath, datedPath)
	cmd := execCommand("rsync", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// TODO:
	//rm -f $backup_path/current
	//ln -s $DATE $backup_path/current

	return nil
}
