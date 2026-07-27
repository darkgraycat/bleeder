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
			bctx.Play(seqName, "")
			fmt.Fprintf(conn, "OK playing\n")

		case "STOP":
			bctx.Stop()
			fmt.Fprintf(conn, "OK stopped\n")

		case "INFO":
			info := bctx.Info()
			fmt.Fprintf(conn, "%s\n", info)

		default:
			fmt.Fprintf(conn, "ERR unknown command: %q\n", cmd)
		}
	}

	return scanner.Err()
}

func getArg(args []string, idx int, fallback string) string {
	if idx >= len(args) || args[idx] == "" {
		return fallback
	}
	return args[idx]
}
