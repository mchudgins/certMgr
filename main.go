// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"os"

	"github.com/mchudgins/golang-service-starter/cmd"
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

func main() {
	log.Printf("golang-frontend-starter: version %s; buildTime: %s; built by: %s; buildNum: %s; (%s)",
		version, buildTime, builder, buildNum, goversion)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting app on host %s...", hostname)

	cmd.Execute()
}
