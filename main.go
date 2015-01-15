package main

import (
	"gopkg.in/BlueDragonX/go-log.v0"
	"os"
	"os/signal"
	"syscall"
)

var (
	logger *log.Logger
)

// Run the app.
func main() {
	// initialize logging
	var err error
	logger = log.NewOrExit()

	// load configuration
	var cfg Config
	if cfg, err = LoadConfig(); err != nil {
		logger.Fatalf("error parsing config: %s", err)
	}
	if errs := cfg.Validate(); len(errs) != 0 {
		logger.Error("config file is invalid:")
		for _, err = range errs {
			logger.Errorf("  %s", err)
		}
		logger.Fatal("could not process config file")
	}

	// replace the logger
	oldLogger := logger
	logTarget := log.Target(cfg.Logging.Target)
	logLevel := log.Level(cfg.Logging.Level)
	if logger, err = log.New(logTarget, logLevel); err != nil {
		oldLogger.Fatalf("failed to create logger:", err)
	}
	oldLogger.Close()

	// begin startup sequence
	var client *Client
	client, err = cfg.Etcd.CreateClient()
	if err != nil {
		logger.Fatalf("failed to create client: %s", err)
	}

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
	logger.Info("executing watchers")
	if err = manager.Execute(exec); err != nil {
		logger.Fatalf("failed to execute: %s", err)
	}

	if len(cfg.Exec) == 0 {
		// run
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		logger.Info("starting")
		manager.Start()
		logger.Info("started")
		<-signals
		logger.Info("stopping")
		manager.Stop()
		logger.Info("stopped")
	}
}
