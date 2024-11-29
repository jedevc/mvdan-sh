// Copyright (c) 2016, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package syntax

import (
	"fmt"
	"io"
	"reflect"
)

func walkStmts(stmts []*Stmt, last []Comment, f func(Node) bool) {
	for _, s := range stmts {
		Walk(nilify(s), f)
	}
	for _, c := range last {
		Walk(nilify(&c), f)
	}
}

func walkWords(words []*Word, f func(Node) bool) {
	for _, w := range words {
		Walk(nilify(w), f)
	}
}

type comparableNode interface {
	comparable
	Node
}

func nilify[T comparableNode](v T) Node {
	var zero T
	if v == zero {
		return nil
	}
	return v
}

// Walk traverses a syntax tree in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Walk invokes f
// recursively for each of the non-nil children of node, followed by
// f(nil).
func Walk(node Node, f func(Node) bool) {
	if node == nil {
		return
	}
	if !f(node) {
		return
	}

	switch node := node.(type) {
	case *File:
		walkStmts(node.Stmts, node.Last, f)
	case *Comment:
	case *Stmt:
		for _, c := range node.Comments {
			if !node.End().After(c.Pos()) {
				defer Walk(nilify(&c), f)
				break
			}
			Walk(nilify(&c), f)
		}
		if node.Cmd != nil {
			Walk(nilify(node.Cmd), f)
		}
		for _, r := range node.Redirs {
			Walk(nilify(r), f)
		}
	case *Assign:
		if node.Name != nil {
			Walk(nilify(node.Name), f)
		}
		if node.Value != nil {
			Walk(nilify(node.Value), f)
		}
		if node.Index != nil {
			Walk(nilify(node.Index), f)
		}
		if node.Array != nil {
			Walk(nilify(node.Array), f)
		}
	case *Redirect:
		if node.N != nil {
			Walk(nilify(node.N), f)
		}
		Walk(nilify(node.Word), f)
		if node.Hdoc != nil {
			Walk(nilify(node.Hdoc), f)
		}
	case *CallExpr:
		for _, a := range node.Assigns {
			Walk(nilify(a), f)
		}
		walkWords(node.Args, f)
	case *Subshell:
		walkStmts(node.Stmts, node.Last, f)
	case *Block:
		walkStmts(node.Stmts, node.Last, f)
	case *IfClause:
		walkStmts(node.Cond, node.CondLast, f)
		walkStmts(node.Then, node.ThenLast, f)
		if node.Else != nil {
			Walk(nilify(node.Else), f)
		}
	case *WhileClause:
		walkStmts(node.Cond, node.CondLast, f)
		walkStmts(node.Do, node.DoLast, f)
	case *ForClause:
		Walk(nilify(node.Loop), f)
		walkStmts(node.Do, node.DoLast, f)
	case *WordIter:
		Walk(nilify(node.Name), f)
		walkWords(node.Items, f)
	case *CStyleLoop:
		if node.Init != nil {
			Walk(nilify(node.Init), f)
		}
		if node.Cond != nil {
			Walk(nilify(node.Cond), f)
		}
		if node.Post != nil {
			Walk(nilify(node.Post), f)
		}
	case *BinaryCmd:
		Walk(nilify(node.X), f)
		Walk(nilify(node.Y), f)
	case *FuncDecl:
		Walk(nilify(node.Name), f)
		Walk(nilify(node.Body), f)
	case *Word:
		for _, wp := range node.Parts {
			Walk(nilify(wp), f)
		}
	case *Lit:
	case *SglQuoted:
	case *DblQuoted:
		for _, wp := range node.Parts {
			Walk(nilify(wp), f)
		}
	case *CmdSubst:
		walkStmts(node.Stmts, node.Last, f)
	case *ParamExp:
		Walk(nilify(node.Param), f)
		if node.Index != nil {
			Walk(nilify(node.Index), f)
		}
		if node.Repl != nil {
			if node.Repl.Orig != nil {
				Walk(nilify(node.Repl.Orig), f)
			}
			if node.Repl.With != nil {
				Walk(nilify(node.Repl.With), f)
			}
		}
		if node.Exp != nil && node.Exp.Word != nil {
			Walk(nilify(node.Exp.Word), f)
		}
	case *ArithmExp:
		Walk(nilify(node.X), f)
	case *ArithmCmd:
		Walk(nilify(node.X), f)
	case *BinaryArithm:
		Walk(nilify(node.X), f)
		Walk(nilify(node.Y), f)
	case *BinaryTest:
		Walk(nilify(node.X), f)
		Walk(nilify(node.Y), f)
	case *UnaryArithm:
		Walk(nilify(node.X), f)
	case *UnaryTest:
		Walk(nilify(node.X), f)
	case *ParenArithm:
		Walk(nilify(node.X), f)
	case *ParenTest:
		Walk(nilify(node.X), f)
	case *CaseClause:
		Walk(nilify(node.Word), f)
		for _, ci := range node.Items {
			Walk(nilify(ci), f)
		}
		for _, c := range node.Last {
			Walk(nilify(&c), f)
		}
	case *CaseItem:
		for _, c := range node.Comments {
			if c.Pos().After(node.Pos()) {
				defer Walk(nilify(&c), f)
				break
			}
			Walk(nilify(&c), f)
		}
		walkWords(node.Patterns, f)
		walkStmts(node.Stmts, node.Last, f)
	case *TestClause:
		Walk(nilify(node.X), f)
	case *DeclClause:
		for _, a := range node.Args {
			Walk(nilify(a), f)
		}
	case *ArrayExpr:
		for _, el := range node.Elems {
			Walk(nilify(el), f)
		}
		for _, c := range node.Last {
			Walk(nilify(&c), f)
		}
	case *ArrayElem:
		for _, c := range node.Comments {
			if c.Pos().After(node.Pos()) {
				defer Walk(nilify(&c), f)
				break
			}
			Walk(nilify(&c), f)
		}
		if node.Index != nil {
			Walk(nilify(node.Index), f)
		}
		if node.Value != nil {
			Walk(nilify(node.Value), f)
		}
	case *ExtGlob:
		Walk(nilify(node.Pattern), f)
	case *ProcSubst:
		walkStmts(node.Stmts, node.Last, f)
	case *TimeClause:
		if node.Stmt != nil {
			Walk(nilify(node.Stmt), f)
		}
	case *CoprocClause:
		if node.Name != nil {
			Walk(nilify(node.Name), f)
		}
		Walk(nilify(node.Stmt), f)
	case *LetClause:
		for _, expr := range node.Exprs {
			Walk(nilify(expr), f)
		}
	case *TestDecl:
		Walk(nilify(node.Description), f)
		Walk(nilify(node.Body), f)
	default:
		panic(fmt.Sprintf("syntax.Walk: unexpected node type %T", node))
	}

	f(nil)
}

