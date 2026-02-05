package selector

import (
	"fmt"

	"jombG/goblast/internal/symbols"
	"jombG/goblast/internal/tests"
	"jombG/goblast/internal/usage"
)

// TestID uniquely identifies a test for execution
type TestID struct {
	Package  string
	TestName string
}

// Strategy defines how to select tests based on changes and usages
type Strategy interface {
	Name() string
	Select(changedSymbols []symbols.Symbol, discoveredTests []tests.Test, usages []usage.Usage) []TestID
}

// GetStrategy returns a strategy by name
func GetStrategy(name string) (Strategy, error) {
	switch name {
	case "symbol-only":
		return &SymbolOnlyStrategy{}, nil
	case "package-fallback":
		return &PackageFallbackStrategy{}, nil
	case "conservative":
		return &ConservativeStrategy{}, nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}

// SymbolOnlyStrategy runs only tests that directly use changed symbols
type SymbolOnlyStrategy struct{}

func (s *SymbolOnlyStrategy) Name() string {
	return "symbol-only"
}

func (s *SymbolOnlyStrategy) Select(changedSymbols []symbols.Symbol, discoveredTests []tests.Test, usages []usage.Usage) []TestID {
	var selected []TestID

	// Build a set of tests that use changed symbols
	testSet := make(map[string]map[string]bool) // pkg -> testName -> exists
	for _, u := range usages {
		// Extract package from test
		pkg := findTestPackage(u.TestName, discoveredTests)
		if pkg == "" {
			continue
		}
		if testSet[pkg] == nil {
			testSet[pkg] = make(map[string]bool)
		}
		testSet[pkg][u.TestName] = true
	}

	// Convert to TestID list
	for pkg, tests := range testSet {
		for testName := range tests {
			selected = append(selected, TestID{
				Package:  pkg,
				TestName: testName,
			})
		}
	}

	return selected
}

// PackageFallbackStrategy runs tests that use changed symbols,
// or all tests in a package if no specific usage detected but package changed
type PackageFallbackStrategy struct{}

func (s *PackageFallbackStrategy) Name() string {
	return "package-fallback"
}

func (s *PackageFallbackStrategy) Select(changedSymbols []symbols.Symbol, discoveredTests []tests.Test, usages []usage.Usage) []TestID {
	var selected []TestID

	// First, collect tests with direct usages (from ANY package, including dependents)
	usedTests := make(map[string]map[string]bool) // pkg -> testName -> exists
	for _, u := range usages {
		pkg := findTestPackage(u.TestName, discoveredTests)
		if pkg == "" {
			continue
		}
		if usedTests[pkg] == nil {
			usedTests[pkg] = make(map[string]bool)
		}
		usedTests[pkg][u.TestName] = true
	}

	// Add all tests with direct usages (from any package)
	for pkg, tests := range usedTests {
		for testName := range tests {
			selected = append(selected, TestID{
				Package:  pkg,
				TestName: testName,
			})
		}
	}

	// Collect packages with changes
	changedPackages := make(map[string]bool)
	for _, sym := range changedSymbols {
		changedPackages[sym.Package] = true
	}

	// For each changed package that has no specific usages - run all tests (fallback)
	for pkg := range changedPackages {
		if len(usedTests[pkg]) == 0 {
			// No specific usages detected - run all tests in package (fallback)
			for _, test := range discoveredTests {
				if test.Package == pkg {
					selected = append(selected, TestID{
						Package:  test.Package,
						TestName: test.Name,
					})
				}
			}
		}
	}

	return deduplicateTestIDs(selected)
}

// ConservativeStrategy runs all tests in all packages with changes
type ConservativeStrategy struct{}

func (s *ConservativeStrategy) Name() string {
	return "conservative"
}

func (s *ConservativeStrategy) Select(changedSymbols []symbols.Symbol, discoveredTests []tests.Test, usages []usage.Usage) []TestID {
	var selected []TestID

	// Collect all packages with changes
	changedPackages := make(map[string]bool)
	for _, sym := range changedSymbols {
		changedPackages[sym.Package] = true
	}

	// Run all tests in changed packages
	for _, test := range discoveredTests {
		if changedPackages[test.Package] {
			selected = append(selected, TestID{
				Package:  test.Package,
				TestName: test.Name,
			})
		}
	}

	return deduplicateTestIDs(selected)
}

// Helper functions

func findTestPackage(testName string, tests []tests.Test) string {
	for _, test := range tests {
		if test.Name == testName {
			return test.Package
		}
	}
	return ""
}

func deduplicateTestIDs(ids []TestID) []TestID {
	seen := make(map[string]bool)
	var unique []TestID

	for _, id := range ids {
		key := fmt.Sprintf("%s::%s", id.Package, id.TestName)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, id)
		}
	}

	return unique
}

// FormatSelection formats selected tests for display
func FormatSelection(strategy string, selected []TestID) string {
	if len(selected) == 0 {
		return fmt.Sprintf("\n=== Test Selection (%s) ===\n\nNo tests selected.\n", strategy)
	}

	// Group by package
	byPackage := make(map[string][]string)
	for _, id := range selected {
		byPackage[id.Package] = append(byPackage[id.Package], id.TestName)
	}

	result := fmt.Sprintf("\n=== Test Selection (%s) ===\n\n", strategy)
	for pkg, testNames := range byPackage {
		result += fmt.Sprintf("Package: %s\n", pkg)
		for _, name := range testNames {
			result += fmt.Sprintf("  - %s\n", name)
		}
		result += "\n"
	}

	result += fmt.Sprintf("Total: %d tests selected\n", len(selected))
	return result
}
