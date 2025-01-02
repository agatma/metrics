// Package staticlint provides list of static analyzers for linter.
package staticlint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	packageDecl = "main"
	funcDecl    = "main"
	packageName = "os"
	funcName    = "Exit"
)

// MainExitAnalyzer defines an analyzer that checks for calls to os.Exit within the main function.
var MainExitAnalyzer = &analysis.Analyzer{
	Name: "main_check",
	Doc:  "check for os.Exit in main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == packageDecl {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if x, ok := node.(*ast.FuncDecl); ok && x.Name.Name == funcDecl {
				analyzeMain(pass, node)
			}
			return true
		})
	}
	return nil, nil
}
func analyzeMain(pass *analysis.Pass, node ast.Node) {
	expr := func(x *ast.ExprStmt) {
		if call, ok := x.X.(*ast.CallExpr); ok {
			if selector, ok := (call.Fun).(*ast.SelectorExpr); ok {
				if ident, ok := (selector.X).(*ast.Ident); ok &&
					ident.Name == packageName &&
					selector.Sel.Name == funcName {
					pass.Reportf(selector.Pos(), "main function contains os.Exit call")
				}
			}
		}
	}
	ast.Inspect(node, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.ExprStmt:
			expr(x)
		}
		return true
	})
}
