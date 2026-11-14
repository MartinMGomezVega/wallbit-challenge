// Package archtest verifica la regla de dependencias de Clean Architecture
// (architecture.md §2.2–2.3) parseando imports con go/parser.
package archtest

import (
	"go/parser"
	"go/scanner"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Violation describe una violación de arquitectura.
type Violation struct {
	File string
	Rule string
}

var forbiddenImports = map[string][]string{
	"internal/domain":      {"internal/application", "internal/infrastructure", "internal/interfaces"},
	"internal/application": {"internal/infrastructure", "internal/interfaces"},
}

// Check recorre root (directorio del módulo) y devuelve las violaciones.
func Check(root, module string) ([]Violation, error) {
	var out []Violation
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		vs, err := checkFile(path, rel, module)
		out = append(out, vs...)
		return err
	})
	return out, err
}

func layerOf(rel string) string {
	for layer := range forbiddenImports {
		if strings.HasPrefix(rel, layer+"/") {
			return layer
		}
	}
	return ""
}

func checkFile(path, rel, module string) ([]Violation, error) {
	var out []Violation
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	if layer := layerOf(rel); layer != "" {
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			for _, bad := range forbiddenImports[layer] {
				if p == module+"/"+bad || strings.HasPrefix(p, module+"/"+bad+"/") {
					out = append(out, Violation{File: rel, Rule: layer + " no puede importar " + p})
				}
			}
		}
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	isTest := strings.HasSuffix(rel, "_test.go")
	if strings.HasPrefix(rel, "internal/domain/") && hasToken(src, "float64") {
		out = append(out, Violation{File: rel, Rule: "float64 prohibido en internal/domain"})
	}
	if !isTest && !strings.HasPrefix(rel, "internal/infrastructure/clock/") && hasSelector(src, "time", "Now") {
		out = append(out, Violation{File: rel, Rule: "time.Now() solo en infrastructure/clock"})
	}
	return out, nil
}

func hasToken(src []byte, ident string) bool {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, nil, 0)
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			return false
		}
		if tok == token.IDENT && lit == ident {
			return true
		}
	}
}

func hasSelector(src []byte, pkg, sel string) bool {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, nil, 0)
	var prev2, prev1 string
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			return false
		}
		cur := lit
		if tok == token.PERIOD {
			cur = "."
		}
		if prev2 == pkg && prev1 == "." && tok == token.IDENT && cur == sel {
			return true
		}
		prev2, prev1 = prev1, cur
	}
}
