package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
)

type checker struct {
	fileSet     *token.FileSet
	files       []*ast.File
	info        *types.Info
	constValues map[types.Type][]constant.Value
}

func newChecker() *checker {
	return &checker{
		fileSet:     token.NewFileSet(),
		constValues: make(map[types.Type][]constant.Value),
		info: &types.Info{
			Defs:  make(map[*ast.Ident]types.Object),
			Types: make(map[ast.Expr]types.TypeAndValue),
		},
	}
}

// parser will parse code in r
func (c *checker) parse(r io.Reader) error {
	f, err := parser.ParseFile(c.fileSet, "set.go", r, 0)
	if err != nil {
		return fmt.Errorf("parsing code: %v", err)
	}
	c.files = append(c.files, f)
	return nil
}

func (c *checker) parseDir(path string) error {
	pkgs, err := parser.ParseDir(c.fileSet, path, nil, 0)
	if err != nil {
		return fmt.Errorf("parsing dir: %v", err)
	}
	for _, p := range pkgs {
		for i := range p.Files {
			c.files = append(c.files, p.Files[i])
		}
	}
	return nil
}

func (c *checker) populateConstValues() error {
	config := &types.Config{
		Importer:    importer.Default(),
		FakeImportC: true,
	}
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
func (c *checker) isExhaustive(s *ast.SwitchStmt, t types.TypeAndValue) (bool, []constant.Value) {
	// No values, nothing to check.
	if len(c.constValues[t.Type]) == 0 {
		return true, nil
	}
	seen := make(map[constant.Value]bool)
	for _, stmt := range s.Body.List {
		switch ss := stmt.(type) {
		case *ast.CaseClause:
			if ss.List == nil {
				return true, nil // We have a default clause so we've covered everything
			}
			for _, e := range ss.List {
				seen[c.info.Types[e].Value] = true
			}
		}
	}
	var missing []constant.Value
	for i, v := range c.constValues[t.Type] {
		if seen[v] {
			continue
		}
		missing = append(missing, c.constValues[t.Type][i])
	}
	if len(missing) == 0 {
		return true, nil
	}
	return false, missing
}

type result struct {
	stmt    *ast.SwitchStmt
	missing []constant.Value
}

func (r result) exhaustive() bool {
	return len(r.missing) == 0
}

func (c *checker) allExhaustive() (bool, []result) {
	var switches []result
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
						e, missing := c.isExhaustive(s, tagType)
						if !e {
							exhaustive = false
							switches = append(switches, result{
								stmt:    s,
								missing: missing,
							})
							return true
						}
					}
				}
			}
			return true
		})
	}
	return exhaustive, switches
}

func (c *checker) positionString(n ast.Node) string {
	return c.fileSet.Position(n.Pos()).String()
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Expected goexhaust path")
		return
	}
	path := os.Args[1]

	c := newChecker()

	err := c.parseDir(path)
	if err != nil {
		log.Fatal(err)
	}
	err = c.populateConstValues()
	if err != nil {
		log.Fatal(err)
	}
	ok, switches := c.allExhaustive()
	if ok {
		return
	}
	for _, s := range switches {
		if s.exhaustive() {
			continue
		}
		p := c.fileSet.Position(s.stmt.Pos())
		log.Printf("Found non exhaustive switch: %v\n", p)
		for _, m := range s.missing {
			log.Printf("Missing: %s\n", m.ExactString())
		}
	}
}
