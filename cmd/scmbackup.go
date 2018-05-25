// Backup SCM repositories.
//
// Synopsis:
//
//     scmbackup [ARGS...]
//
// Arguments:
//
//     -dest=STRING: Destination folder
//     -index=STRING: Override the index file
//     -update=STRING: Update the index without a backup
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type repo struct {
	Scm      string `json:"scm"`
	Dir      string `json:"dir"`
	Url      string `json:"url"`
	Protocol string `json:"protocol"`
	Private  bool   `json:"private"`
}

func getToken() (string, error) {
	fmt.Print("Token: ")
	token, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return strings.TrimSpace(string(token)), err
}

func getUsername() (string, error) {
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	username = strings.TrimSpace(username)
	return username, nil
}

func getPassword() (string, error) {
	fmt.Printf("Password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return string(password), err
}

func updateGithub() ([]repo, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	// Download repository index
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/user/repos", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse json
	ghRepos := []struct {
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
		CloneUrl string `json:"clone_url"`
		SshUrl   string `json:"ssh_url"`
	}{}
	if err := json.Unmarshal(body, &ghRepos); err != nil {
		if len(body) < 80 {
			log.Print("Invalid JSON: ", string(body))
		}
		return nil, err
	}

	// Convert to repo type
	repos := make([]repo, len(ghRepos))
	for i, _ := range ghRepos {
		repos[i].Scm = "git"
		repos[i].Dir = ghRepos[i].FullName
		repos[i].Url = ghRepos[i].SshUrl
		repos[i].Protocol = "ssh"
		repos[i].Private = ghRepos[i].Private
	}
	return repos, nil
}

func updateBitbucket() ([]repo, error) {
	username, err := getUsername()
	if err != nil {
		return nil, err
	}

	token, err := getPassword()
	if err != nil {
		return nil, err
	}

	// Download repository index
	bbUrl := fmt.Sprintf(
		"https://%s@api.bitbucket.org/1.0/user/repositories",
		url.UserPassword(username, token))
	resp, err := http.Get(bbUrl)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Parse json
	bbRepos := []struct {
		Scm       string `json:"scm"`
		Owner     string `json:"owner"`
		Slug      string `json:"slug"`
		IsPrivate bool   `json:"is_private"`
	}{}
	if err := json.Unmarshal(body, &bbRepos); err != nil {
		return nil, err
	}

	// Convert to repo type
	repos := make([]repo, len(bbRepos))
	for i, _ := range bbRepos {
		repos[i].Scm = bbRepos[i].Scm
		repos[i].Dir = path.Join(bbRepos[i].Owner, bbRepos[i].Slug)
		repos[i].Url = fmt.Sprintf("git@bitbucket.org:%s.git", repos[i].Dir)
		repos[i].Protocol = "ssh"
		repos[i].Private = bbRepos[i].IsPrivate
	}
	return repos, nil
}

// Save repository index
func saveIndex(repos []repo, indexPath string) error {
	out, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(indexPath, out, os.ModePerm)
}

// Load repository index
func loadIndex(indexPath string) ([]repo, error) {
	in, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}
	var repos []repo
	err = json.Unmarshal(in, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func backup(repos []repo, dest string) error {
	// Clone repositories
	for _, r := range repos {
		dir := filepath.Join(dest, r.Dir)
		os.MkdirAll(dir, 0777)

		if r.Scm == "git" {
			// This command determines if "dir" already contains a git repo. We
			// cannot simply check for the ".git" directory because it is a
			// headless clone.
			cmd := exec.Command("git", "rev-parse", "--git-dir")
			cmd.Dir = dir
			out, _ := cmd.Output()

			if strings.TrimSpace(string(out)) == "." {
				// Repo already exists. Update.
				cmd := exec.Command("git", "remote", "update")
				cmd.Dir = dir
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}

			} else {
				// Repo is new. Clone for the first time.
				cmd = exec.Command("git", "clone", "--mirror", r.Url, ".")
				cmd.Dir = dir
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}

				// Disable git's garbage collection. In case anything is
				// deleted from git (deleted branch or a force push), the
				// backups will still contain those lost commits.
				cmd = exec.Command("git", "config", "gc.auto", "0")
				cmd.Dir = dir
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}
			}
		}

		// TODO: mercurial repositories
	}
	return nil
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		dest   = fs.String("dest", "", "destination folder")
		index  = fs.String("index", "", "override the index file")
		update = fs.String("update", "", "update the index without a backup")
	)
	fs.Parse(os.Args[1:])

	// TODO: change synopsis to:
	//    scmbackup [ backup | index ] [ OPTIONS ... ]
	if *dest == "" && *update == "" {
		log.Fatal("dest or update argument is required")
	}

	// Default index
	if *index == "" {
		*index = filepath.Join(*dest, "index.json")
	}

	switch *update {
	case "github":
		repos, err := updateGithub()
		if err != nil {
			log.Fatal(err)
		}
		saveIndex(repos, *index)

	case "bitbucket":
		repos, err := updateBitbucket()
		if err != nil {
			log.Fatal(err)
		}
		saveIndex(repos, *index)

	case "":
		repos, err := loadIndex(*index)
		if err != nil {
			log.Fatal(err)
		}
		err = backup(repos, *dest)
		if err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatal("invalid update value")
	}
}
