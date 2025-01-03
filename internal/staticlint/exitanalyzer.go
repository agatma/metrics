// Package staticlint provides list of static analyzers for linter.
// nolint
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
		if file.Name.Name != packageDecl {
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
func isOsExitCall(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == packageName && sel.Sel.Name == funcName
}

func analyzeMain(pass *analysis.Pass, node ast.Node) {
	ast.Inspect(node, func(node ast.Node) bool {
		stmt, ok := node.(*ast.ExprStmt)
		if !ok {
			return true
		}

		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isOsExitCall(call.Fun) {
			pass.Reportf(call.Fun.Pos(), "main function contains os.Exit call")
		}
		return true
	})
}
