package hostkit

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

func TestModelFacingInvestigationFunctionsDoNotProjectRawErrors(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve hostkit source path")
	}
	path := filepath.Join(filepath.Dir(filename), "investigation.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	wanted := map[string]bool{
		"BuildAStockInvestigationPayload": false,
		"runMultiEntityQuoteComparison":   false,
	}
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		if _, guarded := wanted[function.Name.Name]; !guarded {
			continue
		}
		wanted[function.Name.Name] = true
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if ok && selector.Sel.Name == "Error" && len(call.Args) == 0 {
				t.Errorf("%s must use safeerror before writing model-facing diagnostics", fset.Position(call.Pos()))
			}
			return true
		})
	}
	for name, found := range wanted {
		if !found {
			t.Errorf("guarded function %s is missing; update the canonical boundary guard with any rename", name)
		}
	}
}
