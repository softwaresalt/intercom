package retiredarch

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// gatecheckImportPath is the import path of the gatecheck tree, the prefix
// TestNoInitTimeProcessStateReads gives each scanned package.
const gatecheckImportPath = "github.com/softwaresalt/intercom-go/tools/gatecheck"

// initStateReads names, per import path, the package-level identifiers whose
// use reads (or rebinds) process argv or the working directory.
var initStateReads = map[string]map[string]bool{
	"os":            {"Args": true, "Getwd": true, "Chdir": true},
	"path/filepath": {"Abs": true},
	"flag": {"Parse": true, "Parsed": true, "Args": true, "Arg": true,
		"NArg": true, "NFlag": true, "CommandLine": true},
}

// initFunc is one package-level function or method declaration together with
// the file (for import resolution) and package it belongs to.
type initFunc struct {
	pkg  string
	file *ast.File
	decl *ast.FuncDecl
}

// initProcessStateReads reports every read of process argv or the working
// directory that can run during package initialization in pkgs (import path
// to parsed files). Roots are package-level var initializers and init
// function bodies, including any func literal inside them. Reachability fails
// closed: every package-level function a reachable node references, whether
// called or used as a value (assigned to a variable or field, stored in a map,
// passed as an argument, or registered, as the register_<name>.go init
// functions register each subcommand's run function), is treated as invoked,
// as is every method in pkgs sharing a reachable selector's name, because
// neither the eventual call site nor the receiver type is tracked. A selector
// on an import bound to another package in pkgs follows that function; once
// reachable code touches anything outside pkgs (an unscanned import, a dot
// import of one, or a selector no scanned method matches), every method in
// pkgs is followed, because external code such as fmt or sort can invoke
// methods through interfaces with no selector in the scanned source. A watched
// selector counts only when its qualifier is bound to the watched import path;
// a dot import of a watched package is always reported because its
// identifiers would carry no qualifier. Reflection, unsafe and go:linkname are
// out of scope; the gatecheck tree uses none of them.
func initProcessStateReads(fset *token.FileSet, pkgs map[string][]*ast.File) []string {
	type root struct {
		pkg  string
		file *ast.File
		node ast.Node
	}
	var (
		problems   []string
		roots      []root
		funcs      = map[string]initFunc{}   // import path + "." + name
		methods    = map[string][]initFunc{} // method name
		allMethods []initFunc
	)
	for pkg, files := range pkgs {
		for _, f := range files {
			for _, imp := range f.Imports {
				p := strings.Trim(imp.Path.Value, "`\"")
				if imp.Name != nil && imp.Name.Name == "." && initStateReads[p] != nil {
					problems = append(problems, fmt.Sprintf("%s: dot import of %q hides process-state reads", fset.Position(imp.Pos()), p))
				}
			}
			for _, d := range f.Decls {
				switch v := d.(type) {
				case *ast.FuncDecl:
					fn := initFunc{pkg: pkg, file: f, decl: v}
					switch {
					case v.Recv != nil:
						methods[v.Name.Name] = append(methods[v.Name.Name], fn)
						allMethods = append(allMethods, fn)
					case v.Name.Name == "init":
						if v.Body != nil {
							roots = append(roots, root{pkg, f, v.Body})
						}
					default:
						funcs[pkg+"."+v.Name.Name] = fn
					}
				case *ast.GenDecl:
					if v.Tok != token.VAR {
						continue
					}
					for _, s := range v.Specs {
						for _, val := range s.(*ast.ValueSpec).Values {
							roots = append(roots, root{pkg, f, val})
						}
					}
				}
			}
		}
	}
	seen := map[*ast.FuncDecl]bool{}
	follow := func(fns ...initFunc) {
		for _, c := range fns {
			if !seen[c.decl] && c.decl.Body != nil {
				seen[c.decl] = true
				roots = append(roots, root{c.pkg, c.file, c.decl.Body})
			}
		}
	}
	external := false
	reachExternal := func() {
		if !external {
			external = true
			follow(allMethods...)
		}
	}
	for len(roots) > 0 {
		r := roots[0]
		roots = roots[1:]
		names := fileImportPaths(r.file)
		var dotPkgs []string
		for _, imp := range r.file.Imports {
			if imp.Name != nil && imp.Name.Name == "." {
				p := strings.Trim(imp.Path.Value, "`\"")
				if _, scanned := pkgs[p]; scanned {
					dotPkgs = append(dotPkgs, p)
				} else {
					reachExternal()
				}
			}
		}
		var visit func(ast.Node) bool
		visit = func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.Ident:
				for _, p := range append([]string{r.pkg}, dotPkgs...) {
					if c, ok := funcs[p+"."+v.Name]; ok {
						follow(c)
					}
				}
			case *ast.SelectorExpr:
				follow(methods[v.Sel.Name]...)
				imp := ""
				if id, ok := v.X.(*ast.Ident); ok {
					imp = names[id.Name]
				}
				switch {
				case imp != "":
					if initStateReads[imp][v.Sel.Name] {
						problems = append(problems, fmt.Sprintf("%s: %s.%s reachable from package initialization of %s",
							fset.Position(v.Pos()), imp, v.Sel.Name, r.pkg))
					}
					if _, scanned := pkgs[imp]; !scanned {
						reachExternal()
					} else if c, ok := funcs[imp+"."+v.Sel.Name]; ok {
						follow(c)
					}
				case len(methods[v.Sel.Name]) == 0:
					reachExternal()
				}
				ast.Inspect(v.X, visit)
				return false
			}
			return true
		}
		ast.Inspect(r.node, visit)
	}
	sort.Strings(problems)
	return problems
}

