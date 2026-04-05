package cmd

import (
	"bleeder/internal/shared"
	"fmt"
)

type Cmd func(args *shared.Args) error

// Command to play specified .bleed.toml file
func CmdPlay(args *shared.Args) error {
	fmt.Printf("[PLAY] %v\n", args)
	cfg, err := LoadConfig(args.At(2))
	bleed, err := LoadBleed(args.At(2))
	if err != nil {
		return err
	}

	bleeder, err := NewBleeder(cfg).Bleed(bleed)
	if err != nil {
		return err
	}


	fmt.Printf("[PLAY] GetMainIR()\n")
	ir, err := bleeder.GetMainIR()
	if err != nil {
		return err
	}
	// TODO: render IR using player
	fmt.Printf("IR %v\n", ir)

	return nil
}

// Start application in daemon mode listening
func CmdServe(args *shared.Args) error {
	fmt.Printf("[SERVE] %v\n", args)
	return nil
}

// Send partial sequence data to play
func CmdSend(args *shared.Args) error {
	fmt.Printf("[SEND] %v\n", args)
	return nil
}
