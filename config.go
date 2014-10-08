package main

import (
	"flag"
	"gopkg.in/BlueDragonX/yamlcfg.v1"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"os"
	"strings"
)

const (
	DefaultConfigFile = "config.yml"
)

type StringArray []string

func (strs *StringArray) String() string {
	return strings.Join(*strs, ",")
}

func (strs *StringArray) Set(value string) error {
	*strs = append(*strs, value)
	return nil
}

// Root config object.
type Config struct {
	Etcd     EtcdConfig
	Watchers WatchersConfig
	Logging  LoggingConfig
	Exec     []string `yaml:"-"`
}

// SetYAML parses the YAML tree into the object.
func (cfg *Config) SetYAML(tag string, data interface{}) bool {
	yamlcfg.AssertIsMap("config", data)

	if etcdData, ok := yamlcfg.GetMapItem(data, "etcd"); ok {
		cfg.Etcd.SetYAML("etcd", etcdData)
	} else {
		cfg.Etcd = DefaultEtcdConfig()
	}

	if filesData, ok := yamlcfg.GetMapItem(data, "watchers"); ok {
		cfg.Watchers.SetYAML("watchers", filesData)
	} else {
		cfg.Watchers = DefaultWatchersConfig()
	}

	if loggingData, ok := yamlcfg.GetMapItem(data, "logging"); ok {
		cfg.Logging.SetYAML("logging", loggingData)
	} else {
		cfg.Logging = DefaultLoggingConfig()
	}
	return true
}

// Validate the config.
func (cfg *Config) Validate() []error {
	errs := []error{}
	errs = append(errs, cfg.Etcd.Validate()...)
	errs = append(errs, cfg.Logging.Validate()...)
	for _, watcher := range cfg.Watchers {
		errs = append(errs, watcher.Validate()...)
	}
	return errs
}

// Load the Deckhand configuration file.
func LoadConfig() (cfg Config, err error) {
	var file string
	var exec StringArray
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&file, "config", DefaultConfigFile, "YAML configuration file")
	flags.Var(&exec, "exec", "Execute a watcher and exit.")
	flags.Parse(os.Args[1:])

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &cfg)
	cfg.Exec = exec
	return
}
