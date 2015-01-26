package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// An Executor is responsible for executed to perform some action when a
// watched key is changed.
type Executor interface {
	// Return the unique name of the executor.
	Name() string

	// Called to run the executor's actions.
	Execute(client Client) error
}

// A Executor performs template rendering. It will optionally execute a command
// when one or more changes are made by the templating system. If no templates
// are provided the command will always be executed.
type TemplateExecutor struct {
	name      string
	Prefix    string
	Context   []string
	Templates []Template
	Command   []string
}

// Render the templates. Return true if any templates changed.
func (ex *TemplateExecutor) render(context map[string]interface{}) (changed bool, err error) {
	var oneChanged bool
	if ex.Templates == nil {
		return
	}

	for _, tpl := range ex.Templates {
		if oneChanged, err = tpl.Render(context); err != nil {
			return
		}
		if oneChanged {
			logger.Debugf("%s: rendered '%s' -> '%s'", ex.Name, tpl.Src, tpl.Dest)
		} else {
			logger.Debugf("%s: no change to '%s'", ex.Name, tpl.Dest)
		}
		changed = changed || oneChanged
	}
	return
}

// Run the command.
func (ex *TemplateExecutor) run() error {
	if len(ex.Command) == 0 {
		logger.Debugf("%s: command not set", ex.Name)
		return nil
	}

	cmdName := ex.Command[0]
	cmdArgs := ex.Command[1:]
	command := exec.Command(cmdName, cmdArgs...)

	out, err := command.CombinedOutput()
	if err == nil {
		logger.Debugf("%s: command ran", ex.Name)
	} else {
		logger.Errorf("%s: command failed: %s", ex.Name, err)
	}
	outStr := string(out)
	if outStr != "" {
		lines := strings.Split(outStr, "\n")
		for _, line := range lines {
			if err == nil {
				logger.Debugf("out: %s", line)
			} else {
				logger.Errorf("out: %s", line)
			}
		}
	}
	return err
}

// Return the unique name of the executor.
func (ex *TemplateExecutor) Name() string {
	return ex.name
}

// Render the templates using the context retrieved from the provided `client`
// and execute the command. The command will be executed if one of the template
// destinations changes or no templates are present in the Watcher.
func (ex *TemplateExecutor) Execute(client Client) error {
	var err error
	var context map[string]interface{}

	if ex.Context == nil || len(ex.Context) == 0 {
		context = map[string]interface{}{}
	} else if context, err = client.Get(ex.Context); err != nil {
		logger.Errorf("%s: context failed: %s", ex.Name, err)
		return err
	} else {
		for _, key := range strings.Split(ex.Prefix, "/") {
			next, ok := context[key]
			if !ok {
				return fmt.Errorf("%s: context %s is invalid", ex.Name, ex.Prefix)
			}
			context, ok = next.(map[string]interface{})
			if !ok {
				return fmt.Errorf("%s: context %s is invalid", ex.Name, ex.Prefix)
			}
		}
	}
	logger.Debugf("%s: got context: %v", ex.Name, context)

	run := true
	if len(ex.Templates) > 0 {
		run, err = ex.render(context)
	}
	if len(ex.Command) > 0 && err == nil && run {
		err = ex.run()
	}
	return err
}
