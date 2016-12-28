package utils

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mchudgins/certMgr/pkg/certMgr"
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

func StartUpMessage(cfg certMgr.AppConfig) {
	log.WithFields(log.Fields{
		"version":       version,
		"buildTime":     buildTime,
		"builder":       builder,
		"buildNum":      buildNum,
		"goVersion":     goversion,
		"configuration": fmt.Sprintf("%#v", cfg),
	}).Info("certMgr startup")
}
