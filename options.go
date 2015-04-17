package main

import (
	"flag"
	"strings"
)

var DefaultConfigFile string = "/etc/sentinel.yml"

// A string array option capable of being appended to.
type stringsOpt []string

func (strs *stringsOpt) String() string {
	return strings.Join(*strs, ",")
}

func (strs *stringsOpt) Set(value string) error {
	*strs = append(*strs, strings.Split(value, ",")...)
	return nil
}

// Store values retrieved from the cli.
type Options struct {
	Config    string
	Exec      []string
	Etcd      []string
	Prefix    string
	LogTarget string
	LogLevel  string
}

// Parse cli options. Exit on failure.
func ParseOptionsOrExit(args []string) *Options {
	var config string
	var exec stringsOpt
	var etcd stringsOpt
	var prefix string
	var logTarget string
	var logLevel string

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&config, "config", DefaultConfigFile, "YAML configuration file.")
	flags.Var(&exec, "exec", "Execute a watcher and exit. May be provided multiple times.")
	flags.Var(&etcd, "etcd", "The URI of etcd. May be provided multiple times.")
	flags.StringVar(&prefix, "prefix", "", "A prefix to prepend to all key paths.")
	flags.StringVar(&logTarget, "log-target", "", "The target to log to.")
	flags.StringVar(&logLevel, "log-level", "", "The level of logs to log.")
	flags.Parse(args[1:])

	return &Options{
		Config:    config,
		Exec:      []string(exec),
		Etcd:      []string(etcd),
		LogTarget: logTarget,
		LogLevel:  logLevel,
	}
}
