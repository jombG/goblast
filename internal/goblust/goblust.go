package goblust

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"jombG/goblast/internal/symbols"
	"jombG/goblast/internal/tests"
	"jombG/goblast/internal/usage"
)

func Run(base, head string, dryRun, debugSymbols, debugTests, debugTypes bool) error {
	var changedFiles []string

	committedFiles, err := getChangedFiles(base, head)
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}
	changedFiles = append(changedFiles, committedFiles...)

	uncommittedFiles, err := getUncommittedFiles()
	if err != nil {
		return fmt.Errorf("failed to get uncommitted files: %w", err)
	}
	changedFiles = append(changedFiles, uncommittedFiles...)

	changedFiles = deduplicateFiles(changedFiles)

	goFiles := filterGoFiles(changedFiles)
	if len(goFiles) == 0 {
		fmt.Println("No Go files changed. Nothing to test.")
		return nil
	}

	// Extract symbols from changed files (needed for both debug-symbols and debug-usage)
	var extractedSymbols []symbols.Symbol
	if debugSymbols || debugTypes {
		var err error
		extractedSymbols, err = symbols.ExtractFromFiles(goFiles)
		if err != nil {
			return fmt.Errorf("failed to extract symbols: %w", err)
		}
		if debugSymbols {
			fmt.Println(symbols.FormatSymbols(extractedSymbols))
		}
	}

	packages, err := mapFilesToPackages(goFiles)
	if err != nil {
		return fmt.Errorf("failed to map files to packages: %w", err)
	}

	if len(packages) == 0 {
		fmt.Println("No testable packages found for changed files.")
		return nil
	}

	uniquePackages := deduplicate(packages)

	// Discover tests from packages (needed for both debug-tests and debug-usage)
	var discoveredTests []tests.Test
	if debugTests || debugTypes {
		var err error
		discoveredTests, err = tests.DiscoverFromPackages(uniquePackages)
		if err != nil {
			return fmt.Errorf("failed to discover tests: %w", err)
		}
		if debugTests {
			fmt.Println(tests.FormatTests(discoveredTests))
		}
	}

	// Detect usages of changed symbols in tests
	if debugTypes {
		detectedUsages, err := usage.DetectUsages(discoveredTests, extractedSymbols)
		if err != nil {
			return fmt.Errorf("failed to detect usages: %w", err)
		}
		fmt.Println(usage.FormatUsages(detectedUsages))
	}

	testCmd := buildTestCommand(uniquePackages)

	if dryRun {
		fmt.Println(testCmd)
		return nil
	}

	return executeTestCommand(uniquePackages)
}

func getChangedFiles(base, head string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", base, head)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}

	return lines, nil
}

func getUncommittedFiles() ([]string, error) {
	// Get both staged and unstaged changes
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}

	return lines, nil
}

func filterGoFiles(files []string) []string {
	var goFiles []string
	for _, file := range files {
		if strings.HasPrefix(file, "vendor/") || strings.Contains(file, "/vendor/") {
			continue
		}
		if strings.HasSuffix(file, ".go") {
			goFiles = append(goFiles, file)
		}
	}
	return goFiles
}

func mapFilesToPackages(goFiles []string) ([]string, error) {
	var packages []string

	for _, file := range goFiles {
		dir := filepath.Dir(file)

		cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}", "./"+dir)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		pkg := strings.TrimSpace(string(output))
		if pkg != "" {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

func deduplicateFiles(files []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, file := range files {
		if !seen[file] {
			seen[file] = true
			unique = append(unique, file)
		}
	}

	return unique
}

func deduplicate(packages []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, pkg := range packages {
		if !seen[pkg] {
			seen[pkg] = true
			unique = append(unique, pkg)
		}
	}

	return unique
}

func buildTestCommand(packages []string) string {
	args := append([]string{"go", "test"}, packages...)
	return strings.Join(args, " ")
}

func executeTestCommand(packages []string) error {
	cmd := exec.Command("go", "test", packages[0])
	if len(packages) > 1 {
		cmd.Args = append(cmd.Args, packages[1:]...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}

	return nil
}
