package goblust

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Run executes the main logic of goblust.
func Run(base, head string, dryRun bool) error {
	// Get changed files from git
	changedFiles, err := getChangedFiles(base, head)
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	// Filter for .go files and exclude vendor/
	goFiles := filterGoFiles(changedFiles)
	if len(goFiles) == 0 {
		fmt.Println("No Go files changed. Nothing to test.")
		return nil
	}

	// Map files to Go package paths
	packages, err := mapFilesToPackages(goFiles)
	if err != nil {
		return fmt.Errorf("failed to map files to packages: %w", err)
	}

	if len(packages) == 0 {
		fmt.Println("No testable packages found for changed files.")
		return nil
	}

	// Deduplicate packages
	uniquePackages := deduplicate(packages)

	// Build test command
	testCmd := buildTestCommand(uniquePackages)

	if dryRun {
		fmt.Println(testCmd)
		return nil
	}

	// Execute test command
	return executeTestCommand(uniquePackages)
}

// getChangedFiles returns list of changed files between base and head commits.
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

// filterGoFiles filters for .go files and excludes vendor/ directory.
func filterGoFiles(files []string) []string {
	var goFiles []string
	for _, file := range files {
		// Skip vendor directory
		if strings.HasPrefix(file, "vendor/") || strings.Contains(file, "/vendor/") {
			continue
		}
		// Include only .go files
		if strings.HasSuffix(file, ".go") {
			goFiles = append(goFiles, file)
		}
	}
	return goFiles
}

// mapFilesToPackages maps Go files to their package paths using go list.
func mapFilesToPackages(goFiles []string) ([]string, error) {
	var packages []string

	for _, file := range goFiles {
		// Get directory of the file
		dir := filepath.Dir(file)

		// Use go list to get the package path
		cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}", "./"+dir)
		output, err := cmd.Output()
		if err != nil {
			// If go list fails, skip this file (might be deleted or not part of a valid package)
			continue
		}

		pkg := strings.TrimSpace(string(output))
		if pkg != "" {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// deduplicate removes duplicate package paths.
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

// buildTestCommand builds the go test command string for display.
func buildTestCommand(packages []string) string {
	args := append([]string{"go", "test"}, packages...)
	return strings.Join(args, " ")
}

// executeTestCommand runs go test with the specified packages.
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
