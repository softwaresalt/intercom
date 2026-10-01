package retiredarch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// writePathMaskFlow reports whether file f declares writepath's scanFile and
// scanText and, between them, the result of gomask.MaskGoNonCode reaches the
// scan loop:
//
//   - scanText ranges, in a top-level statement, over pysem.SplitLines of its
//     second parameter (pysem qualifier bound to pysemImportPath), and never
//     writes, redeclares, or takes the address of that parameter;
//   - scanFile calls scanText at least once, every such call passes a direct
//     gomask.MaskGoNonCode call (qualifier bound to gomaskImportPath) as the
//     second argument, and scanFile declares nothing named scanText;
//   - no function in f other than scanFile calls scanText, so no production
//     path scans text that skipped the canonical mask.
//
// It is the writepath half of the retired single-definition and canonical-
// consumer assertions; TestScanGoConsumesCanonicalMask is the retiredarch
// half. A package-level import or call count cannot prove this, because
// scanFile could switch to a local clone while a dead helper keeps a
// canonical call.
func writePathMaskFlow(f *ast.File) (declared, flows bool) {
	pkg := gomaskName(f)
	ps := importName(f, pysemImportPath, "pysem")
	var scanFile, scanText *ast.FuncDecl
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Body == nil {
			continue
		}
		switch fn.Name.Name {
		case "scanFile":
			scanFile = fn
		case "scanText":
			scanText = fn
		}
	}
	if scanFile == nil || scanText == nil {
		return false, false
	}
	if pkg == "" || ps == "" {
		return true, false
	}
	return true, scanTextRangesParam(scanText, ps) && scanFileMasks(f, scanFile, pkg)
}

// scanTextRangesParam reports whether fn has a second parameter that a
// top-level `range ps.SplitLines(param)` statement consumes and that fn never
// writes, redeclares in a nested scope, or writes from a closure.
func scanTextRangesParam(fn *ast.FuncDecl, ps string) bool {
	var params []string
	for _, fld := range fn.Type.Params.List {
		for _, id := range fld.Names {
			params = append(params, id.Name)
		}
	}
	if len(params) < 2 || params[1] == "_" {
		return false
	}
	p := params[1]
	if len(writePositions(fn.Body, p)) > 0 || closureWrites(fn.Body, p) || nestedDecl(fn.Body, p) || declaresInBody(fn.Body, p) {
		return false
	}
	for _, st := range fn.Body.List {
		rs, ok := st.(*ast.RangeStmt)
		if !ok || !isSelectorCall(rs.X, ps, "SplitLines") {
			continue
		}
		if arg, ok := rs.X.(*ast.CallExpr).Args[0].(*ast.Ident); ok && arg.Name == p {
			return true
		}
	}
	return false
}

// scanFileMasks reports whether every call to scanText in f sits inside
// scanFile and passes a direct pkg.MaskGoNonCode call as its second argument,
// at least one such call exists, and scanFile declares nothing named
// scanText that could shadow the package function.
func scanFileMasks(f *ast.File, scanFile *ast.FuncDecl, pkg string) bool {
	if declaresInBody(scanFile.Body, "scanText") || fieldsDeclare(scanFile.Type.Params, "scanText") ||
		fieldsDeclare(scanFile.Type.Results, "scanText") {
		return false
	}
	inside, ok := 0, true
	ast.Inspect(f, func(n ast.Node) bool {
		call, isCall := n.(*ast.CallExpr)
		if !isCall {
			return true
		}
		id, isIdent := ast.Unparen(call.Fun).(*ast.Ident)
		if !isIdent || id.Name != "scanText" {
			return true
		}
		if call.Pos() < scanFile.Body.Pos() || call.End() > scanFile.Body.End() {
			ok = false
			return true
		}
		if len(call.Args) != 2 || !isSelectorCall(call.Args[1], pkg, "MaskGoNonCode") {
			ok = false
			return true
		}
		inside++
		return true
	})
	return ok && inside > 0
}

// declaresInBody reports whether body declares name with :=, var, or a
// range short declaration at any depth, or a func literal parameter names it.
func declaresInBody(body *ast.BlockStmt, name string) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			if s.Tok == token.DEFINE {
				for _, l := range s.Lhs {
					if id, ok := l.(*ast.Ident); ok && id.Name == name {
						found = true
					}
				}
			}
		case *ast.RangeStmt:
			if s.Tok == token.DEFINE {
				for _, e := range []ast.Expr{s.Key, s.Value} {
					if id, ok := e.(*ast.Ident); ok && id.Name == name {
						found = true
					}
				}
			}
		case *ast.ValueSpec:
			for _, id := range s.Names {
				if id.Name == name {
					found = true
				}
			}
		case *ast.FuncLit:
			if fieldsDeclare(s.Type.Params, name) || fieldsDeclare(s.Type.Results, name) {
				found = true
			}
		}
		return !found
	})
	return found
}

