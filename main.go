package main

import (
	"bleeder/cmd"
	"bleeder/internal/utils"
	"fmt"
	"os"
)

func main() {
	args := utils.NewArgs(os.Args)

	modes := map[string]cmd.Cmd{
		"play":  cmd.CmdPlay,
		"send":  cmd.CmdSend,
		"serve": cmd.CmdServe,
	}

	mode, ok := modes[args.At(1)]
	if !ok {
		fmt.Fprintln(os.Stderr, "Unknown mode")
		os.Exit(1)
	}

	err := mode(args)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Runtime error", err)
		os.Exit(1)
	}
}
