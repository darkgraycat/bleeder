package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	// TODO: Implement IRP (IR Print) renderer
	// Read serialized IR from stdin
	// Output formatted text to stdout

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// TODO: Parse IR instruction
		// TODO: Format as readable text/tabs
		fmt.Println(line) // Pass-through for now
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("[ERROR]: %v\n", err)
	}
}
