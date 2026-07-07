package cmd

import (
	"bleeder/internal/bleeder"
	"bleeder/internal/player"
	"flag"
	"fmt"
)

type Cmd func(args []string) error

func CmdPlay(args []string) error {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	cfgPath := fs.String("config", defaultConfigPath(), "config file path")
	seqName := fs.String("seq", bleeder.MAIN_NAME, "sequence to play")
	seqVars := fs.String("vars", "", "sequence variables")
	fs.Parse(args)

	bleedPath := fs.Arg(0)
	if bleedPath == "" {
		return fmt.Errorf("usage: bleeder play [flags] <file>")
	}

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	bleed, err := bleeder.LoadBleed(bleedPath)
	if err != nil {
		return fmt.Errorf("loading bleed: %w", err)
	}

	b := bleeder.NewBleeder(bleed)
	irp, err := b.GenSeqIR(*seqName, *seqVars)
	if err != nil {
		return fmt.Errorf("generating %q %q: %w", *seqName, *seqVars, err)
	}

	// TODO: define correct player by config or some
	p := player.NewWAVPlayer(cfg.Audio.SampleRate, cfg.Audio.Channels)
	err = p.Play(irp, 0, irp.Length())
	if err != nil {
		return fmt.Errorf("playing %q %q: %w", *seqName, *seqVars, err)
	}
	return nil
}

func CmdStop(args []string) error {
	return fmt.Errorf("stop is not implemented yet")
}

func CmdListen(args []string) error {
	return fmt.Errorf("listen is not implemented yet")
}

func CmdReload(args []string) error {
	return fmt.Errorf("reload is not implemented yet")
}

func CmdStatus(args []string) error {
	return fmt.Errorf("status is not implemented yet")
}

func CmdInfo(args []string) error {
	// fs := flag.NewFlagSet("info")
	return fmt.Errorf("info is not implemented yet")
}

func CmdHelp(args []string) error {
	return fmt.Errorf("help is not implemented yet")
}
