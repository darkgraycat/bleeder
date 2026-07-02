package main

import (
	"bleeder/cmd"
	"fmt"
	"os"
)

var handlers = map[string]cmd.Cmd{
	"play":   cmd.CmdPlay,
	"listen": cmd.CmdListen,
	"reload": cmd.CmdReload,
	"stop":   cmd.CmdStop,
	"status": cmd.CmdStatus,
	"help":   cmd.CmdHelp,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "TODO: print usage")
		os.Exit(1)
	}

	handler, ok := handlers[os.Args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, "TODO: print usage")
		os.Exit(1)
	}

	err := handler(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
