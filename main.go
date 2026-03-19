package main

import (
	"bleeder/cmd"
	"bleeder/internal/utils"
	"fmt"
	"os"
)

func main() {
	modes := map[string]cmd.Exec{
		"play":  cmd.ExecPlay,
		"send":  cmd.ExecSend,
		"serve": cmd.ExecServe,
	}
	args := utils.NewArgs(os.Args)

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
