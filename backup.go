package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/rjoleary/backup/config"
	"github.com/rjoleary/backup/fetcher"
	"github.com/rjoleary/backup/staging"
	"golang.org/x/term"
)

type flags struct {
	skipCheckDeps bool
	configFile    string
	fetcherMask   sliceFlag
	// password is not passed as a flag as that would be insecure (ex:
	// bash_history). Rather, the user is prompted for to enter a password.
	// DO NOT EXPORT to prevent accidental leakage with fmt/reflect packages.
	password string
}

type sliceFlag []string

func (s *sliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *sliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// checkDeps checks whether the dependencies are installed.
// TODO: Move into plugins.
func checkDeps() error {
	deps := []struct {
		name string
		cmd  []string
	}{
		{"git", []string{"git", "--version"}},
		{"rsync", []string{"rsync", "--version"}},
		{"vim", []string{"vim", "--version"}},
		{"hdiutil", []string{"hdiutil", "help"}},
	}

	// Check all the deps in parallel.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type depResult struct {
		depName string
		err     error
	}
	errs := make(chan depResult)
	for _, dep := range deps {
		go func() {
			errs <- depResult{
				depName: dep.name,
				err:     exec.CommandContext(ctx, dep.cmd[0], dep.cmd[1:]...).Run(),
			}
		}()
	}

	// Collect the results.
	missingDeps := []string{}
	for range deps {
		if err := <-errs; err.err != nil {
			missingDeps = append(missingDeps, err.depName)
		}
	}
	if len(missingDeps) > 0 {
		return fmt.Errorf("missing dependencies: %v", missingDeps)
	}
	return nil
}

func askForPassword() (string, error) {
	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password), err
}

func mask(fetchers []fetcher.Fetcher, mask []string) []fetcher.Fetcher {
	filteredFetchers := []fetcher.Fetcher{}
	for _, f := range fetchers {
		inMask := false
		for _, m := range mask {
			if f.Name() == m {
				inMask = true
				break
			}
		}
		if !inMask {
			filteredFetchers = append(filteredFetchers, f)
		} else {
			log.Printf("Masking filter %v", f)
		}
	}
	return filteredFetchers
}

func backupCommand(f flags, c *config.Config, args []string) error {
	if len(c.Listers()) == 0 && len(c.Fetchers()) == 0 {
		log.Println("Config is empty")
		return nil
	}

	numErrors := 0

	log.Println("Creating staging area...")
	sa, err := staging.New(f.password, 512)
	if err != nil {
		return fmt.Errorf("failed to create staging directory: %v", err)
	}
	defer sa.Cleanup()
	mp, err := sa.MountPoint()
	if err != nil {
		return fmt.Errorf("failed to get mount point: %v", err)
	}

	allFetchers := c.Fetchers()

	for _, l := range c.Listers() {
		log.Printf("Listing %s...", l)
		fetchers, err := l.List()
		if err != nil {
			log.Println("Error:", err)
			numErrors++
			continue
		}

		counter := map[string]int{}
		for i := range fetchers {
			counter[fetchers[i].Name()]++
		}
		for name, count := range counter {
			log.Printf("Found %d %s fetchers", count, name)
		}

		allFetchers = append(allFetchers, fetchers...)
	}

	allFetchers = mask(allFetchers, f.fetcherMask)

	if len(allFetchers) == 0 {
		log.Println("Nothing to fetch")
		return nil
	}

	sort.Slice(allFetchers, func(i, j int) bool {
		return allFetchers[i].Name() < allFetchers[j].Name()
	})

	for _, f := range allFetchers {
		log.Printf("Fetching %s...", f)
		if err := f.Fetch(mp); err != nil {
			log.Printf("Error fetching %s: %v", f, err)
			numErrors++
			continue
		}
	}

	log.Println("Unmounting staging area...")
	diskImage, err := sa.Unmount()
	if err != nil {
		return fmt.Errorf("failed to unmount staging area: %v", err)
	}

	for _, a := range c.Archivers() {
		log.Printf("Archiving %s...", a)
		if err := a.Archive(diskImage); err != nil {
			log.Printf("Error archiving %s: %v", a, err)
			numErrors++
			continue
		}
	}

	if numErrors != 0 {
		return fmt.Errorf("encountered %d error(s)", numErrors)
	}
	return nil
}

func changePasswordCommand(f flags, c *config.Config, args []string) error {
	newPassword, err := askForPassword()
	if err != nil {
		return err
	}
	return c.Save(f.configFile, newPassword)
}

func editCommand(f flags, c *config.Config, args []string) error {
	// Serialize json config.
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Save to temporary file.
	tmp, err := os.CreateTemp("", "backuprc")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	tmp.Close()

	for {
		// Run vim.
		//   -n: Disable swap files.
		cmd := exec.Command("vim", "-n", tmp.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run vim: %v", err)
		}

		// Read temporary file.
		data, err = os.ReadFile(tmp.Name())
		if err != nil {
			return err
		}

		// Deserialize json config.
		if err := json.Unmarshal(data, c); err != nil {
			log.Printf("Error: validation failed %v", err)
			continue
		}
		if err := c.Validate(); err != nil {
			log.Printf("Error: validation failed %v", err)
			continue
		}

		break
	}
	return c.Save(f.configFile, f.password)
}

func realMain() error {
	// Parse flags.
	f := flags{}
	flag.BoolVar(&f.skipCheckDeps, "skip-check-deps", false, "Skip checking dependencies")
	flag.StringVar(&f.configFile, "config-file", os.ExpandEnv("$HOME/.backuprc.json.enc"), "Configuration file")
	flag.Var(&f.fetcherMask, "fetcher-mask", "Skip these fetchers")
	flag.Parse()

	// Check dependencies.
	if !f.skipCheckDeps {
		log.Println("Checking dependencies...")
		if err := checkDeps(); err != nil {
			return err
		}
	}

	// Decrypt config file.
	log.Println("Decrypting config file...")
	var err error
	if f.password, err = askForPassword(); err != nil {
		return err
	}
	c, err := config.Load(f.configFile, f.password)
	if err != nil {
		return err
	}

	availableCmds := map[string]func(flags, *config.Config, []string) error{
		"backup":          backupCommand,
		"change-password": changePasswordCommand,
		"edit":            editCommand,
	}

	// Default to "backup" command.
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"backup"}
	}

	if args[0] == "help" {
		flag.Usage()
		return nil
	}

	// Execute the command.
	if cmd, ok := availableCmds[args[0]]; ok {
		return cmd(f, c, args[1:])
	} else {
		flag.Usage()
	}
	return nil
}

func main() {
	if err := realMain(); err != nil {
		log.Fatalln("Error:", err)
	}
}
