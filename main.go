package main

import (
	"bleeder/cmd"
	"log"
	"os"
)

var handlers = map[string]cmd.Cmd{
	"play": cmd.CmdPlay,
	"live": cmd.CmdLive,
	"info": cmd.CmdInfo,
	"help": cmd.CmdHelp,
}

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatalln("[ERROR]: invalid number of arguments")
	}

	handler, ok := handlers[os.Args[1]]
	if !ok {
		log.Fatalf("[ERROR]: available commands - play, live, info, help")
	}

	err := handler(os.Args[2:])
	if err != nil {
		log.Fatalln(err)
	}
}
