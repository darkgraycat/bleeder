package main

import (
	"bleeder/cmd"
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
	// parse CLI flags
	cmdArgs := cmd.CmdArgs{}
	flag.StringVar(&cmdArgs.BleedPath, "bleed", "", "")
	flag.StringVar(&cmdArgs.CfgPath, "cfg", "config.toml", "")
	flag.StringVar(&cmdArgs.Seq, "seq", "", "")
	flag.StringVar(&cmdArgs.Raw, "raw", "", "")
	flag.Parse()

	// define which cmd to use
	mode := flag.CommandLine.Arg(0)
	fmt.Printf("MODE - %s\n", mode)
	fmt.Printf("Cmd Args: %v\n", cmdArgs)
	exec, ok := cmds[mode]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", mode)
		os.Exit(1)
	}

	// run selected cmd
	err := exec(&cmdArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Runtime error", err)
		os.Exit(1)
	}
}
