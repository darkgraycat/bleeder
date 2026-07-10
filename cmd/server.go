package cmd

import (
	"bleeder/internal/bleeder"
	"bufio"
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) error {
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
			seqName := getArg(args, 1, bleeder.MAIN_NAME)
			fmt.Fprintf(conn, "Playing %s", seqName)

		case "STOP":
			seqName := getArg(args, 1, bleeder.MAIN_NAME)
			fmt.Fprintf(conn, "Stopping %s", seqName)

		case "INFO":
			seqName := getArg(args, 1, bleeder.MAIN_NAME)
			fmt.Fprintf(conn, "Info %s", seqName)

		// TODO: define what actual commands is needed for
		// live-coding and VSCode extension to operate with

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
