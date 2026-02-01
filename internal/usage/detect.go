package usage

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"jombG/goblast/internal/symbols"
	"jombG/goblast/internal/tests"
)

type Usage struct {
	TestName   string
	TestFile   string
	SymbolName string
	SymbolKind string
}

// DetectUsages detects usages of changed symbols in test functions using go/types
func DetectUsages(discoveredTests []tests.Test, changedSymbols []symbols.Symbol) ([]Usage, error) {
	var usages []Usage

	// Load type information for changed symbols
	symbolObjects, err := resolveSymbolObjects(changedSymbols)
	if err != nil {
		return nil, err
	}

	// Group tests by package for efficient loading
	testsByPackage := groupTestsByPackage(discoveredTests)

	// For each package, load with types and check usages
	for pkgPath, pkgTests := range testsByPackage {
		pkgUsages, err := detectUsagesInPackage(pkgPath, pkgTests, symbolObjects, changedSymbols)
		if err != nil {
			// Skip packages that fail to load
			continue
		}
		usages = append(usages, pkgUsages...)
	}

	return usages, nil
}

// resolveSymbolObjects loads packages and resolves symbols to types.Object
func resolveSymbolObjects(changedSymbols []symbols.Symbol) (map[types.Object]symbols.Symbol, error) {
	result := make(map[types.Object]symbols.Symbol)

	// Group symbols by package for efficient loading
	pkgSymbols := make(map[string][]symbols.Symbol)
	for _, sym := range changedSymbols {
		pkgSymbols[sym.Package] = append(pkgSymbols[sym.Package], sym)
	}

	// Load each package with type information
	for pkgName, syms := range pkgSymbols {
		if len(syms) == 0 {
			continue
		}

		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		}

		pkgs, err := packages.Load(cfg, pkgName)
		if err != nil || len(pkgs) == 0 {
			continue
		}

		pkg := pkgs[0]
		if pkg.Types == nil || pkg.TypesInfo == nil {
			continue
		}

		// For each symbol, find its types.Object
		for _, sym := range syms {
			obj := findObjectInPackage(pkg, sym)
			if obj != nil {
				result[obj] = sym
			}
		}
	}

	return result, nil
}

// findObjectInPackage finds the types.Object for a symbol in a package
func findObjectInPackage(pkg *packages.Package, sym symbols.Symbol) types.Object {
	// Look through all definitions in the package
	for ident, obj := range pkg.TypesInfo.Defs {
		if ident.Name == sym.Name && obj != nil {
			// Verify it's the right kind of symbol
			switch sym.Kind {
			case "func":
				if _, ok := obj.(*types.Func); ok {
					// Check it's not a method
					if sig, ok := obj.Type().(*types.Signature); ok {
						if sig.Recv() == nil {
							return obj
						}
					}
				}
			case "method":
				if _, ok := obj.(*types.Func); ok {
					// Check it's a method
					if sig, ok := obj.Type().(*types.Signature); ok {
						if sig.Recv() != nil {
							return obj
						}
					}
				}
			case "type":
				if _, ok := obj.(*types.TypeName); ok {
					return obj
				}
			}
		}
	}
	return nil
}

// groupTestsByPackage groups tests by their package path
func groupTestsByPackage(testList []tests.Test) map[string][]tests.Test {
	result := make(map[string][]tests.Test)
	for _, test := range testList {
		result[test.Package] = append(result[test.Package], test)
	}
	return result
}

// detectUsagesInPackage detects usages in a specific package's tests
func detectUsagesInPackage(pkgPath string, pkgTests []tests.Test, symbolObjects map[types.Object]symbols.Symbol, changedSymbols []symbols.Symbol) ([]Usage, error) {
	var usages []Usage

	// Load the test package with type information
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Tests: true,
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil || len(pkgs) == 0 {
		return nil, err
	}

	// Find the test package (might be named with _test suffix)
	var testPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.TypesInfo != nil {
			testPkg = pkg
			break
		}
	}

	if testPkg == nil || testPkg.TypesInfo == nil {
		return nil, fmt.Errorf("no type info for package %s", pkgPath)
	}

	// For each test in this package
	for _, test := range pkgTests {
		testUsages := findUsagesInTest(testPkg, test, symbolObjects)
		usages = append(usages, testUsages...)
	}

	return usages, nil
}

