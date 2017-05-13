package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	sources["bitbucket"] = backupBitbucket
}

func backupBitbucket(backupPath string) error {
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

	// Parse json

	// Save repository list
	fmt.Printf(string(body))

	return nil
}

/*jq '.[]
    | select(.scm == "git")
    | "git@bitbucket.org:" + .owner + "/" + .slug + ".git"' bitbucket_index.json \
    | xargs -P0 -n1 git clone --mirror

# TODO: mercurial repos
*/
