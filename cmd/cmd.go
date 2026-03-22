package cmd

import (
	"bleeder/internal/utils"
	"fmt"
)

type Cmd func(args *utils.Args) error

// Command to play specified .bleed.toml file
func CmdPlay(args *utils.Args) error {
	fmt.Printf("[PLAY] %v\n", args)
	cfg, err := LoadConfig(args.At(2))
	bleed, err := LoadBleed(args.At(2))
	if err != nil {
		return err
	}
	fmt.Printf("Cfg %v\nBleed %v\n", cfg, bleed)

	bleeder, err := NewBleeder(cfg).Bleed(bleed)
	if err != nil {
		return err
	}
	fmt.Printf("Bleeder %v\n", bleeder)

	ir, err := bleeder.GetMainIR()
	if err != nil {
		return err
	}
	fmt.Printf("IR %v\n", ir)

	return nil
}

// Start application in daemon mode listening
func CmdServe(args *utils.Args) error {
	fmt.Printf("[SERVE] %v\n", args)
	return nil
}

// Send partial sequence data to play
func CmdSend(args *utils.Args) error {
	fmt.Printf("[SEND] %v\n", args)
	return nil
}
