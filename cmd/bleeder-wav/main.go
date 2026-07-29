package main

import (
	"bleeder/internal/renderer"
	"flag"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	fs := flag.NewFlagSet("wav-renderer", flag.ExitOnError)
	sr := fs.Int("sr", 44010, "sample rate")
	ch := fs.Int("ch", 1, "number of channels")
	fs.Parse(os.Args[1:])

	log.Printf("SampleRate=%d Channels=%d", *sr, *ch)

	// TODO: Implement WAV renderer
	// Read serialized IR from stdin
	// Output WAV samples to stdout

	renderer := renderer.NewWAVRenderer(*sr, *ch)

	err := renderer.Render(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("[ERROR]: %v\n", err)
	}
}