// fileImportPaths maps each usable file-local import name in f to its import
// path; blank and dot imports are omitted.
func fileImportPaths(f *ast.File) map[string]string {
	names := map[string]string{}
	for _, imp := range f.Imports {
		p := strings.Trim(imp.Path.Value, "`\"")
		name := path.Base(p)
		if imp.Name != nil {
			name = imp.Name.Name
		}
		if name != "_" && name != "." {
			names[name] = p
		}
	}
	return names
}

// readPackages parses the non-test Go files below dir, keyed by import path
// (prefix plus the slash-separated directory relative to dir). testdata
// directories are skipped.
func readPackages(t *testing.T, fset *token.FileSet, dir, prefix string) map[string][]*ast.File {
	t.Helper()
	pkgs := map[string][]*ast.File{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		f, err := parser.ParseFile(fset, p, src, 0)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, filepath.Dir(p))
		if err != nil {
			return err
		}
		key := prefix
		if rel != "." {
			key = path.Join(prefix, filepath.ToSlash(rel))
		}
		pkgs[key] = append(pkgs[key], f)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return pkgs
}

// TestNoInitTimeProcessStateReads is the Go counterpart of the retired Python
// test_import_with_empty_argv_and_foreign_cwd and
// test_no_module_level_argv_or_cwd_reads. Go runs package-level var
// initializers and init functions at program start, before run() receives its
// arguments, so a read of os.Args or the working directory there would
// recreate the import-time coupling those tests prevented. Every gatecheck
// package is scanned.
func TestNoInitTimeProcessStateReads(t *testing.T) {
	fset := token.NewFileSet()
	pkgs := readPackages(t, fset, filepath.Join("..", ".."), gatecheckImportPath)
	if len(pkgs[gatecheckImportPath]) == 0 || len(pkgs[gatecheckImportPath+"/internal/retiredarch"]) == 0 {
		t.Fatalf("scan found no gatecheck main or retiredarch sources (packages: %d)", len(pkgs))
	}
	inits := 0
	for _, files := range pkgs {
		for _, f := range files {
			for _, d := range f.Decls {
				if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name == "init" {
					inits++
				}
			}
		}
	}
	if inits == 0 {
		t.Fatal("scan found no init functions; the register_<name>.go roots are missing")
	}
	for _, p := range initProcessStateReads(fset, pkgs) {
		t.Error(p)
	}
}

