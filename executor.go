package main

import (
	"os/exec"
	"strings"
)

// An Executor is responsible for performing template rendering and command
// execution for a single watcher.
type Executor struct {
	Name      string
	Templates []Template
	Command   []string
}

// Render the templates. Return true if any templates changed.
func (ex *Executor) render(context map[string]interface{}) (changed bool, err error) {
	var oneChanged bool
	for _, tpl := range ex.Templates {
		if oneChanged, err = tpl.Render(context); err != nil {
			return
		}
		if oneChanged {
			logger.Debugf("%s tpl: rendered '%s' -> '%s'", ex.Name, tpl.Src, tpl.Dest)
		} else {
			logger.Debugf("%s tpl: no change to '%s'", ex.Name, tpl.Dest)
		}
		changed = changed || oneChanged
	}
	return changed, nil
}

// Run the command.
func (ex *Executor) run() error {
	if len(ex.Command) == 0 {
		logger.Debugf("%s has no command, skipping", ex.Name)
		return nil
	}

	cmdName := ex.Command[0]
	cmdArgs := ex.Command[1:]
	command := exec.Command(cmdName, cmdArgs...)

	out, err := command.CombinedOutput()
	if err == nil {
		logger.Debugf("%s cmd ran", ex.Name)
	} else {
		logger.Errorf("%s cmd failed: %s", ex.Name, err)
	}
	outStr := string(out)
	if outStr != "" {
		lines := strings.Split(outStr, "\n")
		for _, line := range lines {
			if err == nil {
				logger.Debugf("%s cmd: %s", ex.Name, line)
			} else {
				logger.Errorf("%s cmd: %s", ex.Name, line)
			}
		}
	}
	return err
}

// Render the templates and execute the command. The command will be executed
// if one of the template destinations changes or no templates are present in
// the Executor. The `context` is passed to render the templates.
func (ex *Executor) Execute(context map[string]interface{}) error {
	var err error
	run := true
	if len(ex.Templates) > 0 {
		run, err = ex.render(context)
	}
	if len(ex.Command) > 0 && err == nil && run {
		err = ex.run()
	}
	return err
}
