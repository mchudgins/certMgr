package utils

import (
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

func StartUpMessage() {
	log.WithFields(log.Fields{
		"version":   version,
		"buildTime": buildTime,
		"builder":   builder,
		"buildNum":  buildNum,
		"goVersion": goversion,
	}).Info("certMgr startup")
}
