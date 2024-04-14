package local

import (
	"errors"
	"os/exec"
	"strings"
)

type Local struct {
	Dir string `json:"dir"`
}

func (l *Local) String() string {
	return l.Dir
}

func (l *Local) Name() string {
	return "local"
}

func (l *Local) Validate() error {
	if l.Dir == "" {
		return errors.New("dir is required")
	}
	return nil
}

func (l *Local) Fetch(stagingDir string) error {
	return exec.Command("rsync",
		// Archive mode, preserving file permissions, dates, etc..
		"--archive",
		// "This tells rsync to copy the referent of symbolic links that point
		// outside the copied tree.  Absolute symlinks are also treated like
		// ordinary files, and so are any symlinks in the source path itself
		// when --relative is used." ~ man rsync
		"--copy-unsafe-links",
		// "A trailing slash on the source changes this behavior to avoid
		// creating an additional directory level at the destination. You can
		// think of a trailing / on a source as meaning 'copy the contents of
		// this directory' as opposed to 'copy the directory by name', but in
		// both cases the attributes of the containing directory are
		// transferred to the containing directory on the destination." ~ man rsync
		strings.TrimSuffix(l.Dir, "/"), stagingDir+"/").Run()
}
