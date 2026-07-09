package cmd

import (
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

		switch args[0] {
		case "PLAY":
			fmt.Fprintln(conn, "OK")

		case "STOP":
			fmt.Fprintln(conn, "OK")

		case "INFO":
			fmt.Fprintln(conn, "playing")

		default:
			fmt.Fprintf(conn, "ERR unknown command: %q\n", args[0])
		}
	}

	return scanner.Err()
}
