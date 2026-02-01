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
	dryRun := flag.Bool("dry-run", false, "print test command without executing")
	debugSymbols := flag.Bool("debug-symbols", false, "print extracted symbols from changed files")
	debugTests := flag.Bool("debug-tests", false, "print discovered test functions from changed files")
	debugTypes := flag.Bool("debug-types", false, "print precise type-based usages of changed symbols in tests")
	strategy := flag.String("strategy", "package-fallback", "test selection strategy: symbol-only, package-fallback, conservative")
	debugSelection := flag.Bool("debug-selection", false, "print selected tests based on strategy")
	flag.Parse()

	// Run the tool
	if err := goblust.Run(*base, *head, *dryRun, *debugSymbols, *debugTests, *debugTypes, *strategy, *debugSelection); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
