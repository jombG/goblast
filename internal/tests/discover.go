package tests

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"strings"
)

type Test struct {
	Package  string
	Name     string
	FileName string
	Position string
}

func DiscoverFromPackages(packages []string) ([]Test, error) {
	var allTests []Test

	for _, pkg := range packages {
		testFiles, err := findTestFilesInPackage(pkg)
		if err != nil {
			continue
		}

		for _, file := range testFiles {
			tests, err := discoverFromFile(file)
			if err != nil {
				continue
			}
			allTests = append(allTests, tests...)
		}
	}

	return allTests, nil
}

func findTestFilesInPackage(packagePath string) ([]string, error) {
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", packagePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	dir := strings.TrimSpace(string(output))

	pattern := filepath.Join(dir, "*_test.go")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func DiscoverFromFiles(files []string) ([]Test, error) {
	var allTests []Test

	testFiles := filterTestFiles(files)

	for _, file := range testFiles {
		tests, err := discoverFromFile(file)
		if err != nil {
			continue
		}
		allTests = append(allTests, tests...)
	}

	return allTests, nil
}

func filterTestFiles(files []string) []string {
	var testFiles []string
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			testFiles = append(testFiles, file)
		}
	}
	return testFiles
}

func discoverFromFile(filePath string) ([]Test, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, err
	}

	var tests []Test

	packagePath := getPackageImportPath(filePath)

	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if isTestFunction(funcDecl) {
				test := extractTest(funcDecl, packagePath, fset, filePath)
				if test != nil {
					tests = append(tests, *test)
				}
			}
			return false
		}
		return true
	})

	return tests, nil
}

func isTestFunction(decl *ast.FuncDecl) bool {
	if decl.Name == nil {
		return false
	}

	name := decl.Name.Name

	if !strings.HasPrefix(name, "Test") {
		return false
	}

	if decl.Type.Params == nil || len(decl.Type.Params.List) == 0 {
		return false
	}

	if decl.Recv != nil {
		return false
	}

	return true
}

func extractTest(decl *ast.FuncDecl, packagePath string, fset *token.FileSet, filePath string) *Test {
	if decl.Name == nil {
		return nil
	}

	pos := fset.Position(decl.Pos())
	return &Test{
		Package:  packagePath,
		Name:     decl.Name.Name,
		FileName: filepath.Base(filePath),
		Position: fmt.Sprintf("%s:%d", filepath.Base(filePath), pos.Line),
	}
}

func getPackageImportPath(filePath string) string {
	dir := filepath.Dir(filePath)
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}", "./"+dir)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to directory name
		return filepath.Base(dir)
	}
	return strings.TrimSpace(string(output))
}

func FormatTests(tests []Test) string {
	if len(tests) == 0 {
		return "No tests found."
	}

	var sb strings.Builder
	sb.WriteString("\n=== Discovered Tests ===\n\n")

	packageTests := make(map[string][]Test)
	for _, test := range tests {
		packageTests[test.Package] = append(packageTests[test.Package], test)
	}

	for pkg, pkgTests := range packageTests {
		sb.WriteString(fmt.Sprintf("Package: %s\n", pkg))
		for _, test := range pkgTests {
			sb.WriteString(fmt.Sprintf("  - %s at %s\n", test.Name, test.Position))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Total: %d tests\n", len(tests)))
	return sb.String()
}
