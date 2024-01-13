package expression

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"strings"
)

func getCode(src string, node ast.Node) string {
	if node == nil || !node.Pos().IsValid() || !node.End().IsValid() {
		return ""
	}
	end := int(node.End()) - 1
	if end > len(src) {
		end = len(src)
	}
	return src[node.Pos()-1 : end]
}

var ErrContainerFuncNotFound = errors.New("parser error: templ container function not found")

func ParseExpression(content string) (expr string, err error) {
	//TODO: Handle whitespace between else and bracket etc.
	if strings.HasPrefix(content, "else {") {
		return "else {", nil
	}
	fset := token.NewFileSet() // positions are relative to fset

	//TODO: Handle whitespace etc.
	if strings.HasPrefix(content, "else if") {
		expr, err = ParseExpression(strings.TrimPrefix(content, "else "))
		if err != nil {
			return expr, err
		}
		return "else " + expr, nil
	}

	prefix := "package main\nfunc templ_container() {\n"
	src := prefix + content

	node, parseErr := parser.ParseFile(fset, "", src, parser.AllErrors)
	if node == nil {
		return expr, parseErr
	}

	// Print the imports from the file's AST.
	ast.Inspect(node, func(n ast.Node) bool {
		// Find the "templ_container" function.
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Name.Name != "templ_container" {
			err = ErrContainerFuncNotFound
			return false
		}
		// We only expect a single statement.
		if len(fn.Body.List) == 0 {
			// No expression found.
			return false
		}
		// Check the container function contents to find the first expression.
		// We expect a statement.
		stmt, ok := fn.Body.List[0].(ast.Stmt)
		if !ok {
			// No Go statement found.
			return false
		}
		// We found something, stop looking.
		switch stmt := stmt.(type) {
		case *ast.IfStmt:
			// Only get the code up until the first `{`.
			expr = getCode(src, stmt)[:int(stmt.Body.Lbrace)-int(stmt.If)+1]
		case *ast.ForStmt:
			// Only get the code up until the first `{`.
			expr = getCode(src, stmt)[:int(stmt.Body.Lbrace)-int(stmt.For)+1]
		case *ast.RangeStmt:
			// Only get the code up until the first `{`.
			expr = getCode(src, stmt)[:int(stmt.Body.Lbrace)-int(stmt.For)+1]
		default:
			// Just an expression.
			fmt.Printf("%T\n", stmt)
			expr = getCode(src, stmt)
		}

		// If we have a parse error that's later than the position of our expression we can ignore it.
		// Because we only want to nibble the first valid expression.
		// Anything after the first expression is likely to be templ code.
		// But if it's in the first expression, it can help us see the problem early in templ.
		if parseErr != nil {
			serr, ok := err.(scanner.ErrorList)
			if !ok {
				return false
			}
			serr.Sort()
			if stmt.End() < token.Pos(serr[0].Pos.Offset) {
				err = serr[0]
			}
		}

		return false
	})

	return expr, err
}
