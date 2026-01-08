package main

import (
	"flag"
	"log"
	"os"

	"github.com/kentik/custom-notification-templates/pkg/schemas"
)

func main() {
	outputPath := flag.String("output", "./docs/EVENT_VIEW_MODEL_DETAILS_REFERENCE.md", "Markdown output file path")

	details := schemas.Details()
	md := schemas.IntoMarkdown(details)

	err := os.WriteFile(*outputPath, []byte(md), 0644)
	if err != nil {
		log.Fatalf("Error writing to %s: %s", *outputPath, err)
	}
}
