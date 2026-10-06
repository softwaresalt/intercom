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

// writePathMaskFlow verifies the canonical mask reaches scanSource's line
// scan through unmasked ReadText output.
func writePathMaskFlow(f *ast.File) (declared, flows bool) {
	pkg := gomaskName(f)
	ps := importName(f, pysemImportPath, "pysem")
	var scanFile, scanSource *ast.FuncDecl
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Body == nil {
			continue
		}
		switch fn.Name.Name {
		case "scanFile":
			scanFile = fn
		case "scanSource":
			scanSource = fn
		}
	}
	if scanFile == nil || scanSource == nil {
		return false, false
	}
	if pkg == "" || ps == "" {
		return true, false
	}
	return true, scanSourceConsumesMask(f, scanFile, scanSource, pkg, ps)
}

func scanSourceConsumesMask(f *ast.File, scanFile, scanSource *ast.FuncDecl, pkg, ps string) bool {
	var params []string
	for _, fld := range scanSource.Type.Params.List {
		for _, id := range fld.Names {
			params = append(params, id.Name)
		}
	}
	if len(params) < 2 || params[1] == "_" {
		return false
	}
	srcParam := params[1]
	if len(writePositions(scanSource.Body, srcParam)) > 0 || closureWrites(scanSource.Body, srcParam) ||
		nestedDecl(scanSource.Body, srcParam) || declaresInBody(scanSource.Body, srcParam) {
		return false
	}

	maskSelectors := map[*ast.SelectorExpr]bool{}
	var maskCall *ast.CallExpr
	maskCalls := 0
	ast.Inspect(scanSource.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isSelectorCall(call, pkg, "MaskGoNonCode") {
			return true
		}
		maskSelectors[call.Fun.(*ast.SelectorExpr)] = true
		maskCalls++
		if arg, ok := call.Args[0].(*ast.Ident); ok && arg.Name == srcParam {
			maskCall = call
		}
		return true
	})
	if maskCalls != 1 || maskCall == nil || !onlyListedSelectors(scanSource.Body, pkg, "MaskGoNonCode", maskSelectors) {
		return false
	}

	maskedName := ""
	maskStatement := -1
	for i, st := range scanSource.Body.List {
		assign, ok := st.(*ast.AssignStmt)
		if !ok || len(assign.Rhs) != 1 || assign.Rhs[0] != maskCall || len(assign.Lhs) != 1 || assign.Tok != token.DEFINE {
			continue
		}
		id, ok := assign.Lhs[0].(*ast.Ident)
		if ok && id.Name != "_" {
			maskedName = id.Name
			maskStatement = i
		}
	}
	if maskedName == "" || len(writePositions(scanSource.Body, maskedName)) != 1 ||
		closureWrites(scanSource.Body, maskedName) || nestedDecl(scanSource.Body, maskedName) {
		return false
	}

	splitSelectors := map[*ast.SelectorExpr]bool{}
	var splitCall *ast.CallExpr
	splitCalls := 0
	ast.Inspect(scanSource.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isSelectorCall(call, ps, "SplitLines") {
			return true
		}
		splitSelectors[call.Fun.(*ast.SelectorExpr)] = true
		splitCalls++
		if arg, ok := call.Args[0].(*ast.Ident); ok && arg.Name == maskedName {
			splitCall = call
		}
		return true
	})
	if splitCalls != 1 || splitCall == nil || !onlyListedSelectors(scanSource.Body, ps, "SplitLines", splitSelectors) {
		return false
	}

	lineName := ""
	splitStatement := -1
	for i, st := range scanSource.Body.List {
		assign, ok := st.(*ast.AssignStmt)
		if !ok || len(assign.Rhs) != 1 || assign.Rhs[0] != splitCall || len(assign.Lhs) != 1 || assign.Tok != token.DEFINE {
			continue
		}
		id, ok := assign.Lhs[0].(*ast.Ident)
		if ok && id.Name != "_" {
			lineName = id.Name
			splitStatement = i
		}
	}
	if lineName == "" || maskStatement >= splitStatement || len(writePositions(scanSource.Body, lineName)) != 1 ||
		closureWrites(scanSource.Body, lineName) || nestedDecl(scanSource.Body, lineName) {
		return false
	}
	hasLineRange := false
	for i, st := range scanSource.Body.List {
		rs, ok := st.(*ast.RangeStmt)
		if !ok || i <= splitStatement {
			continue
		}
		if id, ok := rs.X.(*ast.Ident); ok && id.Name == lineName {
			hasLineRange = true
		}
	}
	if !hasLineRange {
		return false
	}

	if declaresInBody(scanFile.Body, "scanSource") || fieldsDeclare(scanFile.Type.Params, "scanSource") ||
		fieldsDeclare(scanFile.Type.Results, "scanSource") ||
		!onlyListedSelectors(scanFile.Body, pkg, "MaskGoNonCode", map[*ast.SelectorExpr]bool{}) {
		return false
	}
	readNames := map[string]bool{}
	readSelectors := map[*ast.SelectorExpr]bool{}
	ast.Inspect(scanFile.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Rhs) != 1 || !isSelectorCall(assign.Rhs[0], ps, "ReadText") {
			return true
		}
		call := assign.Rhs[0].(*ast.CallExpr)
		readSelectors[call.Fun.(*ast.SelectorExpr)] = true
		if len(assign.Lhs) > 0 {
			if id, ok := assign.Lhs[0].(*ast.Ident); ok {
				readNames[id.Name] = true
			}
		}
		return true
	})
	if len(readNames) == 0 || !onlyListedSelectors(scanFile.Body, ps, "ReadText", readSelectors) {
		return false
	}

	allowedRefs := map[*ast.Ident]bool{scanSource.Name: true}
	calls := 0
	ast.Inspect(scanFile.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok || id.Name != "scanSource" || len(call.Args) != 2 {
			return true
		}
		path, pathOK := call.Args[0].(*ast.Ident)
		source, sourceOK := call.Args[1].(*ast.Ident)
		if pathOK && path.Name == "relPath" && sourceOK && readNames[source.Name] {
			allowedRefs[id] = true
			calls++
		}
		return true
	})
	if calls != 1 {
		return false
	}

	validRefs := true
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "scanSource" && !allowedRefs[id] {
			validRefs = false
		}
		return validRefs
	})
	return validRefs
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
// local masker or references scanSource outside its authorized path.
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
				t.Errorf("%s: scanFile does not pass raw ReadText output through scanSource's canonical mask", p)
			}
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == "scanSource" {
				t.Errorf("%s: scanSource referenced outside its authorized scan path", p)
			}
			return true
		})
	}
	if !found {
		t.Fatal("no writepath source declares scanFile and scanSource")
	}
}

