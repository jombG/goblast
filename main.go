package main

import (
	"flag"
	"fmt"
	"os"

	"jombG/goblast/internal/goblust"
)

func main() {
	// Parse CLI flags
	base := flag.String("base", "main", "base branch for comparison (default: main)")
	head := flag.String("head", "HEAD", "head commit for comparison")
	uncommitted := flag.Bool("uncommitted", false, "check uncommitted changes (ignores base/head)")
	dryRun := flag.Bool("dry-run", false, "print test command without executing")
	flag.Parse()

	// Run the tool
	if err := goblust.Run(*base, *head, *uncommitted, *dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
