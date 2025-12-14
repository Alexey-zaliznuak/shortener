package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
)

var Linter = &analysis.Analyzer{
	Name:     "noexit",
	Doc:      "Reports unwanted use of panic, log.Fatal, and os.Exit outside main function in main package",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Используем фильтр только по *ast.CallExpr
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inMainPkg := pass.Pkg != nil && pass.Pkg.Name() == "main"

	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}

		// проверка что это вызов функции
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Проверяем, находимся ли мы внутри функции main
		inMainFunc := false
		if inMainPkg {
			for i := len(stack) - 1; i >= 0; i-- {
				if fn, ok := stack[i].(*ast.FuncDecl); ok {
					if fn.Name != nil && fn.Name.Name == "main" {
						inMainFunc = true
					}
					break
				}
			}
		}

		// Проверка на panic
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
			pass.Reportf(call.Pos(), "usage of panic is not allowed")
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		obj := pass.TypesInfo.ObjectOf(sel.Sel)
		if obj == nil {
			return true
		}

		if obj.Pkg() != nil {
			pkgPath := obj.Pkg().Path()
			funcName := obj.Name()

			// Проверяем, что это log.Fatal или Exit
			isLogFatal := pkgPath == "log" && funcName == "Fatal"
			isOsExit := pkgPath == "os" && funcName == "Exit"

			if isLogFatal || isOsExit {
				// Разрешено только в main пакете и функции main
				if !inMainFunc {
					pass.Reportf(call.Pos(), "%s.%s is not allowed outside main function in main package",
						obj.Pkg().Name(), funcName)
				}
			}
		}

		return true
	})

	return nil, nil
}

func main() {
	singlechecker.Main(Linter)
}
