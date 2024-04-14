package staging

import (
	"os"
	"path/filepath"
	"testing"
)

func cpFile(dest, src string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0664)
}

func TestStagingArea(t *testing.T) {
	const (
		testPassword = "testpassword123"
		imageSizeGB  = 1

		testFileName    = "testfile"
		testFileContent = "testtesttest"
	)

	backupFile := filepath.Join(t.TempDir(), "backup.sparseimage")
	var backupSuccessful bool

	t.Run("backup", func(t *testing.T) {
		sa, err := New(testPassword, imageSizeGB)
		if err != nil {
			t.Error(err)
		}
		defer sa.Cleanup()

		mp, err := sa.MountPoint()
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(mp, testFileName), []byte(testFileContent), 0664); err != nil {
			t.Fatal(err)
		}

		image, err := sa.Unmount()
		if err != nil {
			t.Fatal(err)
		}

		// Make a copy of the file to restore as part of the next test.
		if err := cpFile(backupFile, image); err != nil {
			t.Fatal(err)
		}
		backupSuccessful = true
	})

	if !backupSuccessful {
		t.Fatal("backup did not complete")
	}

	t.Run("restore", func(t *testing.T) {
		sa, err := Open(testPassword, backupFile)
		if err != nil {
			t.Fatal(err)
		}
		defer sa.Cleanup()

		mp, err := sa.MountPoint()
		if err != nil {
			t.Fatal(err)
		}

		content, err := os.ReadFile(filepath.Join(mp, testFileName))
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != string(testFileContent) {
			t.Fatalf("ReadFile(%q) = %q; want %q", testFileName, content, testFileContent)
		}
	})
}
