package gcs

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/cenkalti/backoff/v4"
	"github.com/schollz/progressbar/v3"
)

type GCS struct {
	Bucket string `json:"bucket"`
}

func (g *GCS) String() string {
	return g.Bucket
}

func (g *GCS) Name() string {
	return "GCS"
}

func (g *GCS) Validate() error {
	if g.Bucket == "" {
		return errors.New("bucket must be set")
	}
	return nil
}

func (g *GCS) Archive(diskImage string) error {
	// Open file and get its size.
	log.Printf("Preparing archive %q for upload...", diskImage)
	f, err := os.Open(diskImage)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	// Compute CRC32C
	log.Println("Computing checksum...")
	hasher := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Create client
	log.Println("Logging into GCP...")
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(g.Bucket)

	// The .sparseimage extension is expected by OSX's Finder to identify the
	// file type.
	destObject := bucket.Object(time.Now().Format("2006-01-02T15-04-05") + ".sparseimage")

	// Render progress bar.
	log.Printf("Uploading archive to gs://%s/%s...", destObject.BucketName(), destObject.ObjectName())
	bar := progressbar.DefaultBytes(fi.Size(), "uploading")

	// Objects added to this set are deleted at the end.
	tmpObjectNames := map[string]struct{}{}
	defer func() {
		log.Printf("Deleting %d temporary objects from GCS...", len(tmpObjectNames))
		for o := range tmpObjectNames {
			if err := bucket.Object(o).Delete(context.Background()); err != nil {
				log.Printf("Failed to delete temporary object on GCS, %q: %v", o, err)
			}
		}
	}()

	// Chunk the upload to support resumes. The Go client is documented to
	// automatically support resumes from "transient errors", which
	// inconveniently seems to exclude my laptop returning from sleep mode.
	const (
		chunkSize    = 128 * 1024 * 1024
		composeLimit = 32 // defined in https://cloud.google.com/storage/docs/gsutil/commands/compose
	)
	var backOffPolicy = backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)

	// Keep uploading until all the bytes are written.
	var bytesWritten int64
	var lastCRC32C uint32
	var composeObjects []*storage.ObjectHandle
	for bytesWritten < fi.Size() {
		chunkObject := bucket.Object(fmt.Sprintf("%s.part.%d", destObject.ObjectName(), bytesWritten))
		composeObjects = append(composeObjects, chunkObject)

		checkpoint, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		if err := backoff.Retry(
			func() error {
				// Seek to the beginning of the chunk in case upload failed
				// in a previous iteration.
				if _, err := f.Seek(checkpoint, io.SeekStart); err != nil {
					return err
				}

				// Write one chunk.
				w := chunkObject.NewWriter(context.Background())
				w.ContentType = ""
				n, err := io.CopyN(w, f, chunkSize)
				if err != nil && !errors.Is(err, io.EOF) {
					w.Close()
					return err
				}
				if err := w.Close(); err != nil {
					return err
				}

				// Chunk written successful. Update the counters.
				bar.Add(int(n))
				bytesWritten += n
				return nil
			},
			backOffPolicy); err != nil {
			return err
		}
		tmpObjectNames[chunkObject.ObjectName()] = struct{}{}

		// Concatenate all the files into the destination file.  GCS limits how
		// many objects can be composed at once, so this is performed
		// periodically.
		if len(composeObjects) == composeLimit || bytesWritten == fi.Size() {
			attrs, err := backoff.RetryWithData(func() (*storage.ObjectAttrs, error) {
				return destObject.ComposerFrom(composeObjects...).Run(context.Background())
			}, backOffPolicy)
			if err != nil {
				return fmt.Errorf("failed to compose objects into %q: %v", destObject.ObjectName(), err)
			}
			lastCRC32C = attrs.CRC32C
			composeObjects = []*storage.ObjectHandle{destObject}
		}
	}
	log.Printf("Archive uploaded to gs://%s/%s", destObject.BucketName(), destObject.ObjectName())

	// Verify checksum.
	if lastCRC32C != hasher.Sum32() {
		return fmt.Errorf("uploaded hash does not match, got %#x, want %#x", lastCRC32C, hasher.Sum32())
	}
	log.Println("Checksum verified.")
	return nil
}
