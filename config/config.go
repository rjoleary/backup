package config

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/rjoleary/backup/archiver"
	"github.com/rjoleary/backup/archiver/gcs"
	localarchiver "github.com/rjoleary/backup/archiver/local"
	"github.com/rjoleary/backup/fetcher"
	"github.com/rjoleary/backup/fetcher/git"
	localfetcher "github.com/rjoleary/backup/fetcher/local"
	"github.com/rjoleary/backup/lister"
	"github.com/rjoleary/backup/lister/bitbucket"
	"github.com/rjoleary/backup/lister/github"
)

type Config struct {
	Version int    `json:"version"`
	Name    string `json:"name"`

	// Listers
	BitBucket []bitbucket.BitBucket `json:"bitbucket"`
	GitHub    []github.GitHub       `json:"github"`

	// Fetchers
	Git          []git.Git            `json:"git"`
	LocalFetcher []localfetcher.Local `json:"local_fetcher"`

	// Archiver
	GCS           []gcs.GCS             `json:"gcs"`
	LocalArchiver []localarchiver.Local `json:"local_archiver"`
}

func Default() *Config {
	return &Config{
		Version: 1,
		Name:    "new-backup",
	}
}

func Load(file string, password string) (*Config, error) {
	cipherText, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		log.Println("Creating new config file...")
		c := Default()
		return c, c.Save(file, password)
	} else if err != nil {
		return nil, err
	}

	plainText, err := decrypt(cipherText, password)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if err := json.Unmarshal(plainText, c); err != nil {
		return nil, err
	}
	return c, c.Validate()
}

func (c *Config) String() string {
	// Override the default Stringer because the struct contains secrets.
	return "backup config"
}

func (c *Config) Save(file string, password string) error {
	// To avoid data corruption, save to a temporary file and atomically
	// replace the old file.
	tmpFile, err := c.saveTmp(password)
	if err != nil {
		return err
	}
	return os.Rename(tmpFile, file)
}

func (c *Config) saveTmp(password string) (string, error) {
	// Serialize json config.
	plainText, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	cipherText, err := encrypt(plainText, password)
	if err != nil {
		return "", err
	}

	// Open the file for writing.
	// TODO: make on the same file system
	f, err := os.CreateTemp("", "backuprc")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(cipherText); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil { // don't defer intentional
		return "", err
	}
	return f.Name(), nil
}

func (c *Config) Validate() error {
	if c.Version != 1 {
		return errors.New("'version' field must be set to 1")
	}
	for _, l := range c.Listers() {
		if err := l.Validate(); err != nil {
			return err
		}
	}
	for _, f := range c.Fetchers() {
		if err := f.Validate(); err != nil {
			return err
		}
	}
	for _, a := range c.Archivers() {
		if err := a.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) Listers() []lister.Lister {
	listers := []lister.Lister{}
	for _, l := range c.GitHub {
		listers = append(listers, &l)
	}
	for _, l := range c.BitBucket {
		listers = append(listers, &l)
	}
	return listers
}

func (c *Config) Fetchers() []fetcher.Fetcher {
	fetchers := []fetcher.Fetcher{}
	for _, f := range c.LocalFetcher {
		fetchers = append(fetchers, &f)
	}
	for _, f := range c.Git {
		fetchers = append(fetchers, &f)
	}
	return fetchers
}

func (c *Config) Archivers() []archiver.Archiver {
	archivers := []archiver.Archiver{}
	for _, a := range c.GCS {
		archivers = append(archivers, &a)
	}
	for _, a := range c.LocalArchiver {
		archivers = append(archivers, &a)
	}
	return archivers
}
