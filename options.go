package main

import (
	"flag"
	"strings"
)

// A string array option capable of being appended to.
type stringsOpt []string

func (strs *stringsOpt) String() string {
	return strings.Join(*strs, ",")
}

func (strs *stringsOpt) Set(value string) error {
	*strs = append(*strs, strings.Split(value, ",")...)
	return nil
}

type Options struct {
	Config    string
	Exec      []string
	Etcd      []string
	LogTarget string
	LogLevel  string
}

// Parse commandline options. Exit on failure.
func ParseOptionsOrExit(args []string) *Options {
	var config string
	var exec stringsOpt
	var etcd stringsOpt
	var logTarget string
	var logLevel string

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&config, "config", "config.yml", "YAML configuration file.")
	flags.Var(&exec, "exec", "Execute a watcher and exit. May be provided multiple times.")
	flags.Var(&etcd, "etcd", "The URI of etcd. May be provided multiple times.")
	flags.StringVar(&logTarget, "log-target", "stderr", "The target to log to.")
	flags.StringVar(&logLevel, "log-level", "info", "The level of logs to log.")
	flags.Parse(args[1:])

	return &Options{
		Config:    config,
		Exec:      []string(exec),
		Etcd:      []string(etcd),
		LogTarget: logTarget,
		LogLevel:  logLevel,
	}
}
