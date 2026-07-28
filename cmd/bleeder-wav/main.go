package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	// TODO: Implement WAV renderer
	// Read serialized IR from stdin
	// Output WAV samples to stdout

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// TODO: Parse IR instruction
		// TODO: Generate WAV samples
		fmt.Fprintf(os.Stderr, "[WAV] Processing: %s\n", line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("[ERROR]: %v\n", err)
	}
}
