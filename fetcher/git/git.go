package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const known_hosts = `
# https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/githubs-ssh-key-fingerprints
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=

# https://bitbucket.org/site/ssh
bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDQeJzhupRu0u0cdegZIa8e86EG2qOCsIsD1Xw0xSeiPDlCr7kq97NLmMbpKTX6Esc30NuoqEEHCuc7yWtwp8dI76EEEB1VqY9QJq6vk+aySyboD5QF61I/1WeTwu+deCbgKMGbUijeXhtfbxSxm6JwGrXrhBdofTsbKRUsrN1WoNgUa8uqN1Vx6WAJw1JHPhglEGGHea6QICwJOAr/6mrui/oB7pkaWKHj3z7d1IC4KWLtY47elvjbaTlkN04Kc/5LFEirorGYVbt15kAUlqGM65pk6ZBxtaO3+30LVlORZkxOh+LKL/BvbZ/iRNhItLqNyieoQj/uh/7Iv4uyH/cV/0b4WDSd3DptigWq84lJubb9t/DnZlrJazxyDCulTmKdOR7vs9gMTo+uoIrPSb8ScTtvw65+odKAlBj59dhnVp9zd7QUojOpXlL62Aw56U4oO+FALuevvMjiWeavKhJqlR7i5n9srYcrNV7ttmDw7kf/97P5zauIhxcjX+xHv4M=
bitbucket.org ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBPIQmuzMBuKdWeF4+a2sjSSpBK0iqitSQ+5BM9KhpexuGt20JpTVM7u5BDZngncgrqDMbWdxMWWOGtZ9UgbqgZE=
bitbucket.org ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIazEu89wgQZ4bqs3d63QSMzYVa0MuJ2e2gKTKqu+UUO
`

type Git struct {
	Dir      string `json:"dir"`
	Url      string `json:"url"`
	Protocol string `json:"protocol"`
	Private  bool   `json:"private"`
}

func (g *Git) String() string {
	return g.Url
}

func (g *Git) Name() string {
	return "Git"
}

func (g *Git) Validate() error {
	if g.Dir == "" {
		return errors.New("dir is required")
	}
	if g.Url == "" {
		return errors.New("url is required")
	}
	if g.Protocol == "" {
		return errors.New("protocol is required")
	}
	return nil
}

func (g *Git) Fetch(stagingDir string) error {
	dir := filepath.Join(stagingDir, g.Dir)
	os.MkdirAll(dir, 0777)

	f, err := os.CreateTemp("", "backup_known_hosts")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(known_hosts)); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	sshEnv := fmt.Sprintf("GIT_SSH_COMMAND=ssh -o UserKnownHostsFile=%s -o StrictHostKeyChecking=yes", f.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// This command determines if "dir" already contains a git repo. We
	// cannot simply check for the ".git" directory because it is a
	// headless clone.
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	cmd.Env = append(cmd.Environ(), sshEnv)
	cmd.Dir = dir
	out, _ := cmd.Output()

	if strings.TrimSpace(string(out)) == "." {
		// Repo already exists. Update.
		cmd := exec.CommandContext(ctx, "git", "remote", "update")
		cmd.Env = append(cmd.Environ(), sshEnv)
		cmd.Dir = dir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

	} else {
		// Repo is new. Clone for the first time.
		cmd = exec.CommandContext(ctx, "git", "clone", "--mirror", g.Url, ".")
		cmd.Env = append(cmd.Environ(), sshEnv)
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
		cmd = exec.CommandContext(ctx, "git", "config", "gc.auto", "0")
		cmd.Dir = dir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
