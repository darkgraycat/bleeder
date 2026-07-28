package cmd

import (
	"bleeder/internal/core"
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Cmd func(args []string) error

func CmdPlay(args []string) error {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	cfgPath := fs.String("cfg", defaultConfigPath(), "config file path")
	seqName := fs.String("seq", core.MAIN_NAME, "sequence to play")
	seqVars := fs.String("vars", "", "sequence variables")
	fs.Parse(args)

	bleedPath := fs.Arg(0)
	if bleedPath == "" {
		return fmt.Errorf("usage: bleeder play [flags] <file>")
	}

	_, err := LoadConfig(*cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	bleed, err := core.LoadBleed(bleedPath)
	if err != nil {
		return fmt.Errorf("loading bleed: %w", err)
	}

	bctx := core.NewBleedContext(bleed)
	return bctx.Render(*seqName, *seqVars, os.Stdout)
}

func CmdLive(args []string) error {
	fs := flag.NewFlagSet("live", flag.ExitOnError)
	cfgPath := fs.String("cfg", defaultConfigPath(), "config file path")
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

	bctx := core.NewBleedContext(bleed)
	go bctx.Run(os.Stdout)

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
		go handleConnection(conn, bctx)
	}
}

func CmdInfo(args []string) error {
	return fmt.Errorf("info is not implemented yet")
}

func CmdHelp(args []string) error {
	return fmt.Errorf("help is not implemented yet")
}

func handleConnection(conn net.Conn, bctx *core.BleedContext) error {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	fmt.Fprintln(conn, "READY")
	for scanner.Scan() {
		args := strings.Fields(scanner.Text())
		if len(args) == 0 {
			continue
		}
		var err error
		var res string
		cmd := strings.ToUpper(args[0])
		switch cmd {
		case "PLAY":
			seqName := getArg(args, 1, core.MAIN_NAME)
			err = bctx.Play(seqName, "")
			res = "playing " + seqName
		case "STOP":
			err = bctx.Stop()
			res = "stopped"
		case "SYNC":
			err = bctx.Sync()
			res = "synced"
		case "INFO":
			err = nil
			res = bctx.Info()
		default:
			log.Printf("[TCP] Unknown command: %q\n", cmd)
			fmt.Fprintf(conn, "ERR unknown command: %q\n", cmd)
			continue
		}
		if err != nil {
			log.Printf("[TCP] %s error: %v\n", cmd, err)
			fmt.Fprintf(conn, "ERR %v\n", err)
		} else {
			log.Printf("[TCP] %s response: %s\n", cmd, res)
			fmt.Fprintf(conn, "OK %s\n", res)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("[TCP] Connection error: %v\n", err)
		return err
	}
	return nil
}

func getArg(args []string, idx int, fallback string) string {
	if idx >= len(args) || args[idx] == "" {
		return fallback
	}
	return args[idx]
}
