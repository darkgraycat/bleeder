package cmd

import (
	"bleeder/internal/core"
	"bufio"
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn net.Conn, ctx *CmdContext) error {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		args := strings.Fields(scanner.Text())
		if len(args) == 0 {
			continue
		}
		cmd := strings.ToUpper(args[0])
		switch cmd {
		case "PLAY":
			seqName := getArg(args, 1, core.MAIN_NAME)
			err := ctx.Play(seqName, "")
			if err != nil {
				fmt.Fprintf(conn, "ERR %v\n", err)
				continue
			}

		case "STOP":
			err := ctx.Stop()
			if err != nil {
				fmt.Fprintf(conn, "ERR %v\n", err)
			} else {
				fmt.Fprintf(conn, "OK stopped\n")
			}

		case "INFO":
			info := ctx.Info()
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
