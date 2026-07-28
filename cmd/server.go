package cmd

import (
	"bleeder/internal/core"
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func handleConnection(conn net.Conn, bctx *core.BleedContext) error {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	fmt.Fprintln(conn, "READY")
	for scanner.Scan() {
		log.Printf("[TCP] << %s\n", scanner.Text())
		args := strings.Fields(scanner.Text())
		if len(args) == 0 {
			continue
		}
		cmd := strings.ToUpper(args[0])
		switch cmd {
		case "PLAY":
			seqName := getArg(args, 1, core.MAIN_NAME)
			err := bctx.Play(seqName, "")
			if err != nil {
				log.Printf("[TCP] PLAY error: %v\n", err)
				fmt.Fprintf(conn, "ERR %v\n", err)
			} else {
				log.Printf("[TCP] Playing %s\n", seqName)
				fmt.Fprintf(conn, "OK playing %s\n", seqName)
			}

		case "STOP":
			err := bctx.Stop()
			if err != nil {
				log.Printf("[TCP] STOP error: %v\n", err)
				fmt.Fprintf(conn, "ERR %v\n", err)
			} else {
				log.Println("[TCP] Stopped")
				fmt.Fprintf(conn, "OK stopped\n")
			}

		case "SYNC":
			err := bctx.Sync()
			if err != nil {
				log.Printf("[TCP] SYNC error: %v\n", err)
				fmt.Fprintf(conn, "ERR %v\n", err)
			} else {
				log.Println("[TCP] Synced")
				fmt.Fprintf(conn, "OK synced\n")
			}

		case "INFO":
			info := bctx.Info()
			log.Printf("[TCP] Info: %s\n", info)
			fmt.Fprintf(conn, "%s\n", info)

		default:
			log.Printf("[TCP] Unknown command: %q\n", cmd)
			fmt.Fprintf(conn, "ERR unknown command: %q\n", cmd)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[TCP] Connection error: %v\n", err)
	}

	return scanner.Err()
}

func getArg(args []string, idx int, fallback string) string {
	if idx >= len(args) || args[idx] == "" {
		return fallback
	}
	return args[idx]
}
