package mainosexit

import (
	"go/ast"
	"strings"

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

func isOsExit(call *ast.CallExpr) bool {
	if selExpr, fOk := call.Fun.(*ast.SelectorExpr); fOk {
		if ident, sOk := selExpr.X.(*ast.Ident); sOk &&
			ident.Name == "os" &&
			selExpr.Sel.Name == "Exit" {
			return true
		}
	}
	return false
}

//nolint:nilnil // nilnil
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Проверка на название пакета
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Проверка на функцию main.
			if fn, ok := n.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "main") {
				// Проходим именно по телу функции main.
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if callExpr, cOk := n.(*ast.CallExpr); cOk {
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
