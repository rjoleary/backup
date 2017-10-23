package main

import (
	//"bufio"
	//"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"net/url"
	//"os"
	//"path/filepath"
	//"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type githubSource struct{}

func init() {
	sources["github"] = githubSource{}
}

type githubConfig struct {
	Username, Token *string
}

func getToken(cfg githubConfig) (string, error) {
	if cfg.Token != nil {
		return *cfg.Token, nil
	}

	// TODO: strip whitespace
	fmt.Printf("Token: ")
	token, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return string(token), err
}

func (githubSource) backup(backupPath string, config json.RawMessage) error {
	cfg := githubConfig{}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}

	// TODO: this is weirdly shared with bitbucket getUsername
	username, err := getUsername(cfg.Username)
	if err != nil {
		return err
	}
	_ = username // TODO: username not used

	token, err := getToken(cfg)
	if err != nil {
		return err
	}

	// Download repository list
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/user/repos", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "token "+token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))

	/*bbUrl := fmt.Sprintf(
		"https://%s@api.bitbucket.org/1.0/user/repositories",
		url.UserToken(username, string(token)).String())
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
	}*/

	return nil
}
