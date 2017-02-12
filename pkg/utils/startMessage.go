package utils

import (
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
)

var (
	// boilerplate variables for good SDLC hygiene.  These are auto-magically
	// injected by the Makefile & linker working together.
	version   string
	buildTime string
	builder   string
	buildNum  string
	goversion string
)

func StartUpMessage(cfg interface{}) {
	var appName string

	appName = path.Base(os.Args[0])

	log.WithFields(log.Fields{
		"version":       version,
		"buildTime":     buildTime,
		"builder":       builder,
		"buildNum":      buildNum,
		"goVersion":     goversion,
		"configuration": fmt.Sprintf("%#v", cfg),
	}).Infof("%s startup", appName)
}