// fieldsDeclare reports whether fl names a field called name.
func fieldsDeclare(fl *ast.FieldList, name string) bool {
	if fl == nil {
		return false
	}
	for _, fld := range fl.List {
		for _, id := range fld.Names {
			if id.Name == name {
				return true
			}
		}
	}
	return false
}

// maskerDefinitions returns "path:Name" for every function or method declared
// in the given non-test sources whose lower-cased name contains both "mask"
// and "noncode", sorted. srcs maps a forward-slash path to its source.
func maskerDefinitions(srcs map[string]string) ([]string, error) {
	var defs []string
	fset := token.NewFileSet()
	for p, src := range srcs {
		f, err := parser.ParseFile(fset, p, src, 0)
		if err != nil {
			return nil, err
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := strings.ToLower(fn.Name.Name)
			if strings.Contains(name, "mask") && strings.Contains(name, "noncode") {
				defs = append(defs, p+":"+fn.Name.Name)
			}
		}
	}
	sort.Strings(defs)
	return defs, nil
}

// canonicalMaskerDef is the one masker definition the module may contain.
const canonicalMaskerDef = "internal/gomask/gomask.go:MaskGoNonCode"

// TestMaskerDefinedExactlyOnce is the Go analogue of the retired Python
// test_masker_is_defined_exactly_once, scoped to the whole tools/gatecheck
// module: exactly one non-test masker definition exists, and it is
// gomask.MaskGoNonCode. Importing gomask is compiler-enforced, but nothing
// else stops a package from growing a second, drifting masker.
func TestMaskerDefinedExactlyOnce(t *testing.T) {
	root := filepath.Join("..", "..")
	srcs := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != root && (d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		srcs[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) == 0 {
		t.Fatal("no Go sources found under tools/gatecheck")
	}
	defs, err := maskerDefinitions(srcs)
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 1 || defs[0] != canonicalMaskerDef {
		t.Errorf("masker definitions = %q, want exactly [%s]", defs, canonicalMaskerDef)
	}
}

// TestWritePathScanFileConsumesCanonicalMask pins the canonical masker to
// writepath's scan path, and proves no other writepath source declares a
// local masker or calls scanText.
func TestWritePathScanFileConsumesCanonicalMask(t *testing.T) {
	dir := filepath.Join("..", "writepath")
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	found := false
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		if _, locals := maskerUse(f); len(locals) > 0 {
			t.Errorf("%s: local masker %q shadows gomask.MaskGoNonCode", p, locals)
		}
		declared, flows := writePathMaskFlow(f)
		if declared {
			found = true
			if !flows {
				t.Errorf("%s: scanFile does not pass a gomask.MaskGoNonCode result to scanText's scan loop", p)
			}
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if id, ok := ast.Unparen(c.Fun).(*ast.Ident); ok && id.Name == "scanText" {
					t.Errorf("%s: scanText called outside the file that declares scanFile", p)
				}
			}
			return true
		})
	}
	if !found {
		t.Fatal("no writepath source declares both scanFile and scanText")
	}
}

