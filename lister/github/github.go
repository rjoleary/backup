package github

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/google/go-github/v61/github"
	"github.com/rjoleary/backup/fetcher"
	"github.com/rjoleary/backup/fetcher/git"
)

type GitHub struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (g *GitHub) String() string {
	return fmt.Sprintf("GitHub %s workspace", g.Username)
}

func (g *GitHub) Name() string {
	return "GitHub"
}

func (g *GitHub) Validate() error {
	word := regexp.MustCompile(`^\S+$`)
	if g.Username == "" {
		return errors.New("username is not set")
	}
	if g.Token == "" {
		return errors.New("token is not set")
	}
	if !word.MatchString(g.Username) {
		return errors.New("username is not a single word")
	}
	if !word.MatchString(g.Token) {
		return errors.New("token is not a single word")
	}
	return nil
}

func (g *GitHub) List() ([]fetcher.Fetcher, error) {
	client := github.NewClient(nil).WithAuthToken(g.Token)
	opt := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByUser(context.Background(), g.Username, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Convert to repo type
	fetchers := make([]fetcher.Fetcher, 0, len(allRepos))
	for i := range allRepos {
		r := allRepos[i]
		if r.FullName != nil && r.SSHURL != nil && r.Private != nil {
			fetchers = append(fetchers, &git.Git{
				Dir:      *r.FullName,
				Url:      *r.SSHURL,
				Protocol: "ssh",
				Private:  *r.Private,
			})
		}
	}
	return fetchers, nil
}
