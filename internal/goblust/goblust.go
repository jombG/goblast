package goblust

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"jombG/goblast/internal/selector"
	"jombG/goblast/internal/symbols"
	"jombG/goblast/internal/tests"
	"jombG/goblast/internal/usage"
)

func Run(base, head string, dryRun, debugFiles, debugSymbols, debugTests, debugTypes bool, strategyName string, debugSelection bool) error {
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
	if debugFiles {
		fmt.Println("Affected Go files:")
		for _, f := range goFiles {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}
	if len(goFiles) == 0 {
		fmt.Println("No Go files changed. Nothing to test.")
		return nil
	}

	// Extract symbols from changed files (always needed for selection)
	extractedSymbols, err := symbols.ExtractFromFiles(goFiles)
	if err != nil {
		return fmt.Errorf("failed to extract symbols: %w", err)
	}
	if debugSymbols {
		fmt.Println(symbols.FormatSymbols(extractedSymbols))
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

	// Discover tests from packages (always needed for selection)
	discoveredTests, err := tests.DiscoverFromPackages(uniquePackages)
	if err != nil {
		return fmt.Errorf("failed to discover tests: %w", err)
	}
	if debugTests {
		fmt.Println(tests.FormatTests(discoveredTests))
	}

	// Detect usages of changed symbols in tests (always needed for selection)
	detectedUsages, err := usage.DetectUsages(discoveredTests, extractedSymbols)
	if err != nil {
		return fmt.Errorf("failed to detect usages: %w", err)
	}
	if debugTypes {
		fmt.Println(usage.FormatUsages(detectedUsages))
	}

	// Apply selection strategy
	strategy, err := selector.GetStrategy(strategyName)
	if err != nil {
		return fmt.Errorf("failed to get strategy: %w", err)
	}

	selectedTests := strategy.Select(extractedSymbols, discoveredTests, detectedUsages)

	if debugSelection {
		fmt.Println(selector.FormatSelection(strategy.Name(), selectedTests))
	}

	if len(selectedTests) == 0 {
		fmt.Println("No tests selected by strategy. Nothing to run.")
		return nil
	}

	testCmd := buildTestCommandFromSelection(selectedTests)

	if dryRun {
		fmt.Println(testCmd)
		return nil
	}

	return executeSelectedTests(selectedTests)
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
	seen := make(map[string]struct{})
	var unique []string

	for _, file := range files {
		if _, ok := seen[file]; !ok {
			seen[file] = struct{}{}
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

func buildTestCommandFromSelection(selected []selector.TestID) string {
	// Group tests by package
	byPackage := make(map[string][]string)
	for _, test := range selected {
		byPackage[test.Package] = append(byPackage[test.Package], test.TestName)
	}

	// Build command string
	var parts []string
	for pkg, testNames := range byPackage {
		testPattern := strings.Join(testNames, "|")
		parts = append(parts, fmt.Sprintf("go test %s -run '^(%s)$'", pkg, testPattern))
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return strings.Join(parts, " && ")
}

func executeSelectedTests(selected []selector.TestID) error {
	// Group tests by package
	byPackage := make(map[string][]string)
	for _, test := range selected {
		byPackage[test.Package] = append(byPackage[test.Package], test.TestName)
	}

	// Execute tests for each package
	for pkg, testNames := range byPackage {
		testPattern := "^(" + strings.Join(testNames, "|") + ")$"

		cmd := exec.Command("go", "test", pkg, "-run", testPattern)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go test failed for package %s: %w", pkg, err)
		}
	}

	return nil
}
