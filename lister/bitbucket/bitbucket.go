package bitbucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/rjoleary/backup/fetcher"
	"github.com/rjoleary/backup/fetcher/git"
)

type BitBucket struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (b *BitBucket) String() string {
	return fmt.Sprintf("BitBucket %s workspace", b.Username)
}

func (b *BitBucket) Name() string {
	return "BitBucket"
}

func (b *BitBucket) Validate() error {
	word := regexp.MustCompile(`^\S+$`)
	if b.Username == "" {
		return errors.New("username is not set")
	}
	if b.Password == "" {
		return errors.New("password is not set")
	}
	if !word.MatchString(b.Username) {
		return errors.New("username is not a single word")
	}
	if !word.MatchString(b.Password) {
		return errors.New("password is not a single word")
	}
	return nil
}

func (b *BitBucket) List() ([]fetcher.Fetcher, error) {
	// Download repository index
	bbUrl := fmt.Sprintf(
		"https://%s@api.bitbucket.org/2.0/repositories/%s",
		url.UserPassword(b.Username, b.Password), b.Username)
	resp, err := http.Get(bbUrl)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Parse json
	parsed := struct {
		Values []struct {
			FullName  string `json:"full_name"`
			IsPrivate bool   `json:"is_private"`
		} `json:"values"`
	}{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	fetchers := make([]fetcher.Fetcher, 0, len(parsed.Values))
	for i := range parsed.Values {
		r := parsed.Values[i]
		fetchers = append(fetchers, &git.Git{
			Dir:      r.FullName,
			Url:      fmt.Sprintf("git@bitbucket.org:%s.git", r.FullName),
			Protocol: "ssh",
			Private:  r.IsPrivate,
		})
	}
	return fetchers, nil
}
