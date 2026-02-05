package symbols

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"strings"
)

type Symbol struct {
	Package  string
	Name     string
	Kind     string
	Receiver string
	Exported bool
	Position string
}

func ExtractFromFiles(files []string) ([]Symbol, error) {
	var allSymbols []Symbol

	for _, file := range files {
		symbols, err := extractFromFile(file)
		if err != nil {
			continue
		}
		allSymbols = append(allSymbols, symbols...)
	}

	return allSymbols, nil
}

func extractFromFile(filePath string) ([]Symbol, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, err
	}

	var symbols []Symbol
	packagePath := getPackageImportPath(filePath)

	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			symbol := extractFunction(decl, packagePath, fset, filePath)
			if symbol != nil {
				symbols = append(symbols, *symbol)
			}
			return false

		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						symbol := extractType(typeSpec, packagePath, fset, filePath)
						if symbol != nil {
							symbols = append(symbols, *symbol)
						}
					}
				}
			}
			return false
		}
		return true
	})

	return symbols, nil
}

func extractFunction(decl *ast.FuncDecl, pkgName string, fset *token.FileSet, filePath string) *Symbol {
	if decl.Name == nil {
		return nil
	}

	pos := fset.Position(decl.Pos())
	symbol := &Symbol{
		Package:  pkgName,
		Name:     decl.Name.Name,
		Exported: ast.IsExported(decl.Name.Name),
		Position: fmt.Sprintf("%s:%d", filepath.Base(filePath), pos.Line),
	}

	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		symbol.Kind = "method"
		symbol.Receiver = extractReceiverType(decl.Recv.List[0].Type)
	} else {
		symbol.Kind = "func"
	}

	return symbol
}

func extractReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	case *ast.IndexExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

func extractType(spec *ast.TypeSpec, pkgName string, fset *token.FileSet, filePath string) *Symbol {
	if spec.Name == nil {
		return nil
	}

	pos := fset.Position(spec.Pos())
	symbol := &Symbol{
		Package:  pkgName,
		Name:     spec.Name.Name,
		Kind:     "type",
		Exported: ast.IsExported(spec.Name.Name),
		Position: fmt.Sprintf("%s:%d", filepath.Base(filePath), pos.Line),
	}

	return symbol
}

func getPackageImportPath(filePath string) string {
	dir := filepath.Dir(filePath)

	var listPath string
	if filepath.IsAbs(dir) {
		listPath = dir
	} else {
		listPath = "./" + dir
	}

	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}", listPath)
	output, err := cmd.Output()
	if err != nil {
		return filepath.Base(dir)
	}
	return strings.TrimSpace(string(output))
}

func FormatSymbols(symbols []Symbol) string {
	if len(symbols) == 0 {
		return "No symbols found."
	}

	var sb strings.Builder
	sb.WriteString("\n=== Extracted Symbols ===\n\n")

	for _, sym := range symbols {
		visibility := "unexported"
		if sym.Exported {
			visibility = "exported"
		}

		switch sym.Kind {
		case "func":
			sb.WriteString(fmt.Sprintf("[%s] func %s.%s at %s\n",
				visibility, sym.Package, sym.Name, sym.Position))
		case "method":
			sb.WriteString(fmt.Sprintf("[%s] method (%s) %s.%s at %s\n",
				visibility, sym.Receiver, sym.Package, sym.Name, sym.Position))
		case "type":
			sb.WriteString(fmt.Sprintf("[%s] type %s.%s at %s\n",
				visibility, sym.Package, sym.Name, sym.Position))
		}
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d symbols\n", len(symbols)))
	return sb.String()
}
