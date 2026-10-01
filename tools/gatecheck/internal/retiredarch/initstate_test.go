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

// initFSOps names, per import path, the path-taking filesystem entry points.
// A relative path argument resolves against the working directory, so
// touching the filesystem during package initialization couples the binary to
// the launch cwd exactly as a Getwd read would; the list watches every
// path-taking entry point rather than proving each argument absolute.
var initFSOps = map[string]map[string]bool{
	"os": {"Open": true, "OpenFile": true, "OpenInRoot": true, "OpenRoot": true, "Create": true,
		"ReadFile": true, "WriteFile": true, "ReadDir": true, "DirFS": true, "CopyFS": true,
		"Stat": true, "Lstat": true, "Readlink": true, "Symlink": true, "Link": true,
		"Mkdir": true, "MkdirAll": true, "MkdirTemp": true, "CreateTemp": true,
		"Remove": true, "RemoveAll": true, "Rename": true, "Truncate": true,
		"Chmod": true, "Chown": true, "Lchown": true, "Chtimes": true},
	"io/ioutil":     {"ReadFile": true, "WriteFile": true, "ReadDir": true, "TempDir": true, "TempFile": true},
	"path/filepath": {"EvalSymlinks": true, "Glob": true, "Walk": true, "WalkDir": true},
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

// initFilesystemOps reports every initFSOps entry point that can run during
// package initialization in pkgs (import path to parsed files). It shares
// initProcessStateReads' roots (package-level var initializers and init
// bodies, including func literals inside them) and its fail-closed treatment
// of function and method values, with one deliberate narrowing: a function or
// method value passed as an argument to a function in pkgs is followed only
// when that callee may invoke it, that is, when the matching parameter is used
// anywhere other than as the right-hand side of a store into a map element or
// struct field. This lets the register_<name>.go init functions store each
// subcommand's run function, which is invoked from main at run time, without
// following every run path, while a callee that calls, forwards or returns the
// value is still followed. A value passed to a function outside pkgs is always
// followed, and once reachable code touches anything outside pkgs every method
// in pkgs is followed, as in initProcessStateReads. A watched selector counts
// only when its qualifier is bound to the watched import path; a dot import of
// a watched package is always reported.
func initFilesystemOps(fset *token.FileSet, pkgs map[string][]*ast.File) []string {
	type root struct {
		pkg  string
		file *ast.File
		node ast.Node
	}
	var (
		problems   []string
		roots      []root
		funcs      = map[string]initFunc{}
		methods    = map[string][]initFunc{}
		allMethods []initFunc
		declared   = map[string]map[string]bool{}
	)
	declare := func(pkg, name string) {
		if declared[pkg] == nil {
			declared[pkg] = map[string]bool{}
		}
		declared[pkg][name] = true
	}
	for pkg, files := range pkgs {
		for _, f := range files {
			for _, imp := range f.Imports {
				p := strings.Trim(imp.Path.Value, "`\"")
				if imp.Name != nil && imp.Name.Name == "." && initFSOps[p] != nil {
					problems = append(problems, fmt.Sprintf("%s: dot import of %q hides filesystem operations", fset.Position(imp.Pos()), p))
				}
			}
			for _, d := range f.Decls {
				switch v := d.(type) {
				case *ast.FuncDecl:
					if v.Recv == nil {
						declare(pkg, v.Name.Name)
					}
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
					for _, s := range v.Specs {
						switch s := s.(type) {
						case *ast.TypeSpec:
							declare(pkg, s.Name.Name)
						case *ast.ValueSpec:
							for _, n := range s.Names {
								declare(pkg, n.Name)
							}
							if v.Tok == token.VAR {
								for _, val := range s.Values {
									roots = append(roots, root{pkg, f, val})
								}
							}
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
	deferred := map[string][]root{}
	selected := map[string]bool{}
	reflective := false
	selectName := func(name string) {
		selected[name] = true
		roots = append(roots, deferred[name]...)
		delete(deferred, name)
	}
	reachReflection := func() {
		reflective = true
		for name := range deferred {
			selectName(name)
		}
	}
	for len(roots) > 0 {
		r := roots[0]
		roots = roots[1:]
		names := fileImportPaths(r.file)
		var dotPkgs []string
		externalDot := false
		for _, imp := range r.file.Imports {
			if imp.Name != nil && imp.Name.Name == "." {
				p := strings.Trim(imp.Path.Value, "`\"")
				if _, scanned := pkgs[p]; scanned {
					dotPkgs = append(dotPkgs, p)
				} else {
					externalDot = true
					reachExternal()
				}
			}
		}
		// fieldKey reports whether key can only name a struct field.
		fieldKey := func(key *ast.Ident) bool {
			if key.Obj != nil || externalDot {
				return false
			}
			switch key.Name {
			case "true", "false", "nil", "iota":
				return false
			}
			for _, p := range append([]string{r.pkg}, dotPkgs...) {
				if declared[p][key.Name] {
					return false
				}
			}
			return true
		}
		// callee resolves a call's function expression to a package-level
		// function in pkgs; ok is false for anything else (a method, a
		// function value, or a function outside pkgs).
		callee := func(fun ast.Expr) (fn initFunc, ok bool) {
			switch v := ast.Unparen(fun).(type) {
			case *ast.Ident:
				for _, p := range append([]string{r.pkg}, dotPkgs...) {
					if c, found := funcs[p+"."+v.Name]; found {
						return c, true
					}
				}
			case *ast.SelectorExpr:
				if id, isIdent := v.X.(*ast.Ident); isIdent {
					if imp := names[id.Name]; imp != "" {
						c, found := funcs[imp+"."+v.Sel.Name]
						return c, found
					}
				}
			}
			return initFunc{}, false
		}
		var visit func(ast.Node) bool
		visit = func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.CallExpr:
				c, scanned := callee(v.Fun)
				ast.Inspect(v.Fun, visit)
				for i, arg := range v.Args {
					if scanned && funcValueArg(arg) && !paramMayInvoke(c.decl, i) {
						continue
					}
					ast.Inspect(arg, visit)
				}
				return false
			case *ast.CompositeLit:
				if v.Type != nil {
					ast.Inspect(v.Type, visit)
				}
				for _, elt := range v.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						key, isIdent := kv.Key.(*ast.Ident)
						lit, isLit := ast.Unparen(kv.Value).(*ast.FuncLit)
						if isIdent && isLit && fieldKey(key) && !selected[key.Name] && !reflective {
							deferred[key.Name] = append(deferred[key.Name], root{r.pkg, r.file, lit})
							continue
						}
					}
					ast.Inspect(elt, visit)
				}
				return false
			case *ast.Ident:
				for _, p := range append([]string{r.pkg}, dotPkgs...) {
					if c, ok := funcs[p+"."+v.Name]; ok {
						follow(c)
					}
				}
			case *ast.SelectorExpr:
				follow(methods[v.Sel.Name]...)
				selectName(v.Sel.Name)
				imp := ""
				if id, ok := v.X.(*ast.Ident); ok {
					imp = names[id.Name]
				}
				switch {
				case imp != "":
					if initFSOps[imp][v.Sel.Name] {
						problems = append(problems, fmt.Sprintf("%s: %s.%s reachable from package initialization of %s",
							fset.Position(v.Pos()), imp, v.Sel.Name, r.pkg))
					}
					switch imp {
					case "reflect", "text/template", "html/template":
						reachReflection()
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

// funcValueArg reports whether arg can be a bare function or method value: an
// identifier or a selector (a package-qualified function, a method value or a
// field). Any other expression is visited normally.
func funcValueArg(arg ast.Expr) bool {
	switch ast.Unparen(arg).(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return true
	}
	return false
}

// paramMayInvoke reports whether fn may invoke the value passed as its i-th
// argument. It is false only when the parameter is named and every use of it
// in fn's body is the right-hand side of an assignment whose matching
// left-hand side is a map element or struct field, so the value is stored for
// later rather than called, forwarded or returned.
func paramMayInvoke(fn *ast.FuncDecl, i int) bool {
	var params []*ast.Ident
	variadic := false
	for _, field := range fn.Type.Params.List {
		if _, ok := field.Type.(*ast.Ellipsis); ok {
			variadic = true
		}
		if len(field.Names) == 0 {
			params = append(params, nil)
		}
		params = append(params, field.Names...)
	}
	if i >= len(params) {
		if !variadic || len(params) == 0 {
			return true
		}
		i = len(params) - 1
	}
	p := params[i]
	if p == nil || p.Name == "_" {
		return false
	}
	stored := map[*ast.Ident]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != len(as.Rhs) {
			return true
		}
		for j, rhs := range as.Rhs {
			id, isIdent := ast.Unparen(rhs).(*ast.Ident)
			if !isIdent || id.Name != p.Name {
				continue
			}
			switch as.Lhs[j].(type) {
			case *ast.IndexExpr, *ast.SelectorExpr:
				stored[id] = true
			}
		}
		return true
	})
	invoked := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == p.Name && !stored[id] {
			invoked = true
		}
		return !invoked
	})
	return invoked
}

// TestNoInitTimeFilesystemOps extends TestNoInitTimeProcessStateReads to the
// path-taking filesystem entry points in initFSOps: a relative path touched
// during package initialization resolves against the launch cwd. Every
// gatecheck package is scanned.
func TestNoInitTimeFilesystemOps(t *testing.T) {
	fset := token.NewFileSet()
	pkgs := readPackages(t, fset, filepath.Join("..", ".."), gatecheckImportPath)
	if len(pkgs[gatecheckImportPath]) == 0 || len(pkgs[gatecheckImportPath+"/internal/retiredarch"]) == 0 {
		t.Fatalf("scan found no gatecheck main or retiredarch sources (packages: %d)", len(pkgs))
	}
	for _, p := range initFilesystemOps(fset, pkgs) {
		t.Error(p)
	}
}

func TestInitFilesystemOps_Mutations(t *testing.T) {
	const a, b = "example.com/m/a", "example.com/m/b"
	cases := []struct {
		name string
		pkgs map[string][]string
		red  bool
	}{
		{"var initializer reads a relative file", map[string][]string{a: {`package a
import "os"
var data, _ = os.ReadFile("relative")`}}, true},
		{"init stats a path", map[string][]string{a: {`package a
import "os"
func init() { _, _ = os.Stat("x") }`}}, true},
		{"var initializer calls an opening helper", map[string][]string{a: {`package a
import "os"
var x = f()
func f() error { _, err := os.Open("x"); return err }`}}, true},
		{"aliased os import", map[string][]string{a: {`package a
import o "os"
var data, _ = o.ReadFile("x")`}}, true},
		{"ioutil read", map[string][]string{a: {`package a
import "io/ioutil"
var data, _ = ioutil.ReadFile("x")`}}, true},
		{"glob in init", map[string][]string{a: {`package a
import "path/filepath"
func init() { _, _ = filepath.Glob("*.go") }`}}, true},
		{"dot import of os", map[string][]string{a: {`package a
import . "os"
var _ = Getpid()`}}, true},
		{"cross-package call", map[string][]string{
			a: {`package a
import "example.com/m/b"
var x = b.F()`},
			b: {`package b
import "os"
func F() error { return os.MkdirAll("out", 0o755) }`}}, true},
		{"closure invoked in init", map[string][]string{a: {`package a
import "os"
func init() { func() { _ = os.Remove("x") }() }`}}, true},
		{"function-valued var invoked", map[string][]string{a: {`package a
import "os"
var read = load
var x = read()
func load() error { _, err := os.ReadFile("x"); return err }`}}, true},
		{"local function variable in init", map[string][]string{a: {`package a
import "os"
func init() { f := load; _ = f() }
func load() error { _, err := os.ReadFile("x"); return err }`}}, true},
		{"function argument called by callee", map[string][]string{a: {`package a
import "os"
var x = apply(load)
func apply(fn func() error) error { return fn() }
func load() error { _, err := os.ReadFile("x"); return err }`}}, true},
		{"function argument forwarded by callee", map[string][]string{a: {`package a
import "os"
var x = outer(load)
func outer(fn func() error) error { return apply(fn) }
func apply(fn func() error) error { return fn() }
func load() error { _, err := os.ReadFile("x"); return err }`}}, true},
		{"function argument returned by callee", map[string][]string{a: {`package a
import "os"
var g = wrap(load)
var x = g()
func wrap(fn func() error) func() error { return fn }
func load() error { _, err := os.ReadFile("x"); return err }`}}, true},
		{"function argument to variadic callee", map[string][]string{a: {`package a
import "os"
var x = all(nil, load)
func all(_ []int, fns ...func() error) error { return fns[0]() }
func load() error { _, err := os.ReadFile("x"); return err }`}}, true},
		{"function passed to external code", map[string][]string{a: {`package a
import ("os"; "strings")
var x = strings.Map(f, "abc")
func f(r rune) rune { _, _ = os.Stat("x"); return r }`}}, true},
		{"method value passed to invoking callee", map[string][]string{a: {`package a
import "os"
type T struct{}
func (T) m() error { _, err := os.ReadFile("x"); return err }
var x = apply(T{}.m)
func apply(fn func() error) error { return fn() }`}}, true},
		{"init calls a reading method", map[string][]string{a: {`package a
import "os"
type T struct{}
func (T) m() { _, _ = os.Lstat("x") }
func init() { T{}.m() }`}}, true},
		{"function stored in a field read at init", map[string][]string{a: {`package a
import "os"
type S struct{ f func() error }
var s = S{f: load}
var x = s.f()
func load() error { _, err := os.ReadFile("x"); return err }`}}, true},
		{"func literal field invoked at init", map[string][]string{a: {`package a
import "os"
type S struct{ f func() error }
var s = S{f: func() error { _, err := os.ReadFile("x"); return err }}
var x = s.f()`}}, true},
		{"func literal field selected before the literal", map[string][]string{a: {`package a
import "os"
type S struct{ f func() }
func init() { g := s.f; g() }
var s = S{f: func() { _ = os.Remove("x") }}`}}, true},
		{"func literal field selected in a later file", map[string][]string{a: {`package a
import "os"
type S struct{ f func() }
var s = []S{{f: func() { _ = os.Remove("x") }}}`, `package a
func init() { run() }
func run() { s[0].f() }`}}, true},
		{"func literal element in a slice", map[string][]string{a: {`package a
import "os"
var fns = []func(){func() { _ = os.Remove("x") }}`}}, true},
		{"func literal map value under a local const key", map[string][]string{a: {`package a
import "os"
func init() {
	const k = "k"
	m := map[string]func(){k: func() { _ = os.Remove("x") }}
	m[k]()
}`}}, true},
		{"func literal map value under a package const key", map[string][]string{a: {`package a
import "os"
var m = map[string]func(){k: func() { _ = os.Remove("x") }}`, `package a
const k = "k"
func init() { m[k]() }`}}, true},
		{"func literal map value under a universe key", map[string][]string{a: {`package a
import "os"
var m = map[bool]func(){true: func() { _ = os.Remove("x") }}
func init() { m[true]() }`}}, true},
		{"func literal field reached through reflection", map[string][]string{a: {`package a
import ("os"; "reflect")
type S struct{ F func() }
var s = S{F: func() { _ = os.Remove("x") }}
var v = reflect.ValueOf(s).Field(0)`}}, true},
		{"func literal field reached through a template", map[string][]string{a: {`package a
import ("os"; "text/template")
type S struct{ F func() string }
var s = S{F: func() string { _ = os.Remove("x"); return "" }}
var t = template.Must(template.New("t").Parse("{{call .F}}"))`}}, true},
		{"func literal stored in a field", map[string][]string{a: {`package a
import "os"
type S struct{ f func() error }
var s = []S{{f: func() error { _, err := os.ReadFile("x"); return err }}}`}}, false},
		{"registered function reads a file", map[string][]string{a: {`package a
import "os"
var cmds = map[string]func() error{}
func init() { register("x", run) }
func register(name string, fn func() error) { cmds[name] = fn }
func run() error { _, err := os.ReadFile("x"); return err }`}}, false},
		{"registered method value reads a file", map[string][]string{a: {`package a
import "os"
type T struct{}
func (T) run() error { _, err := os.ReadFile("x"); return err }
var cmds = map[string]func() error{}
func init() { register("x", T{}.run) }
func register(name string, fn func() error) { cmds[name] = fn }`}}, false},
		{"reads only in main", map[string][]string{a: {`package a
import "os"
func main() { _, _ = os.ReadFile("x") }`}}, false},
		{"initializer calls a clean func", map[string][]string{a: {`package a
import "os"
var x = f()
func f() int { return os.Getpid() }`}}, false},
		{"unrelated package named os", map[string][]string{a: {`package a
import os "example.com/other"
var x = os.ReadFile("x")`}}, false},
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
			got := initFilesystemOps(fset, pkgs)

			if red := len(got) > 0; red != tc.red {
				t.Errorf("red = %v (%q), want %v", red, got, tc.red)
			}
		})
	}
}
