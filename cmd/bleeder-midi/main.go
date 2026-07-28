package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	// TODO: Implement MIDI renderer
	// Read serialized IR from stdin
	// Output MIDI events to stdout

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// TODO: Parse IR instruction
		// TODO: Generate MIDI events
		fmt.Fprintf(os.Stderr, "[MIDI] Processing: %s\n", line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("[ERROR]: %v\n", err)
	}
}
