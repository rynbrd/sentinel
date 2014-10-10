package main

import (
	"fmt"
	"gopkg.in/BlueDragonX/simplelog.v1"
	"os"
	"os/signal"
	"syscall"
)

// Run the app.
func main() {
	// initialize logging
	var err error
	var logger *simplelog.Logger
	if logger, err = simplelog.NewLogger(simplelog.CONSOLE, "sentinel"); err != nil {
		fmt.Println("failed to create logger:", err)
		os.Exit(1)
	}

	// load configuration
	var cfg Config
	if cfg, err = LoadConfig(); err != nil {
		logger.Fatal("error parsing config: %s", err)
	}
	if errs := cfg.Validate(); len(errs) != 0 {
		logger.Error("config file is invalid:")
		for _, err = range errs {
			logger.Error("  %s", err)
		}
		logger.Fatal("could not process config file")
	}

	// begin startup sequence
	if cfg.Logging.Syslog {
		if logger, err = simplelog.NewLogger(simplelog.SYSLOG, "sentinel"); err != nil {
			fmt.Println("failed to create syslog logger:", err)
			os.Exit(1)
		}
	}
	logger.SetLevel(cfg.Logging.Level)

	var client *Client
	client, err = cfg.Etcd.CreateClient(logger)
	if err != nil {
		logger.Fatal("failed to create client: %s", err)
	}

	manager, err := cfg.Watchers.CreateWatchManager(client, logger)
	if err != nil {
		logger.Fatal("failed to create watch manager")
	}

	var exec []string
	if len(cfg.Exec) > 0 {
		exec = cfg.Exec
	} else {
		exec = []string{}
		for name := range manager.Watchers {
			exec = append(exec, name)
		}
	}

	// exec
	logger.Notice("executing watchers")
	if err = manager.Execute(exec); err != nil {
		logger.Fatal("failed to execute: %s", err)
	}

	if len(cfg.Exec) == 0 {
		// run
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		logger.Notice("starting")
		manager.Start()
		logger.Notice("started")
		<-signals
		logger.Notice("stopping")
		manager.Stop()
		logger.Notice("stopped")
	}
}