// findUsagesInTest finds symbol usages in a specific test function
func findUsagesInTest(pkg *packages.Package, test tests.Test, symbolObjects map[types.Object]symbols.Symbol) []Usage {
	var usages []Usage

	// Find the test function in the AST
	var testFunc *ast.FuncDecl
	for _, file := range pkg.Syntax {
		// Check if this is the right file
		if filepath.Base(pkg.Fset.File(file.Pos()).Name()) != test.FileName {
			continue
		}

		// Find the test function
		ast.Inspect(file, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if funcDecl.Name != nil && funcDecl.Name.Name == test.Name {
					testFunc = funcDecl
					return false
				}
			}
			return true
		})

		if testFunc != nil {
			break
		}
	}

	if testFunc == nil || testFunc.Body == nil {
		return usages
	}

	// Traverse the test function body and check all identifiers and selectors
	ast.Inspect(testFunc.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Ident:
			// Check if this identifier refers to a changed symbol
			if obj := pkg.TypesInfo.Uses[node]; obj != nil {
				if sym, found := symbolObjects[obj]; found {
					usages = append(usages, Usage{
						TestName:   test.Name,
						TestFile:   test.Position,
						SymbolName: sym.Name,
						SymbolKind: sym.Kind,
					})
				}
			}
		case *ast.SelectorExpr:
			// Check selector expressions (method calls, field access)
			if sel := pkg.TypesInfo.Selections[node]; sel != nil {
				if obj := sel.Obj(); obj != nil {
					if sym, found := symbolObjects[obj]; found {
						usages = append(usages, Usage{
							TestName:   test.Name,
							TestFile:   test.Position,
							SymbolName: sym.Name,
							SymbolKind: sym.Kind,
						})
					}
				}
			}
			// Also check the selector identifier itself (for functions in other packages)
			if obj := pkg.TypesInfo.Uses[node.Sel]; obj != nil {
				if sym, found := symbolObjects[obj]; found {
					usages = append(usages, Usage{
						TestName:   test.Name,
						TestFile:   test.Position,
						SymbolName: sym.Name,
						SymbolKind: sym.Kind,
					})
				}
			}
		}
		return true
	})

	// Deduplicate usages
	return deduplicateUsages(usages)
}

// deduplicateUsages removes duplicate usage entries
func deduplicateUsages(usages []Usage) []Usage {
	seen := make(map[string]bool)
	var unique []Usage

	for _, usage := range usages {
		key := fmt.Sprintf("%s:%s:%s", usage.TestName, usage.SymbolName, usage.SymbolKind)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, usage)
		}
	}

	return unique
}

// FormatUsages formats detected usages for display
func FormatUsages(usages []Usage) string {
	if len(usages) == 0 {
		return "No usages detected."
	}

	var sb strings.Builder
	sb.WriteString("\n=== Detected Usages (Type-Based) ===\n\n")

	// Group by test
	testUsages := make(map[string][]Usage)
	for _, usage := range usages {
		key := fmt.Sprintf("%s (%s)", usage.TestName, usage.TestFile)
		testUsages[key] = append(testUsages[key], usage)
	}

	for testKey, usages := range testUsages {
		sb.WriteString(fmt.Sprintf("Test: %s\n", testKey))
		for _, usage := range usages {
			sb.WriteString(fmt.Sprintf("  - uses %s %s\n", usage.SymbolKind, usage.SymbolName))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Total: %d precise usages detected\n", len(usages)))
	return sb.String()
}
