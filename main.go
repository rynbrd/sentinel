package main

import (
	"fmt"
	"gopkg.in/BlueDragonX/simplelog.v1"
	"os"
	"os/signal"
	"syscall"
)

var (
	logger *simplelog.Logger
)

// Run the app.
func main() {
	// initialize logging
	var err error
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

	// replace the logger
	loggerDest := 0
	if cfg.Logging.Syslog {
		loggerDest |= simplelog.SYSLOG
	}
	if cfg.Logging.Console {
		loggerDest |= simplelog.CONSOLE
	}

	oldLogger := logger
	if logger, err = simplelog.NewLogger(loggerDest, "sentinel"); err != nil {
		oldLogger.Fatal("failed to create logger:", err)
	}
	logger.SetLevel(cfg.Logging.Level)

	// begin startup sequence
	var client *Client
	client, err = cfg.Etcd.CreateClient()
	if err != nil {
		logger.Fatal("failed to create client: %s", err)
	}
	oldLogger.Close()

	manager, err := cfg.Watchers.CreateWatchManager(client)
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
