package gcs

import (
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestArchive(t *testing.T) {
	const testBucketEnvVar = "TEST_BUCKET_NAME"
	testBucket := os.Getenv(testBucketEnvVar)
	if testBucket == "" {
		t.Fatalf("%s environment variable not set. It must be set to the name of a GCS bucket for testing", testBucketEnvVar)
	}

	for _, tt := range []struct {
		name string
		size int64
	}{
		{
			"0bytes",
			0,
		},
		{
			"100MB",
			1 * 1000 * 1000,
		},
		{
			"1GB",
			1 * 1000 * 1000 * 1000,
		},
		{
			"1GiB",
			1 * 1024 * 1024 * 1024,
		},
		{
			"5GB",
			1 * 1000 * 1000 * 1000,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Fill a file with random bytes.
			r := rand.New(rand.NewSource(0))
			f, err := os.Create(filepath.Join(t.TempDir(), tt.name+"_test"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.CopyN(f, r, tt.size); err != nil {
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}

			gcs := &GCS{testBucket}
			if err := gcs.Archive(f.Name()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
