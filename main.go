package main

import (
	"fmt"
	"gopkg.in/BlueDragonX/go-log.v0"
	"gopkg.in/BlueDragonX/go-settings.v0"
	"os"
	"os/signal"
	"syscall"
)

var (
	logger *log.Logger
	config *settings.Settings
)

func configure() {
	options := ParseOptionsOrExit(os.Args)
	config = settings.LoadOrExit(options.Config)
	config.Set("exec", options.Exec)

	if len(options.Etcd) > 0 {
		config.Set("etcd.uris", options.Etcd)
	}

	if options.LogTarget != "" {
		config.Set("logging.target", options.LogTarget)
	}

	if options.LogLevel != "" {
		config.Set("logging.level", options.LogLevel)
	}

	logOpts := []log.Option{}
	if logTarget, err := config.String("logging.target"); err == nil {
		logOpts = append(logOpts, log.Target(logTarget))
	}
	if logLevel, err := config.String("logging.level"); err == nil {
		logOpts = append(logOpts, log.Level(logLevel))
	}
	logger = log.NewOrExit(logOpts...)

	etcdURI := config.StringDflt("etcd.uri", "")
	etcdURIs := config.StringArrayDflt("etcd.uris", []string{})
	if len(etcdURIs) == 0 && etcdURI != "" {
		rawURIs := make([]interface{}, 1)
		rawURIs[0] = etcdURI
		config.Set("etcd.uris", rawURIs)
	}
}

// Run the app.
func main() {
	configure()

	// begin startup sequence
	client, err := NewClient(config.ObjectDflt("etcd", &settings.Settings{}))
	if err != nil {
		logger.Fatalf("failed to create client: %s", err)
	}

	fmt.Printf("Etcd Prefix: %s\n", client.prefix)
	fmt.Printf("Etcd Client: %v\n", client.client)

	watchersConfig := config.ObjectMapDflt("watchers", map[string]*settings.Settings{})
	watchers := make([]*Watcher, 0, len(watchersConfig))
	for _, watcherConfig := range watchersConfig {
		if watcher, err := NewWatcher(client, watcherConfig); err == nil {
			watchers = append(watchers, watcher)
			fmt.Printf("Added watcher %s\n", watcher.Name())
		} else {
			logger.Fatal(err.Error())
		}
	}

	manager := NewWatchManager(client, watchers)
	fmt.Printf("WatchManager: %v\n", manager)

	// exec
	logger.Info("executing watchers")
	exec := config.StringArrayDflt("exec", []string{})
	if err = manager.Execute(exec); err != nil {
		logger.Fatalf("failed to execute: %s", err)
	}

	if len(exec) == 0 {
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
