package cmd

import (
	"bleeder/internal/core"
	"flag"
	"fmt"
	"log"
	"net"
)

type Cmd func(args []string) error

func CmdPlay(args []string) error {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	cfgPath := fs.String("config", defaultConfigPath(), "config file path")
	seqName := fs.String("seq", core.MAIN_NAME, "sequence to play")
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

	bleed, err := core.LoadBleed(bleedPath)
	if err != nil {
		return fmt.Errorf("loading bleed: %w", err)
	}

	ctx := NewCmdContext(cfg, bleed)
	return ctx.Play(*seqName, *seqVars)
}

func CmdLive(args []string) error {
	fs := flag.NewFlagSet("live", flag.ExitOnError)
	cfgPath := fs.String("config", defaultConfigPath(), "config file path")
	fs.Parse(args)

	bleedPath := fs.Arg(0)
	if bleedPath == "" {
		return fmt.Errorf("usage: bleeder live [flags] <file>")
	}

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	bleed, err := core.LoadBleed(bleedPath)
	if err != nil {
		return fmt.Errorf("loading bleed: %w", err)
	}

	ctx := NewCmdContext(cfg, bleed)

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
		err = handleConnection(conn, ctx)
		if err != nil {
			log.Println("[ERROR] ", err)
			continue
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
