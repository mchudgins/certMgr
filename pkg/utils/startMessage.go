package utils

import (
	"log"
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
	log.Printf("certMgr: version %s; buildTime: %s; built by: %s; buildNum: %s; (%s)",
		version, buildTime, builder, buildNum, goversion)
}
