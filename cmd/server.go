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
		cmd := strings.TrimSpace(scanner.Text())

		switch cmd {
		case "":
			continue

		case "PLAY":
			fmt.Fprintln(conn, "OK")

		case "STOP":
			fmt.Fprintln(conn, "OK")

		case "INFO":
			fmt.Fprintln(conn, "playing")

		default:
			fmt.Fprintf(conn, "ERR unknown command: %s\n", cmd)
		}
	}

	return scanner.Err()
}
