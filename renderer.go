package main

import (
	"fmt"
	"gopkg.in/BlueDragonX/go-hash.v1"
	"gopkg.in/BlueDragonX/simplelog.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type Template struct {
	Src  string
	Dest string
}

// Render the template to a temporary and return true if the original was changed.
func (t *Template) Render(context map[string]interface{}) (changed bool, err error) {
	// create the destination directory
	dir := filepath.Dir(t.Dest)
	if err = os.MkdirAll(dir, 0777); err != nil {
		return
	}

	// create a temp file to write
	var tmp *os.File
	prefix := fmt.Sprintf(".%s-", filepath.Base(t.Dest))
	if tmp, err = ioutil.TempFile(dir, prefix); err != nil {
		return
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	// render the template to the temp file
	var tpl *template.Template
	name := filepath.Base(t.Src)
	if tpl, err = template.New(name).ParseFiles(t.Src); err != nil {
		return
	}
	if err = tpl.Execute(tmp, context); err != nil {
		return
	}
	tmp.Close()

	// return if the old and new files are the same
	var destHash string
	if destHash, err = hash.File(t.Dest); err != nil {
		return
	}

	var tmpHash string
	if tmpHash, err = hash.File(tmp.Name()); err != nil {
		return
	}

	if destHash == tmpHash {
		return
	}

	// replace the old file with the new one
	if err = os.Rename(tmp.Name(), t.Dest); err == nil {
		changed = true
	}
	return
}

// A renderer generates files from a collection of templates.
type Renderer struct {
	templates []Template
	logger    *simplelog.Logger
}

func NewRenderer(templates []Template, logger *simplelog.Logger) *Renderer {
	item := &Renderer{
		templates,
		logger,
	}
	return item
}

func (renderer *Renderer) Render(context map[string]interface{}) (changed bool, err error) {
	var oneChanged bool
	for _, template := range renderer.templates {
		if oneChanged, err = template.Render(context); err != nil {
			return
		}
		if oneChanged {
			renderer.logger.Debug("template '%s' rendered to '%s'", template.Src, template.Dest)
		} else {
			renderer.logger.Debug("template '%s' did not change", template.Dest)
		}
		changed = changed || oneChanged
	}
	return changed, nil
}
