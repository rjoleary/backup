// Package staging provides an encrypted mountpoint for staging the backup
// image.
package staging

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"howett.net/plist"
)

const (
	diskName   = "backup.sparseimage"
	volumeName = "backup"
)

type StagingArea struct {
	// tmpDir contains the disk image for New() call.
	tmpDir        string
	diskImagePath string
}

func New(password string, diskSizeGB int) (*StagingArea, error) {
	tmpDir, err := os.MkdirTemp("", "backup")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %v", err)
	}
	diskImagePath := filepath.Join(tmpDir, diskName)

	// Create the disk image.
	cmd := exec.Command("hdiutil", "create",
		// Pass passed to stdin, null-byte terminated.
		"-stdinpass",
		"-size", fmt.Sprintf("%dg", diskSizeGB),
		"-encryption",
		// Unused blocks do not take up space.
		"-type", "SPARSE",
		// HFS+ is the most similiar to Linux's ext4.
		"-fs", "Case-sensitive Journaled HFS+",
		"-volname", volumeName,
		// Automatically mount the image.
		"-attach",
		diskImagePath)
	cmd.Stdin = bytes.NewBufferString(password + "\x00")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create disk image: %v", err)
	}

	return &StagingArea{
		tmpDir: tmpDir,
		// hdiutil always appends .sparseimage
		diskImagePath: diskImagePath,
	}, nil
}

func Open(password string, fileName string) (*StagingArea, error) {
	cmd := exec.Command("hdiutil", "attach",
		// Pass passed to stdin, null-byte terminated.
		"-stdinpass",
		"-readonly",
		fileName)
	cmd.Stdin = bytes.NewBufferString(password + "\x00")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create disk image: %v", err)
	}

	return &StagingArea{
		diskImagePath: fileName,
	}, nil
}

func (sa *StagingArea) MountPoint() (string, error) {
	// This is more complicated than just calling "mount" because there is a mapper to
	output, err := exec.Command("hdiutil", "info", "-plist").Output()
	if err != nil {
		return "", err
	}

	info := struct {
		Framework string `plist:"framework"`
		Images    []struct {
			AutoDiskMount  bool   `plist:"autodiskmount"`
			BlockCount     int    `plist:"blockcount"`
			BlockSize      int    `plist:"blocksize"`
			DiskImages2    bool   `plist:"diskimages2"`
			HDIDPID        int    `plist:"hdid-pid"`
			IconPath       string `plist:"icon-path"`
			ImageEncrypted bool   `plist:"image-encrypted"`
			ImagePath      string `plist:"image-path"`
			ImageType      string `plist:"image-type"`
			OwnerMode      int    `plist:"owner-mode"`
			OwnerUID       int    `plist:"owner-uid"`
			Removable      bool   `plist:"removable"`
			SystemEntities []struct {
				ContentHint string `plist:"content-hint"`
				DevEntry    string `plist:"dev-entry"`
				MountPoint  string `plist:"mount-point"`
			} `plist:"system-entities"`
			Writeable bool `plist:"writeable"`
		} `plist:"images"`
	}{}

	if err := plist.NewDecoder(bytes.NewReader(output)).Decode(&info); err != nil {
		return "", err
	}

	for _, image := range info.Images {
		if image.ImagePath == sa.diskImagePath {
			for _, entity := range image.SystemEntities {
				if entity.MountPoint != "" {
					return entity.MountPoint, nil
				}
			}
			return "", errors.New("could not find mountpoint")
		}
	}
	return "", errors.New("could not find attached image")
}

// Unmount unmounts the filesystem and returns a path to the disk image.
func (sa *StagingArea) Unmount() (string, error) {
	mp, err := sa.MountPoint()
	if err != nil {
		return "", err
	}

	if err := exec.Command("hdiutil", "detach", mp).Run(); err != nil {
		return "", err
	}
	return sa.diskImagePath, nil
}

func (sa *StagingArea) Cleanup() error {
	sa.Unmount()
	if sa.tmpDir == "" {
		return nil
	}
	return os.RemoveAll(sa.tmpDir)
}
