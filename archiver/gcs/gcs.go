package gcs

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/schollz/progressbar/v3"
)

type GCS struct {
	Bucket string `json:"bucket"`
}

func (l *GCS) String() string {
	return l.Bucket
}

func (l *GCS) Name() string {
	return "GCS"
}

func (l *GCS) Validate() error {
	if l.Bucket == "" {
		return errors.New("bucket must be set")
	}
	return nil
}

func (l *GCS) Archive(diskImage string) error {
	formattedTime := time.Now().Format("2006-01-02T15-04-05")
	// The .sparseimage extension is expected by OSX's Finder to identify the
	// file type.
	fileName := formattedTime + ".sparseimage"

	r, err := os.Open(diskImage)
	if err != nil {
		return err
	}
	defer r.Close()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	w := client.Bucket(l.Bucket).Object(fileName).NewWriter(ctx)
	defer w.Close()
	w.ContentType = ""
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	defer w.Close()

	fi, err := r.Stat()
	if err != nil {
		return err
	}
	bar := progressbar.DefaultBytes(
		fi.Size(),
		"uploading",
	)

	if _, err := io.Copy(io.MultiWriter(w, bar), r); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}
