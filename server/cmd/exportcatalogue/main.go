// Command exportcatalogue writes the pattern catalogue as protojson for the web client.
// Source of truth is internal/layout/patterns.textproto (same file the server embeds).
package main

import (
	"fmt"
	"os"

	lifepb "game_of_life/server/gen/life/v1"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
)

func main() {
	raw, err := os.ReadFile("internal/layout/patterns.textproto")
	if err != nil {
		fmt.Fprintf(os.Stderr, "exportcatalogue: %v\n", err)
		os.Exit(1)
	}
	var catalogue lifepb.Catalogue
	if err := prototext.Unmarshal(raw, &catalogue); err != nil {
		fmt.Fprintf(os.Stderr, "exportcatalogue: %v\n", err)
		os.Exit(1)
	}
	out, err := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: true,
	}.Marshal(&catalogue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "exportcatalogue: %v\n", err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(append(out, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "exportcatalogue: %v\n", err)
		os.Exit(1)
	}
}
