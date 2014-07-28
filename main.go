package main

import (
	"fmt"
	"gopkg.in/BlueDragonX/simplelog.v1"
	"os"
	"os/signal"
	"strings"
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

	if len(cfg.Exec) > 0 {
		// exec
		logger.Notice("executing %s", strings.Join(cfg.Exec, ", "))
		if err = manager.Execute(cfg.Exec); err != nil {
			logger.Fatal("failed to execute: %s", err)
		}
	} else {
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
