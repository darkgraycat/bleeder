package cmd

import (
	"bleeder/internal/bleeder"
	"bleeder/internal/player"
	"flag"
	"fmt"
	"log"
	"net"
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

func CmdLive(args []string) error {
	fs := flag.NewFlagSet("live", flag.ExitOnError)
	cfgPath := fs.String("config", defaultConfigPath(), "config file path")
	fs.Parse(args)

	// bleedPath := fs.Arg(0)
	// if bleedPath == "" {
	// 	return fmt.Errorf("usage: bleeder live [flags] <file>")
	// }

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// bleed, err := bleeder.LoadBleed(bleedPath)
	// if err != nil {
	// 	return fmt.Errorf("loading bleed: %w", err)
	// }

	port := fmt.Sprintf(":%d", cfg.Live.Port)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("tcp server: %w", err)
	}
	defer listener.Close()

	log.Printf("[INIT:LIVE] Listening on %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("[ERROR] ", err)
			continue
		}

		if err := handleConnection(conn); err != nil {
			log.Println("[ERROR] ", err)
		}
	}
}

func CmdInfo(args []string) error {
	// fs := flag.NewFlagSet("info")
	return fmt.Errorf("info is not implemented yet")
}

func CmdHelp(args []string) error {
	return fmt.Errorf("help is not implemented yet")
}
