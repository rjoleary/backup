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
	"regexp"
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

func parseConfig(data []byte) (cfg config, err error) {
	err = json.Unmarshal(data, &cfg)
	return
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

	//src, ok := sources[t.sourceName]
	//if !ok {
	//	return errors.New("bad source name") // TODO: error types
	//}

	// TODO: currently this is debug prints for testing
	if *configFile != "" {
		data, err := ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		cfg, err := parseConfig(data)
		if err != nil {
			log.Fatal(err)
		}
		b, _ := json.MarshalIndent(&cfg, "", "\t")
		fmt.Println(string(b))
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
		if err := sources[t].backup(backupPath, json.RawMessage{}); err != nil {
			log.Fatal(err)
		}
	}
}
