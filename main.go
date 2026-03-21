package main

import (
	"bleeder/cmd"
	"bleeder/internal/utils"
	"fmt"
	"os"
)

func main() {
	// parse CLI args and define which cmd to use
	args := utils.NewArgs(os.Args)
	cmds := map[string]cmd.Cmd{
		"play":  cmd.CmdPlay,
		"serve": cmd.CmdServe,
		"send":  cmd.CmdSend,
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
