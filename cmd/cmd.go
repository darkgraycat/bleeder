package cmd

import (
	"bleeder/internal/player"
	"fmt"
)

type CmdArgs struct {
	BleedPath string
	CfgPath   string
	Seq       string
	Raw       string
}

type Cmd func(args *CmdArgs) error

// Command to play specified .bleed.toml file
func CmdPlay(args *CmdArgs) error {
	fmt.Printf("[PLAY] %v\n", args)
	cfg, err := LoadConfig(args.CfgPath)
	if err != nil {
		return fmt.Errorf("Config - %v", err)
	}
	bleed, err := LoadBleed(args.BleedPath)
	if err != nil {
		return fmt.Errorf("Bleed - %v", err)
	}
	bleeder, err := NewBleeder(cfg).Bleed(bleed)
	if err != nil {
		return err
	}

	pr, err := bleeder.GetMainIR()
	if err != nil {
		return err
	}

	// TODO: define correct player by config or some
	p := player.NewWAVPlayer(cfg.Audio.SampleRate, cfg.Audio.Channels)
	err = p.Play(pr, 0, pr.Length())
	if err != nil {
		return err
	}
	return nil
}

// Start application in daemon mode listening
func CmdListen(args *CmdArgs) error {
	// TODO
	fmt.Printf("[LISTEN] %v\n", args)
	return nil
}

// Send partial sequence data to play
func CmdSend(args *CmdArgs) error {
	// TODO
	fmt.Printf("[SEND] %v\n", args)
	return nil
}
