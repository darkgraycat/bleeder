package main

import (
	"bleeder/cmd"
	"bleeder/internal/utils"
	"fmt"
	"os"
)

func main() {
	args := utils.NewArgs(os.Args)

	mode := args.At(1)

	switch mode {
	case "play":
		cmd.ExecPlay(args, nil)
		cmd.ModePlay{}.Run(args)
	case "send":
		cmd.ModeSend{}.Run(args.Positional)
	default:

	}

	cfg, err := cmd.LoadConfig("config.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config file", err)
		os.Exit(1)
	}

	bleed, err := cmd.LoadBleed("experiments/song_a.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading bleed file", err)
		os.Exit(1)
	}

	err = cmd.Execute(cfg, bleed)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Runtime error", err)
		os.Exit(1)
	}
}
