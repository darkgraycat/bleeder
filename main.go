package main

import (
	"bleeder/cmd"
	// "bleeder/experiments"
	"bleeder/internal/shared"
	"fmt"
	"os"
)

func main() {
	// experiments.Run()
	// return

	// parse CLI args and define which cmd to use
	args := shared.NewArgs(os.Args)
	cmds := map[string]cmd.Cmd{
		"play":  cmd.CmdPlay,
		"serve": cmd.CmdServe,
		"send":  cmd.CmdSend,
	}

	fmt.Printf("Args P: %v, F: %v\n", args.Positional, args.Flags)

	if args.Length() < 2 {
		fmt.Fprintln(os.Stderr, "Mode is not specified")
		os.Exit(1)
	}

	cmd, ok := cmds[args.At(1)]
	if !ok {
		fmt.Fprintln(os.Stderr, "Unknown mode")
		os.Exit(1)
	}

	// run selected cmd
	err := cmd(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Runtime error", err)
		os.Exit(1)
	}
}
