package main

import "github.com/go-kit/kit/log/term"

func colorFn(keyvals ...interface{}) term.FgBgColor {
	for i := 1; i < len(keyvals); i += 2 {
		if _, ok := keyvals[i].(error); ok {
			return term.FgBgColor{Fg: term.White, Bg: term.Red}
		}
	}
	for i := 0; i < len(keyvals); i += 2 {
		if key := keyvals[i]; key != nil && key == "level" {
			switch keyvals[i+1] {
			case "debug":
				return term.FgBgColor{Fg: term.Gray}
			case "info":
				return term.FgBgColor{Fg: term.Green}
			case "warn":
				return term.FgBgColor{Fg: term.Yellow}
			case "error":
				return term.FgBgColor{Fg: term.Red}
			case "crit":
				return term.FgBgColor{Fg: term.White, Bg: term.Red}
			default:
				continue
			}
		}
	}
	return term.FgBgColor{}
}
