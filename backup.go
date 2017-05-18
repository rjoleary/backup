package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type backupFunc func(backupDir string) error

var sources = map[string]backupFunc{}

var execCommand = exec.Command

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		backupRoot = fs.String("backup_root", "", "root of the backup directory")
		dryrun     = fs.Bool("dryrun", false, "print commands instead of executing them")
		targetStr  = fs.String("targets", "", "comma separated list of targets")
	)
	fs.Parse(os.Args[1:])

	if *backupRoot == "" {
		log.Fatal("error: no backup_root")
	}

	if *dryrun {
		execCommand = func(cmd string, args ...string) *exec.Cmd {
			line := make([]interface{}, len(args)+1)
			line[0] = cmd
			for i := range args {
				line[i+1] = args[i]
			}
			fmt.Println(line...)
			return nil
		}
	}

	targets := strings.Split(*targetStr, ",")
	alphanum := regexp.MustCompile("^[a-z]+$")
	for _, t := range targets {
		if !alphanum.Match([]byte(t)) {
			log.Fatalf("error: non-alphanumeric target %q", t)
		}
		if _, ok := sources[t]; !ok {
			log.Fatalf("error: unrecognized target %q", t)
		}
	}

	for _, t := range targets {
		backupPath := filepath.Join(*backupRoot, t)
		os.MkdirAll(backupPath, os.ModePerm)
		if err := sources[t](backupPath); err != nil {
			log.Fatal(err)
		}
	}
}
