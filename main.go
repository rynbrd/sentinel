package main

import (
	"gopkg.in/BlueDragonX/go-log.v0"
	"gopkg.in/BlueDragonX/go-settings.v0"
	"os"
	"os/signal"
	"syscall"
)

var logger *log.Logger = log.NewOrExit()

// Do basic configuration and return the config object.
func configure() *settings.Settings {
	options := ParseOptionsOrExit(os.Args)
	config := settings.LoadOrExit(options.Config)

	// set config values from cli options
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

	// configure the logger
	if logTarget, err := config.String("logging.target"); err == nil {
		if logTargetObj, err := log.NewTarget(logTarget); err == nil {
			logger.SetTarget(logTargetObj)
		} else {
			Fatalf("%s\n", err)
		}
	}
	if logLevel, err := config.String("logging.level"); err == nil {
		logger.SetLevel(log.NewLevel(logLevel))
	}

	// normalize etcd configuration
	etcdURI := config.StringDflt("etcd.uri", "")
	etcdURIs := config.StringArrayDflt("etcd.uris", []string{})
	if len(etcdURIs) == 0 && etcdURI != "" {
		rawURIs := make([]interface{}, 1)
		rawURIs[0] = etcdURI
		config.Set("etcd.uris", rawURIs)
	}

	// propogate default prefix to watchers
	if prefix := config.StringDflt("etcd.prefix", ""); prefix != "" {
		for _, watcher := range config.ObjectMapDflt("watchers", map[string]*settings.Settings{}) {
			if watcherPrefix, err := watcher.String("prefix"); err == nil {
				watcher.Set("prefix", JoinPath(prefix, watcherPrefix))
			} else {
				watcher.Set("prefix", prefix)
			}
		}
	}

	return config
}

// Run the app.
func main() {
	config := configure()
	sentinel := ConfigSentinel(config)
	stop := make(chan bool)
	defer close(stop)

	logger.Info("starting sentinel")
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-signals
		stop <- true
		logger.Infof("got signal %s, stopping", sig)
	}()

	if sentinel.Client.Wait(stop) {
		sentinel.Run(stop)
	}
}
