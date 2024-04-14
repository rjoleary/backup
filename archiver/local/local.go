package local

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Local struct {
	Directory string `json:"directory"`
}

func (l *Local) String() string {
	return l.Directory
}

func (l *Local) Name() string {
	return "Local"
}

func (l *Local) Validate() error {
	if l.Directory == "" {
		return errors.New("directory must be set")
	}
	return nil
}

func (l *Local) Archive(diskImage string) error {
	formattedTime := time.Now().Format("2006-01-02T15-04-05")
	// The .sparseimage extension is expected by OSX's Finder to identify the
	// file type.
	fileName := filepath.Join(l.Directory, formattedTime+".sparseimage")
	if err := copyFile(fileName, diskImage); err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}
	return nil
}

func copyFile(dest, src string) error {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open: %v", err)
	}
	defer r.Close()

	w, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("could not open: %v", err)
	}
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("could not copy: %v", err)
	}
	return nil
}
