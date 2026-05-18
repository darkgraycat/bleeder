package cmd

import (
	"bleeder/internal/player"
	"bleeder/internal/shared/logs"
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
	logs.Info("PLAY %v", args)

	logs.Debug("loading config")
	cfg, err := LoadConfig(args.CfgPath)
	if err != nil {
		return fmt.Errorf("config - %v", err)
	}
	logs.Debug("config loaded")

	logs.Debug("loading bleed")
	bleed, err := LoadBleed(args.BleedPath)
	if err != nil {
		return fmt.Errorf("bleed - %v", err)
	}
	logs.Debug("bleed loaded")

	bleeder, err := NewBleeder(cfg).Bleed(bleed)
	if err != nil {
		return err
	}

	irp, err := bleeder.GenMainIR()
	if err != nil {
		return err
	}

	// TODO: define correct player by config or some
	p := player.NewWAVPlayer(cfg.Audio.SampleRate, cfg.Audio.Channels)
	err = p.Play(irp, 0, irp.Length())
	if err != nil {
		return err
	}
	return nil
}

// Start application in daemon mode listening
func CmdListen(args *CmdArgs) error {
	// TODO
	logs.Info("LISTEN %v", args)
	return nil
}

// Send partial sequence data to play
func CmdSend(args *CmdArgs) error {
	// TODO
	logs.Info("SEND %v", args)
	return nil
}
