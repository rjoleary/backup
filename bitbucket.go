package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type bitbucketSource struct{}

func init() {
	sources["bitbucket"] = bitbucketSource{}
}

type bitbucketConfig struct {
	username, password string
}

func (bitbucketSource) newConfig() interface{} {
	return bitbucketConfig{}
}

func (bitbucketSource) backup(backupPath string, config interface{}) error {
	// Username
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	username = strings.TrimSpace(username)

	// Password
	fmt.Printf("Password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return err
	}

	// Download repository list
	bbUrl := fmt.Sprintf(
		"https://%s@api.bitbucket.org/1.0/user/repositories",
		url.UserPassword(username, string(password)).String())
	resp, err := http.Get(bbUrl)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Save repository index
	indexPath := filepath.Join(backupPath, "index.json")
	out := bytes.Buffer{}
	if err := json.Indent(&out, body, "", "  "); err != nil {
		return err
	}
	if err := ioutil.WriteFile(indexPath, out.Bytes(), os.ModePerm); err != nil {
		return err
	}

	// Parse json
	repos := []struct {
		Scm   string `json:"scm"`
		Owner string `json:"owner"`
		Slug  string `json:"slug"`
	}{}
	if err := json.Unmarshal(body, &repos); err != nil {
		return err
	}

	// Clone repositories
	for _, v := range repos {
		if v.Scm == "git" {
			repo := fmt.Sprintf("git@bitbucket.org:%s/%s.git", v.Owner, v.Slug)
			cmd := execCommand("git", "clone", "--mirror", repo)
			cmd.Dir = backupPath
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}

		// TODO: mercurial repositories
	}

	return nil
}
