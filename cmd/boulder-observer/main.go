package main

import (
	"flag"
	"io/ioutil"

	"github.com/letsencrypt/boulder/cmd"
	"github.com/letsencrypt/boulder/cmd/boulder-observer/observer"
	"gopkg.in/yaml.v2"
)

func main() {
	configPath := flag.String(
		"config", "config.yaml", "Path to boulder-observer configuration file")
	flag.Parse()

	configYAML, err := ioutil.ReadFile(*configPath)
	cmd.FailOnError(err, "failed to read config file")
	var config observer.NewObs
	err = yaml.Unmarshal(configYAML, &config)
	cmd.FailOnError(err, "failed to parse YAML config")

	observer, err := observer.New(config)
	cmd.FailOnError(err, "failed to initialize boulder-observer daemon")

	observer.Start()
}
