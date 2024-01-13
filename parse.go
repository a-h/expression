package expression

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
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
var ErrExpectedNodeNotFound = errors.New("parser error: expected node not found")

var prefixRegexps = []*regexp.Regexp{
	regexp.MustCompile(`^if`),
	regexp.MustCompile(`^else`),
	regexp.MustCompile(`^for`),
	regexp.MustCompile(`^switch`),
	regexp.MustCompile(`^case`),
	regexp.MustCompile(`^default`),
}
var prefixExtractors = []Extractor{
	IfExtractor{},
	ElseExtractor{},
	ForExtractor{},
	SwitchExtractor{},
	CaseExtractor{},
	DefaultExtractor{},
}

func ParseExpression(content string) (expr string, err error) {
	//TODO: Handle whitespace between else and bracket etc.
	if strings.HasPrefix(content, "else {") {
		return "else {", nil
	}

	//TODO: Handle whitespace etc.
	if strings.HasPrefix(content, "else if") {
		expr, err = parseExp(strings.TrimPrefix(content, "else "), IfExtractor{})
		if err != nil {
			return expr, err
		}
		return "else " + expr, nil
	}

	if strings.HasPrefix(content, "case") {
		expr = "switch {\n" + content + "\n}"
		expr, err = parseExp(expr, CaseExtractor{})
		if err != nil {
			return expr, err
		}
		return expr, nil
	}

	for i, re := range prefixRegexps {
		if re.MatchString(content) {
			expr, err = parseExp(content, prefixExtractors[i])
			if err != nil {
				return expr, err
			}
			return expr, nil
		}
	}

	expr, err = parseExp(content, ExprExtractor{})
	if err != nil {
		return expr, err
	}
	return expr, nil
	//TODO: If we're doing an expression, check the end to see if it's children.
}

type IfExtractor struct{}

func (e IfExtractor) Code(src string, body []ast.Stmt) (expr string, err error) {
	stmt, ok := body[0].(*ast.IfStmt)
	if !ok {
		return expr, ErrExpectedNodeNotFound
	}
	expr = getCode(src, stmt)[:int(stmt.Body.Lbrace)-int(stmt.If)+1]
	return expr, nil
}

type ElseExtractor struct{}

func (e ElseExtractor) Code(src string, body []ast.Stmt) (expr string, err error) {
	stmt, ok := body[0].(*ast.ExprStmt)
	if !ok {
		return expr, ErrExpectedNodeNotFound
	}
	expr = getCode(src, stmt)
	return expr, nil
}

type ForExtractor struct{}

func (e ForExtractor) Code(src string, body []ast.Stmt) (expr string, err error) {
	stmt := body[0]
	switch stmt := stmt.(type) {
	case *ast.ForStmt:
		// Only get the code up until the first `{`.
		expr = getCode(src, stmt)[:int(stmt.Body.Lbrace)-int(stmt.For)+1]
		return expr, nil
	case *ast.RangeStmt:
		// Only get the code up until the first `{`.
		expr = getCode(src, stmt)[:int(stmt.Body.Lbrace)-int(stmt.For)+1]
		return expr, nil
	}
	return expr, ErrExpectedNodeNotFound
}

type SwitchExtractor struct{}

func (e SwitchExtractor) Code(src string, body []ast.Stmt) (expr string, err error) {
	stmt := body[0]
	switch stmt := stmt.(type) {
	case *ast.SwitchStmt:
		// Only get the code up until the first `{`.
		expr = getCode(src, stmt)[:int(stmt.Body.Lbrace)-int(stmt.Switch)+1]
		return expr, nil
	case *ast.TypeSwitchStmt:
		// Only get the code up until the first `{`.
		expr = getCode(src, stmt)[:int(stmt.Body.Lbrace)-int(stmt.Switch)+1]
		return expr, nil
	}
	return expr, ErrExpectedNodeNotFound
}

type CaseExtractor struct{}

func (e CaseExtractor) Code(src string, body []ast.Stmt) (expr string, err error) {
	sw, ok := body[0].(*ast.SwitchStmt)
	if !ok {
		return expr, ErrExpectedNodeNotFound
	}
	stmt, ok := sw.Body.List[0].(*ast.CaseClause)
	if !ok {
		return expr, ErrExpectedNodeNotFound
	}
	start := int(stmt.Pos() - 1)
	end := stmt.Colon
	return src[start:end], nil
}

type DefaultExtractor struct{}

func (e DefaultExtractor) Code(src string, body []ast.Stmt) (expr string, err error) {
	stmt, ok := body[0].(*ast.CaseClause)
	if !ok {
		return expr, ErrExpectedNodeNotFound
	}
	expr = getCode(src, stmt)
	return expr, nil
}

type ExprExtractor struct{}

func (e ExprExtractor) Code(src string, body []ast.Stmt) (expr string, err error) {
	stmt, ok := body[0].(*ast.ExprStmt)
	if !ok {
		return expr, ErrExpectedNodeNotFound
	}
	expr = getCode(src, stmt)
	return expr, nil
}

type ChildrenExtractor struct{}

func (e ChildrenExtractor) Code(src string, body []ast.Stmt) (expr string, err error) {
	stmt, ok := body[0].(*ast.ExprStmt)
	if !ok {
		return expr, ErrExpectedNodeNotFound
	}
	expr = getCode(src, stmt)
	// Check that the three chars after expr in the source are `...`
	if !strings.HasPrefix(src[stmt.End():], "...") {
		return expr, ErrExpectedNodeNotFound
	}
	return expr, nil
}

type Extractor interface {
	Code(src string, body []ast.Stmt) (expr string, err error)
}

func parseExp(content string, extractor Extractor) (expr string, err error) {
	prefix := "package main\nfunc templ_container() {\n"
	src := prefix + content

	node, parseErr := parser.ParseFile(token.NewFileSet(), "", src, parser.AllErrors)
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
		if fn.Body.List == nil || len(fn.Body.List) == 0 {
			return false
		}
		expr, err = extractor.Code(src, fn.Body.List)
		return false
	})
	if err != nil {
		return expr, err
	}

	return expr, err
}