// TestWritePathMaskFlow_Mutations proves the writepath source pin rejects
// paths that bypass or alter the canonical mask flow.
func TestWritePathMaskFlow_Mutations(t *testing.T) {
	const imp = "import \"" + gomaskImportPath + "\"\n"
	const pimp = "import \"" + pysemImportPath + "\"\n"
	const sourceFn = `func scanSource(relPath, src string) ([]string, error) {
	masked := gomask.MaskGoNonCode(src)
	lines := pysem.SplitLines(masked)
	for range lines {
	}
	return nil, nil
}
`
	const fileFn = `func scanFile(root, relPath string) ([]string, error) {
	text, err := pysem.ReadText(root)
	if err != nil {
		return nil, err
	}
	return scanSource(relPath, text)
}
`
	const head = "package p\n" + imp + pimp
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"production shape", head + sourceFn + fileFn, true},
		{"canonical import under an alias",
			"package p\nimport gm \"" + gomaskImportPath + "\"\n" + pimp + strings.Replace(sourceFn, "gomask.", "gm.", 1) + fileFn, true},
		{"local clone instead of canonical mask", head + strings.Replace(sourceFn, "gomask.MaskGoNonCode(src)", "clone(src)", 1) + fileFn, false},
		{"mask receives a different source", head + strings.Replace(sourceFn, "gomask.MaskGoNonCode(src)", "gomask.MaskGoNonCode(raw())", 1) + fileFn, false},
		{"line scan bypasses the mask", head + strings.Replace(sourceFn, "SplitLines(masked)", "SplitLines(src)", 1) + fileFn, false},
		{"other package masker", head + strings.Replace(sourceFn, "gomask.", "other.", 1) + fileFn, false},
		{"foreign package imported as gomask", "package p\nimport gomask \"example.com/other\"\n" + pimp + sourceFn + fileFn, false},
		{"canonical package not imported", "package p\n" + pimp + sourceFn + fileFn, false},
		{"foreign package imported as pysem", "package p\n" + imp + "import pysem \"example.com/other\"\n" + sourceFn + fileFn, false},
		{"mask moved to scanFile", head + strings.Replace(sourceFn, "SplitLines(masked)", "SplitLines(src)", 1) +
			strings.Replace(fileFn, "return scanSource(relPath, text)", "masked := gomask.MaskGoNonCode(text)\n\treturn scanSource(relPath, masked)", 1), false},
		{"second source call", head + sourceFn + strings.Replace(fileFn, "return scanSource(relPath, text)", "_ = scanSource(relPath, text)\n\treturn scanSource(relPath, text)", 1), false},
		{"source called from another function", head + sourceFn + fileFn + "func raw(s string) { _ = scanSource(\"x\", s) }\n", false},
		{"file never calls source", head + sourceFn + strings.Replace(fileFn, "return scanSource(relPath, text)", "return nil, nil", 1), false},
		{"source shadowed in scanFile", head + sourceFn + strings.Replace(fileFn, "return scanSource(relPath, text)", "scanSource := other\n\treturn scanSource(relPath, text)", 1), false},
		{"source parameter is reassigned", head + strings.Replace(sourceFn, "masked :=", "src = raw()\n\tmasked :=", 1) + fileFn, false},
		{"masked text is reassigned", head + strings.Replace(sourceFn, "lines :=", "masked = raw()\n\tlines :=", 1) + fileFn, false},
		{"masked text address escapes", head + strings.Replace(sourceFn, "lines :=", "clobber(&masked)\n\tlines :=", 1) + fileFn, false},
		{"masked text is shadowed", head + strings.Replace(sourceFn, "lines :=", "masked := raw()\n\tlines :=", 1) + fileFn, false},
		{"closure writes masked text", head + strings.Replace(sourceFn, "lines :=", "f := func() { masked = raw() }\n\tf()\n\tlines :=", 1) + fileFn, false},
		{"loop ranges over different text", head + strings.Replace(sourceFn, "range lines", "range src", 1) + fileFn, false},
		{"second SplitLines path", head + strings.Replace(sourceFn, "for range lines", "_ = pysem.SplitLines(src)\n\tfor range lines", 1) + fileFn, false},
		{"SplitLines bound to a variable", head + strings.Replace(sourceFn, "for range lines", "sl := pysem.SplitLines\n\t_ = sl(masked)\n\tfor range lines", 1) + fileFn, false},
		{"source bound to a package variable", head + sourceFn + fileFn + "var rawScan = scanSource\n", false},
		{"source used as a value in scanFile", head + sourceFn + strings.Replace(fileFn, "return scanSource(relPath, text)", "f := scanSource\n\t_ = f(relPath, text)\n\treturn nil, nil", 1), false},
		{"source passed as an argument", head + sourceFn + fileFn + "func reg() { register(scanSource) }\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), "x.go", tc.src, 0)
			if err != nil {
				t.Fatal(err)
			}
			declared, flows := writePathMaskFlow(f)
			if !declared {
				t.Fatal("scanFile/scanSource not found")
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
