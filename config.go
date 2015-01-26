package main

import (
	"gopkg.in/BlueDragonX/go-settings.v0"
)

func ConfigTemplates(configs []*settings.Settings) []Template {
	templates := make([]Template, len(configs))
	for n, config := range configs {
		src := config.StringDflt("src", "")
		dest := config.StringDflt("dest", "")
		if src == "" {
			logger.Fatalf("config '%s.src' is missing", config.Key)
		}
		if dest == "" {
			logger.Fatalf("config '%s.dest' is missing", config.Key)
		}
		templates[n] = Template{Src: src, Dest: dest}
	}
	return templates
}

func ConfigSentinel(config *settings.Settings) *Sentinel {
	client, err := NewEtcdClient(config.ObjectDflt("etcd", &settings.Settings{}))
	if err != nil {
		logger.Fatalf("failed to create client: %s", err)
	}

	sentinel := Sentinel{Client: client}

	watchers, err := config.ObjectMap("watchers")
	if err != nil {
		logger.Fatal("config 'watchers' is invalid")
	}
	if len(watchers) == 0 {
		logger.Fatal("config 'watchers' is missing")
	}

	for name, watcher := range watchers {
		prefix := watcher.StringDflt("prefix", "")
		watch := ResolvePaths(prefix, watcher.StringArrayDflt("watch", []string{}))
		context := ResolvePaths(prefix, watcher.StringArrayDflt("context", []string{}))

		templatesConfig, err := watcher.ObjectArray("templates")
		if err != nil {
			logger.Fatalf("config '%s.templates' is invalid", watcher.Key)
		}
		templates := ConfigTemplates(templatesConfig)

		var command []string
		if cmdStr, err := watcher.String("command"); err == nil {
			command = []string{"bash", "-c", cmdStr}
		} else if cmdArray, err := watcher.StringArray("command"); err == nil {
			command = cmdArray
		}

		if len(templates) == 0 && len(command) == 0 {
			logger.Fatalf("watcher %s templates and command both missing")
		}

		executor := &TemplateExecutor{
			name:      name,
			prefix:    prefix,
			context:   context,
			Templates: templates,
			Command:   command,
		}

		sentinel.Add(watch, executor)
	}

	return &sentinel
}
