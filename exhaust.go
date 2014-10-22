package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"strings"

	"code.google.com/p/go.tools/go/exact"
	_ "code.google.com/p/go.tools/go/gcimporter"
	"code.google.com/p/go.tools/go/types"
)

type checker struct {
	fileSet     *token.FileSet
	files       []*ast.File
	info        *types.Info
	constValues map[types.Type][]exact.Value
}

func newChecker() *checker {
	return &checker{
		fileSet:     token.NewFileSet(),
		constValues: make(map[types.Type][]exact.Value),
		info: &types.Info{
			Defs:  make(map[*ast.Ident]types.Object),
			Types: make(map[ast.Expr]types.TypeAndValue),
		},
	}
}

// parser will parse code in r
func (c *checker) parse(r io.Reader) error {
	f, err := parser.ParseFile(c.fileSet, "thing.go", r, 0)
	if err != nil {
		return fmt.Errorf("parsing code: %v", err)
	}
	c.files = append(c.files, f)
	return nil
}

func (c *checker) populateConstValues() error {
	config := &types.Config{}
	_, err := config.Check("", c.fileSet, c.files, c.info)
	if err != nil {
		return fmt.Errorf("checking code: %v", err)
	}
	for _, o := range c.info.Defs {
		cnst, ok := o.(*types.Const)
		if !ok {
			continue
		}
		found := false
		for _, v := range c.constValues[cnst.Type()] {
			if v == cnst.Val() {
				// We've already seen this value
				found = true
				break
			}
		}
		if !found {
			c.constValues[cnst.Type()] = append(c.constValues[cnst.Type()], cnst.Val())
		}
	}
	return nil
}

// isExhaustive will check to make sure we have dealt with all possible values
// of the type t in switch s.
func (c *checker) isExhaustive(s *ast.SwitchStmt, t types.TypeAndValue) bool {
	// No values, nothing to check.
	if len(c.constValues[t.Type]) == 0 {
		return true
	}
	seen := make(map[exact.Value]bool)
	for _, stmt := range s.Body.List {
		switch ss := stmt.(type) {
		case *ast.CaseClause:
			if ss.List == nil {
				return true // We have a default clause so we've covered everything
			}
			for _, e := range ss.List {
				seen[c.info.Types[e].Value] = true
			}
		}
	}
	if len(seen) == 0 {
		return false
	}
	for _, v := range c.constValues[t.Type] {
		if !seen[v] {
			return false // TODO: Detailed error
		}
	}
	return true
}

func (c *checker) allExhaustive() (bool, map[*ast.SwitchStmt]bool) {
	switches := make(map[*ast.SwitchStmt]bool)
	exhaustive := true
	for _, f := range c.files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch s := n.(type) {
			case *ast.SwitchStmt:
				if s.Tag == nil {
					break
				}
				tagType := c.info.Types[s.Tag]
				// We have a tag. Is it a constant we know about?
				for t, _ := range c.constValues {
					if tagType.Type == t {
						e := c.isExhaustive(s, tagType)
						if !e {
							exhaustive = false
							switches[s] = false
							return true
						}
					}
				}
			}
			return true
		})
		if !exhaustive {
			return false, switches // Stop checking more files, we already know we're not exhaustive
		}
	}
	return exhaustive, switches
}

func (c *checker) positionString(n ast.Node) string {
	return c.fileSet.Position(n.Pos()).String()
}

func main() {
	c := newChecker()
	err := c.parse(strings.NewReader(validCode))
	if err != nil {
		log.Fatal(err)
	}
	err = c.populateConstValues()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(c.constValues)
	log.Println()
}