func TestInitProcessStateReads_Mutations(t *testing.T) {
	const a, b = "example.com/m/a", "example.com/m/b"
	cases := []struct {
		name string
		pkgs map[string][]string
		red  bool
	}{
		{"var initializer reads cwd", map[string][]string{a: {`package a
import "os"
var cwd, _ = os.Getwd()`}}, true},
		{"init reads argv", map[string][]string{a: {`package a
import "os"
func init() { _ = os.Args }`}}, true},
		{"aliased os import", map[string][]string{a: {`package a
import o "os"
var argv = o.Args`}}, true},
		{"var initializer calls a reading func", map[string][]string{a: {`package a
import "os"
var x = f()
func f() []string { return os.Args }`}}, true},
		{"reading func in another file", map[string][]string{a: {`package a
func init() { g() }`, `package a
import "path/filepath"
func g() { _, _ = filepath.Abs(".") }`}}, true},
		{"init calls a reading method", map[string][]string{a: {`package a
import "os"
type T struct{}
func (T) m() { _ = os.Chdir("/") }
func init() { T{}.m() }`}}, true},
		{"init parses flags", map[string][]string{a: {`package a
import "flag"
func init() { flag.Parse() }`}}, true},
		{"closure in init reads argv", map[string][]string{a: {`package a
import "os"
func init() { func() { _ = os.Args }() }`}}, true},
		{"parenthesized call", map[string][]string{a: {`package a
import "os"
var x = (f)()
func f() string { d, _ := os.Getwd(); return d }`}}, true},
		{"dot import of os", map[string][]string{a: {`package a
import . "os"
var _ = Getpid()`}}, true},
		{"cross-package call", map[string][]string{
			a: {`package a
import "example.com/m/b"
var x = b.F()`},
			b: {`package b
import "os"
func F() string { d, _ := os.Getwd(); return d }`}}, true},
		{"function-valued var initializer", map[string][]string{a: {`package a
import "os"
var read = getArgs
var args = read()
func getArgs() []string { return os.Args }`}}, true},
		{"local function variable in init", map[string][]string{a: {`package a
import "os"
func init() { f := getArgs; _ = f() }
func getArgs() []string { return os.Args }`}}, true},
		{"function argument called by callee", map[string][]string{a: {`package a
import "os"
var x = apply(getArgs)
func apply(fn func() []string) []string { return fn() }
func getArgs() []string { return os.Args }`}}, true},
		{"method value", map[string][]string{a: {`package a
import "os"
type T struct{}
func (T) m() []string { return os.Args }
var mv = T{}.m
var x = mv()`}}, true},
		{"function-valued field", map[string][]string{a: {`package a
import "os"
type S struct{ f func() []string }
var s = S{f: getArgs}
var x = s.f()
func getArgs() []string { return os.Args }`}}, true},
		{"map of functions", map[string][]string{a: {`package a
import "os"
var m = map[string]func() []string{"a": getArgs}
var x = m["a"]()
func getArgs() []string { return os.Args }`}}, true},
		{"function passed to external code", map[string][]string{a: {`package a
import ("os"; "strings")
var x = strings.Map(f, "abc")
func f(r rune) rune { _ = os.Args; return r }`}}, true},
		{"method dispatched by external code", map[string][]string{a: {`package a
import ("fmt"; "os")
type T struct{}
func (T) String() string { d, _ := os.Getwd(); return d }
var s = fmt.Sprint(T{})`}}, true},
		{"method dispatched by dot-imported external code", map[string][]string{a: {`package a
import (. "fmt"; "os")
type T struct{}
func (T) String() string { d, _ := os.Getwd(); return d }
var s = Sprint(T{})`}}, true},
		{"registered function reads argv", map[string][]string{a: {`package a
import "os"
func init() { register(run) }
func register(func() []string) {}
func run() []string { return os.Args }`}}, true},
		{"reads only in main", map[string][]string{a: {`package a
import "os"
func init() { register(run) }
func register(func([]string) int) {}
func run(args []string) int { return len(args) }
func main() { _ = run(os.Args[1:]) }`}}, false},
		{"unreferenced function reads argv", map[string][]string{a: {`package a
import "os"
var read = clean
var x = read()
func clean() int { return 1 }
func later() []string { return os.Args }`}}, false},
		{"initializer calls a clean func", map[string][]string{a: {`package a
import "os"
var x = f()
func f() int { return os.Getpid() }`}}, false},
		{"unrelated package named os", map[string][]string{a: {`package a
import os "example.com/other"
var x = os.Args`}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			pkgs := map[string][]*ast.File{}
			keys := make([]string, 0, len(tc.pkgs))
			for k := range tc.pkgs {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				for i, src := range tc.pkgs[k] {
					f, err := parser.ParseFile(fset, path.Join(k, string(rune('a'+i))+".go"), src, 0)
					if err != nil {
						t.Fatal(err)
					}
					pkgs[k] = append(pkgs[k], f)
				}
			}
			got := initProcessStateReads(fset, pkgs)
			if red := len(got) > 0; red != tc.red {
				t.Errorf("red = %v (%q), want %v", red, got, tc.red)
			}
		})
	}
}