// TestWritePathMaskFlow_Mutations proves writePathMaskFlow is red when the
// writepath scan path stops consuming the canonical mask, and green on the
// production shape.
func TestWritePathMaskFlow_Mutations(t *testing.T) {
	const imp = "import \"" + gomaskImportPath + "\"\n"
	const pimp = "import \"" + pysemImportPath + "\"\n"
	const st = "func scanText(relPath, maskedText string) []string {\n\tfor range pysem.SplitLines(maskedText) {\n\t}\n\treturn nil\n}\n"
	const sf = "func scanFile(root, relPath string) ([]string, error) {\n\ttext := read(root)\n\treturn scanText(relPath, gomask.MaskGoNonCode(text)), nil\n}\n"
	const head = "package p\n" + imp + pimp
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"production shape", head + st + sf, true},
		{"canonical import under an alias",
			"package p\nimport gm \"" + gomaskImportPath + "\"\n" + pimp + st + strings.Replace(sf, "gomask.", "gm.", 1), true},
		{"local clone instead of canonical mask", head + st + strings.Replace(sf, "gomask.MaskGoNonCode(text)", "clone(text)", 1) +
			"func dead(s string) { _ = gomask.MaskGoNonCode(s) }\n", false},
		{"unmasked text scanned", head + st + strings.Replace(sf, "gomask.MaskGoNonCode(text)", "text", 1), false},
		{"other package masker", head + st + strings.Replace(sf, "gomask.", "other.", 1), false},
		{"foreign package imported as gomask", "package p\nimport gomask \"example.com/other\"\n" + pimp + st + sf, false},
		{"canonical package not imported", "package p\n" + pimp + st + sf, false},
		{"foreign package imported as pysem", "package p\n" + imp + "import pysem \"example.com/other\"\n" + st + sf, false},
		{"mask bound to a variable first",
			head + st + "func scanFile(root, relPath string) ([]string, error) {\n\tm := gomask.MaskGoNonCode(read(root))\n\treturn scanText(relPath, m), nil\n}\n", false},
		{"second scanText call unmasked",
			head + st + "func scanFile(root, relPath string) ([]string, error) {\n\ttext := read(root)\n\t_ = scanText(relPath, text)\n\treturn scanText(relPath, gomask.MaskGoNonCode(text)), nil\n}\n", false},
		{"scanText called from another function", head + st + sf + "func raw(s string) []string { return scanText(\"x\", s) }\n", false},
		{"scanFile never calls scanText",
			head + st + "func scanFile(root, relPath string) ([]string, error) {\n\t_ = gomask.MaskGoNonCode(read(root))\n\treturn nil, nil\n}\n", false},
		{"scanText shadowed in scanFile",
			head + st + "func scanFile(root, relPath string) ([]string, error) {\n\tscanText := other\n\ttext := read(root)\n\treturn scanText(relPath, gomask.MaskGoNonCode(text)), nil\n}\n", false},
		{"scanText ranges over the path, not the text",
			head + strings.Replace(st, "SplitLines(maskedText)", "SplitLines(relPath)", 1) + sf, false},
		{"scanText overwrites its masked parameter",
			head + strings.Replace(st, "{\n\tfor", "{\n\tmaskedText = raw()\n\tfor", 1) + sf, false},
		{"scanText takes the masked parameter's address",
			head + strings.Replace(st, "{\n\tfor", "{\n\tclobber(&maskedText)\n\tfor", 1) + sf, false},
		{"scanText shadows the masked parameter",
			head + strings.Replace(st, "{\n\tfor", "{\n\tmaskedText := raw()\n\tfor", 1) + sf, false},
		{"scanText closure writes the masked parameter",
			head + strings.Replace(st, "{\n\tfor", "{\n\tf := func() { maskedText = raw() }\n\tf()\n\tfor", 1) + sf, false},
		{"scan loop not top-level",
			head + strings.Replace(strings.Replace(st, "\tfor range", "\t{\n\tfor range", 1), "\t}\n\treturn", "\t}\n\t}\n\treturn", 1) + sf, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), "x.go", tc.src, 0)
			if err != nil {
				t.Fatal(err)
			}
			declared, flows := writePathMaskFlow(f)
			if !declared {
				t.Fatal("scanFile/scanText not found")
			}
			if flows != tc.want {
				t.Errorf("writePathMaskFlow flows = %v, want %v", flows, tc.want)
			}
		})
	}
}

// TestMaskerDefinitions_Mutations proves maskerDefinitions sees a second or a
// relocated masker, so TestMaskerDefinedExactlyOnce is red on those trees.
func TestMaskerDefinitions_Mutations(t *testing.T) {
	const canon = "package gomask\nfunc MaskGoNonCode(s string) string { return s }\n"
	cases := []struct {
		name string
		srcs map[string]string
		want []string
	}{
		{"canonical only", map[string]string{"internal/gomask/gomask.go": canon}, []string{canonicalMaskerDef}},
		{"local clone in writepath", map[string]string{
			"internal/gomask/gomask.go":       canon,
			"internal/writepath/writepath.go": "package writepath\nfunc maskGoNonCode(s string) string { return s }\n",
		}, []string{canonicalMaskerDef, "internal/writepath/writepath.go:maskGoNonCode"}},
		{"case-variant clone", map[string]string{
			"internal/gomask/gomask.go": canon,
			"main.go":                   "package main\nfunc MaskGoNoncode(s string) string { return s }\n",
		}, []string{canonicalMaskerDef, "main.go:MaskGoNoncode"}},
		{"method clone", map[string]string{
			"internal/gomask/gomask.go": canon,
			"internal/x/x.go":           "package x\ntype T struct{}\nfunc (T) MaskGoNonCode(s string) string { return s }\n",
		}, []string{canonicalMaskerDef, "internal/x/x.go:MaskGoNonCode"}},
		{"masker moved", map[string]string{"internal/other/mask.go": strings.Replace(canon, "gomask", "other", 1)},
			[]string{"internal/other/mask.go:MaskGoNonCode"}},
		{"no masker", map[string]string{"internal/gomask/gomask.go": "package gomask\n"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := maskerDefinitions(tc.srcs)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("maskerDefinitions = %q, want %q", got, tc.want)
			}
			exactlyOnce := len(got) == 1 && got[0] == canonicalMaskerDef
			if exactlyOnce != (tc.name == "canonical only") {
				t.Errorf("exactly-once verdict = %v on %q", exactlyOnce, tc.name)
			}
		})
	}
}
