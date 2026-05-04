package main

import (
	"bleeder/cmd"
	"bleeder/experiments"
	"bleeder/internal/shared/logs"
	"flag"
	"fmt"
	"os"
)

var cmds = map[string]cmd.Cmd{
	"play":   cmd.CmdPlay,
	"listen": cmd.CmdListen,
	"send":   cmd.CmdSend,
}

func main() {
	logs.SetLogLevel(1)

	// TODO remove
	experiments.Run()
	if 2 > 1 {
		logs.Info("Exit")
		return
	}

	// parse CLI flags
	cmdMode := ""
	cmdArgs := cmd.CmdArgs{}
	flag.StringVar(&cmdMode, "mode", "play", "")
	flag.StringVar(&cmdArgs.BleedPath, "bleed", "", "")
	flag.StringVar(&cmdArgs.CfgPath, "cfg", "config.toml", "")
	flag.StringVar(&cmdArgs.Seq, "seq", "", "")
	flag.StringVar(&cmdArgs.Raw, "raw", "", "")
	flag.Parse()

	// define which cmd to use
	exec, ok := cmds[cmdMode]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", cmdMode)
		os.Exit(1)
	}

	// run selected cmd
	err := exec(&cmdArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in", err)
		os.Exit(1)
	}
}
