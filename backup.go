package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type config struct {
	Targets map[string]json.RawMessage `json:"targets"`
}

type source interface {
	backup(backupPath string, config json.RawMessage) error
}

// sources is registered by individual source files.
var sources = map[string]source{}

var execCommand = exec.Command

func parseConfigFile(data []byte) (cfg config, err error) {
	err = json.Unmarshal(data, &cfg)
	return
}

func readConfigFile(filename string) (config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config{}, err
	}
	return parseConfigFile(data)
}

func (c *config) getSourceName(target string, config json.RawMessage) (string, error) {
	// Unmarshal enough of the json to get the source.
	sourceStruct := struct {
		Source *string `json:"source"`
	}{}
	if err := json.Unmarshal(config, &sourceStruct); err != nil {
		return "", err
	}
	if sourceStruct.Source == nil {
		return target, nil
	}
	return *sourceStruct.Source, nil
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		backupRoot = fs.String("backup_root", "", "root of the backup directory")
		dryrun     = fs.Bool("dryrun", false, "print commands instead of executing them")
		targetStr  = fs.String("targets", "", "comma separated list of targets")
		configFile = fs.String("config_file", "", "location of configuration file")
	)
	fs.Parse(os.Args[1:])

	if *backupRoot == "" {
		log.Fatal("error: no backup_root")
	}
	if *configFile == "" {
		log.Fatal("error: config file required")
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

	cfg, err := readConfigFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	targets := strings.Split(*targetStr, ",")
	for _, t := range targets {
		if _, ok := cfg.Targets[t]; !ok {
			log.Fatalf("error: unrecognized target %q", t)
		}
	}

	for _, t := range targets {
		backupPath := filepath.Join(*backupRoot, t)
		os.MkdirAll(backupPath, os.ModePerm)
		if err := sources[t].backup(backupPath, cfg.Targets[t]); err != nil {
			log.Fatal(err)
		}
	}
}
