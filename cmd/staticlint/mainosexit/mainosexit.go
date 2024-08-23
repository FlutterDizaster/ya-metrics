package mainosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer проверяет функцию main на наличие прямого вызова os.Exit.
// Используется вместе с пакетом multichecker.
// Проверяется сначала название пакета и затем наличие функции main.
// При нахождении функции main, проверяется наличие вызова функции os.Exit.
// Если функция main вызывает функцию os.Exit, то выводится сообщение с указанием места вызова в коде.
//
//nolint:gochecknoglobals // for multichecker
var Analyzer = &analysis.Analyzer{
	Name: "mainosexit",
	Doc:  "check for os.Exit call in main function",
	Run:  run,
}

//nolint:govet // shadow declaration of 'ok' not affect on logic
func isOsExit(call *ast.CallExpr) bool {
	if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok &&
			ident.Name == "os" &&
			selExpr.Sel.Name == "Exit" {
			return true
		}
	}
	return false
}

//nolint:govet, nilnil // shadow declaration of 'ok' not affect on logic
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Проверка на название пакета
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Проверка на функцию main.
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				// Проходим именно по телу функции main.
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if callExpr, ok := n.(*ast.CallExpr); ok {
						// Проверка на вызов функции os.Exit.
						if isOsExit(callExpr) {
							pass.Reportf(callExpr.Pos(), "os.Exit in main function")
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}
