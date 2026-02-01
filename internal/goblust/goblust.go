package goblust

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run(base, head string, dryRun bool) error {
	changedFiles, err := getChangedFiles(base, head)
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	goFiles := filterGoFiles(changedFiles)
	if len(goFiles) == 0 {
		fmt.Println("No Go files changed. Nothing to test.")
		return nil
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