// DebugPrint prints the provided syntax tree, spanning multiple lines and with
// indentation. Can be useful to investigate the content of a syntax tree.
func DebugPrint(w io.Writer, node Node) error {
	p := debugPrinter{out: w}
	p.print(reflect.ValueOf(node))
	p.printf("\n")
	return p.err
}

type debugPrinter struct {
	out   io.Writer
	level int
	err   error
}

func (p *debugPrinter) printf(format string, args ...any) {
	_, err := fmt.Fprintf(p.out, format, args...)
	if err != nil && p.err == nil {
		p.err = err
	}
}

func (p *debugPrinter) newline() {
	p.printf("\n")
	for i := 0; i < p.level; i++ {
		p.printf(".  ")
	}
}

func (p *debugPrinter) print(x reflect.Value) {
	switch x.Kind() {
	case reflect.Interface:
		if x.IsNil() {
			p.printf("nil")
			return
		}
		p.print(x.Elem())
	case reflect.Ptr:
		if x.IsNil() {
			p.printf("nil")
			return
		}
		p.printf("*")
		p.print(x.Elem())
	case reflect.Slice:
		p.printf("%s (len = %d) {", x.Type(), x.Len())
		if x.Len() > 0 {
			p.level++
			p.newline()
			for i := 0; i < x.Len(); i++ {
				p.printf("%d: ", i)
				p.print(x.Index(i))
				if i == x.Len()-1 {
					p.level--
				}
				p.newline()
			}
		}
		p.printf("}")

	case reflect.Struct:
		if v, ok := x.Interface().(Pos); ok {
			p.printf("%v:%v", v.Line(), v.Col())
			return
		}
		t := x.Type()
		p.printf("%s {", t)
		p.level++
		p.newline()
		for i := 0; i < t.NumField(); i++ {
			p.printf("%s: ", t.Field(i).Name)
			p.print(x.Field(i))
			if i == x.NumField()-1 {
				p.level--
			}
			p.newline()
		}
		p.printf("}")
	default:
		if s, ok := x.Interface().(fmt.Stringer); ok && !x.IsZero() {
			p.printf("%#v (%s)", x.Interface(), s)
		} else {
			p.printf("%#v", x.Interface())
		}
	}
}
