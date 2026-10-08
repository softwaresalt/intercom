// This file (pin.go) ports selection_pathspec_pin (015.011-T AC-6/AG-1,
// re-derived by 032.004-T; Unit A decisions D-030-4 and D-030-6) as a
// SOURCE-TEXT-ANCHORED pin: rather than Python's inspect.getsource(), it
// re-parses tools/gatecheck/internal/retiredarch/select.go from disk with
// go/parser. The package-level scanScope declaration (033.002-T) is the
// single home of the scan-scope literals that selectRepoPaths and
// shouldScanRepoPath both derive from. The pin holds when:
//
//   - Frozen declarations (033.005-T/033.006-T): pathspecFrozenDecls owns
//     PathspecOK and prefixFrozenDecls owns PrefixOK; every declaration in
//     a set must be token-equal to this file's own copy of the plan's
//     §A-CANON texts (canonicalDecls).
//   - Closed world (sharedRulesOK): select.go declares exactly the
//     closedWorldDecls names, each once and with its kind, and the
//     identifiers scanScope and scanArm occur only in confinedIdentDecls.
//   - Scope data (scopeDataOK): the scanScope arm values match
//     pathspecPinLiterals and prefixPinLiterals, with includeTests true
//     exactly on includeTestsPinPrefixes (the round-4 includeTests rule).
//   - Package closure (packageClosureOK, 033.007-T): every non-_test.go .go
//     file in select.go's directory parses; only select.go mentions
//     scanScope or scanArm or declares a closed-world name; no file
//     declares a universe name (universeDeclNames, checked by a
//     toolchain-stable test against types.Universe), imports "unsafe" or
//     "C", carries a //go:linkname directive, or calls os/syscall
//     Setenv, Unsetenv or Clearenv (round-4 environment-mutation rule);
//     and go/build ImportDir reports no non-Go sources (round-4 allowlist).
//
// Every rule fails closed: a shared-rule, scope-data or closure violation,
// or any read, parse or ImportDir error, clears both flags.
//
// Residuals (not enforced here): R-A1, runner wiring and downstream
// dispatch outside the frozen surface (stash DC921AF6); R-A2, git
// environment or configuration set outside this package, narrowed by
// D7BF9F74 (the frozen gitRunnerEnv now isolates DefaultGitRunner's child
// from ambient GIT_* variables and global/system config, and the frozen
// DefaultGitRunner refuses a non-absolute git, but repository-local config
// such as .git/config and a git binary earlier on an absolute PATH entry
// remain trusted); R-A3, _test.go files are outside package closure;
// R-A4, out-of-model attacks (cross-package linkname, reflect, unsafe
// elsewhere, binary patching). A coordinated pin.go + select.go edit
// cannot be caught here: A-T4's independent cmd/x/main.go probe is that
// control, backed by human review of any pin.go diff.
//
// H-11 (self-comparison caveat, gate-reliability plan §8): the expected
// literal lists below (pathspecPinLiterals, prefixPinLiterals) are this
// file's OWN independent expectation, never derived from select.go's
// text. Nothing in select.go is read to produce them. A mutation to
// select.go's literals is therefore genuinely falsifiable against this
// fixed, independently-authored expectation -- not a vacuous
// self-comparison of a value against itself.
package retiredarch

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// pathspecPinLiterals are the exact git-ls-files pathspec arguments
// selectRepoPaths must pass literally in its own call to git.
var pathspecPinLiterals = []string{"config.toml.example", "cmd/**", "internal/**"}

// prefixPinLiterals are the exact prefix literals shouldScanRepoPath must
// compare against literally in its own body.
var prefixPinLiterals = []string{"cmd/", "internal/"}

// canonicalDecls is this file's OWN, independently authored copy of the
// plan's §A-CANON declarations (033.005-T, H-11): it is never derived from
// select.go's text. Frozen declarations of select.go must be token-equal to
// the same-named declaration parsed from this text; comments and whitespace
// are not part of the contract.
const canonicalDecls = `package retiredarch

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

type GitRunner func(root string, pathspecs ...string) ([]byte, error)

func DefaultGitRunner(root string, pathspecs ...string) ([]byte, error) {
	args := append([]string{"ls-files", "--"}, pathspecs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = gitRunnerEnv(os.Environ())
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	if !filepath.IsAbs(cmd.Path) {
		return nil, fmt.Errorf("retiredarch: refusing non-absolute git path %q", cmd.Path)
	}
	var stdout, stderrBuf bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderrBuf.String()); msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

func gitRunnerEnv(environ []string) []string {
	env := make([]string, 0, len(environ)+3)
	for _, kv := range environ {
		name, _, _ := strings.Cut(kv, "=")
		folded := []byte(name)
		for i, c := range folded {
			if 'a' <= c && c <= 'z' {
				folded[i] = c - ('a' - 'A')
			}
		}
		if strings.HasPrefix(string(folded), "GIT_") && string(folded) != "GIT_CEILING_DIRECTORIES" {
			continue
		}
		env = append(env, kv)
	}
	return append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull)
}

type scanArm struct {
	pathspec     string
	prefix       string
	exact        string
	includeTests bool
}

var scanScope = []scanArm{
	{pathspec: "config.toml.example", exact: "config.toml.example"},
	{pathspec: "cmd/**", prefix: "cmd/", includeTests: true},
	{pathspec: "internal/**", prefix: "internal/", includeTests: false},
}

func shouldScanRepoPath(path string) bool {
	for _, a := range scanScope {
		if a.prefix == "" {
			if path == a.exact {
				return true
			}
			continue
		}
		if strings.HasPrefix(path, a.prefix) {
			if !a.includeTests && (strings.Contains(path, "/testdata/") || strings.HasSuffix(path, "_test.go")) {
				return false
			}
			return strings.HasSuffix(path, ".go")
		}
	}
	return false
}

func selectRepoPaths(root string, git GitRunner) ([]string, error) {
	pathspecs := make([]string, 0, len(scanScope))
	for _, a := range scanScope {
		pathspecs = append(pathspecs, a.pathspec)
	}
	out, err := git(root, pathspecs...)
	if err != nil {
		return nil, err
	}
	listing, err := pysem.GitText(out)
	if err != nil {
		return nil, err
	}
	var selected []string
	for _, path := range pysem.SplitLines(listing) {
		if shouldScanRepoPath(path) {
			selected = append(selected, path)
		}
	}
	sort.Strings(selected)
	return selected, nil
}
`

// closedWorldDecls is the closed world of select.go (033.005-T): every
// top-level declaration name it must contain exactly once, with its kind.
// The import declaration is keyed "import". Anything else fails closed.
var closedWorldDecls = map[string]token.Token{
	"import":             token.IMPORT,
	"GitRunner":          token.TYPE,
	"DefaultGitRunner":   token.FUNC,
	"gitRunnerEnv":       token.FUNC,
	"scanArm":            token.TYPE,
	"scanScope":          token.VAR,
	"shouldScanRepoPath": token.FUNC,
	"engineForPath":      token.FUNC,
	"scanPath":           token.FUNC,
	"selectRepoPaths":    token.FUNC,
}

// confinedIdentDecls are the §A-CANON declaration names, the only
// declarations of select.go in which the identifiers scanScope and scanArm
// may occur (fixed from 033.005-T onward). gitRunnerEnv (040-S U5) is
// deliberately absent: it is frozen through canonicalDecls and
// pathspecFrozenDecls but has no reason to mention scanScope or scanArm,
// so it gets no confinement exemption.
var confinedIdentDecls = map[string]bool{
	"import":             true,
	"GitRunner":          true,
	"DefaultGitRunner":   true,
	"scanArm":            true,
	"scanScope":          true,
	"shouldScanRepoPath": true,
	"selectRepoPaths":    true,
}

// pathspecFrozenDecls is the frozen set PathspecOK requires to be
// token-equal to canonicalDecls.
var pathspecFrozenDecls = []string{"import", "GitRunner", "DefaultGitRunner", "gitRunnerEnv", "scanArm", "scanScope", "selectRepoPaths"}

// prefixFrozenDecls is the frozen set PrefixOK requires to be token-equal
// to canonicalDecls (033.006-T).
var prefixFrozenDecls = []string{"import", "scanArm", "scanScope", "shouldScanRepoPath"}

// includeTestsPinPrefixes are the prefix arms whose includeTests field must
// be true; every other arm must leave it false.
var includeTestsPinPrefixes = []string{"cmd/"}

// universeDeclNames is this file's authored list of Go universe-scope
// names (033.007-T): no file in select.go's package may declare one at
// package scope, so the frozen texts cannot be re-bound by shadowing a
// builtin. A toolchain-stable test checks that it covers every universe
// identifier the §A-CANON texts mention.
var universeDeclNames = map[string]bool{
	"any": true, "bool": true, "byte": true, "comparable": true,
	"complex64": true, "complex128": true, "error": true,
	"float32": true, "float64": true, "int": true, "int8": true,
	"int16": true, "int32": true, "int64": true, "rune": true,
	"string": true, "uint": true, "uint8": true, "uint16": true,
	"uint32": true, "uint64": true, "uintptr": true,
	"true": true, "false": true, "iota": true, "nil": true,
	"append": true, "cap": true, "clear": true, "close": true,
	"complex": true, "copy": true, "delete": true, "imag": true,
	"len": true, "make": true, "max": true, "min": true, "new": true,
	"panic": true, "print": true, "println": true, "real": true,
	"recover": true,
}

// envMutators are the os and syscall functions no file in select.go's
// package may reference (round-4 SEC4-3): DefaultGitRunner derives its
// child's environment from the process environment through gitRunnerEnv,
// so an in-package mutation of that environment must stay impossible.
var envMutators = map[string]bool{"Setenv": true, "Unsetenv": true, "Clearenv": true}

// armValues is one scan-scope arm's field values as read from the AST of
// the package-level scope declaration (an absent key is the zero value).
type armValues struct {
	pathspec     string
	prefix       string
	exact        string
	includeTests bool
}

// declTok is one (token, literal) pair of a declaration's token stream.
type declTok struct {
	tok token.Token
	lit string
}

// PathspecPin is the result of the selection pathspec pin check.
type PathspecPin struct {
	SelectFound bool
	GuardFound  bool
	PathspecOK  bool
	PrefixOK    bool
}

// OK reports whether every part of the pin held.
func (p PathspecPin) OK() bool {
	return p.SelectFound && p.GuardFound && p.PathspecOK && p.PrefixOK
}

// checkPathspecPin parses selectGoPath (a filesystem path to a Go source
// file shaped like select.go) and evaluates the pin against it.
// PathspecOK requires selectRepoPaths to exist, the shared rules
// (sharedRulesOK) to hold, the scope field values to satisfy scopeDataOK,
// and the pathspecFrozenDecls set to be token-equal to canonicalDecls.
// PrefixOK requires shouldScanRepoPath to exist, the shared rules and
// scopeDataOK to hold, and the prefixFrozenDecls set to be token-equal to
// canonicalDecls. Package closure (packageClosureOK) is required by both
// flags. A shared-rule, scope-data or closure violation clears both flags.
// A read or parse error of select.go fails closed (every field false); a
// missing function fails its half closed.
func checkPathspecPin(selectGoPath string) PathspecPin {
	src, err := os.ReadFile(selectGoPath)
	if err != nil {
		return PathspecPin{}
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, selectGoPath, src, parser.ParseComments)
	if err != nil {
		return PathspecPin{}
	}

	selectBody := findFuncBody(file, "selectRepoPaths")
	guardBody := findFuncBody(file, "shouldScanRepoPath")

	decls, shared := sharedRulesOK(file)
	scopeOK := scopeDataOK(scopeFieldValues(decls))
	closureOK := packageClosureOK(selectGoPath)

	result := PathspecPin{
		SelectFound: selectBody != nil,
		GuardFound:  guardBody != nil,
	}
	result.PathspecOK = selectBody != nil && shared && scopeOK && closureOK &&
		frozenDeclsOK(fset, src, decls, pathspecFrozenDecls)
	result.PrefixOK = guardBody != nil && shared && scopeOK && closureOK &&
		frozenDeclsOK(fset, src, decls, prefixFrozenDecls)
	return result
}

// packageClosureOK enumerates every non-_test.go .go entry of selectGoPath's
// directory, requires it to be a readable, parseable regular file that
// satisfies closureFileOK, and requires go/build ImportDir (all build
// constraints ignored, cgo enabled) to succeed and report no non-Go
// sources. Any error, a non-regular entry, or a missing select.go fails
// closed.
func packageClosureOK(selectGoPath string) bool {
	dir := filepath.Dir(selectGoPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	selectName := filepath.Base(selectGoPath)
	sawSelect := false
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if !e.Type().IsRegular() {
			return false
		}
		path := filepath.Join(dir, name)
		src, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, src, parser.ParseComments)
		if err != nil {
			return false
		}
		isSelect := name == selectName
		sawSelect = sawSelect || isSelect
		if !closureFileOK(file, isSelect) {
			return false
		}
	}
	if !sawSelect {
		return false
	}
	ctx := build.Default
	ctx.UseAllFiles = true
	ctx.CgoEnabled = true
	pkg, err := ctx.ImportDir(dir, 0)
	if err != nil {
		return false
	}
	return len(pkg.CgoFiles)+len(pkg.CFiles)+len(pkg.CXXFiles)+len(pkg.MFiles)+
		len(pkg.HFiles)+len(pkg.FFiles)+len(pkg.SFiles)+len(pkg.SwigFiles)+
		len(pkg.SwigCXXFiles)+len(pkg.SysoFiles) == 0
}

// closureFileOK applies the per-file package-closure rules to one parsed
// file. isSelect marks select.go itself, the only file that may mention
// scanScope or scanArm or declare a closed-world name. Malformed import
// path literals, imports of unsafe or C, dot-imports of os or syscall, and
// any reference to an envMutators function through an os or syscall import
// fail closed. Import paths are only parsed, never resolved: whether an
// imported package exists is left to the build gate, not this pin.
func closureFileOK(file *ast.File, isSelect bool) bool {
	imports := make(map[string]string)
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil || path == "unsafe" || path == "C" {
			return false
		}
		local := path[strings.LastIndex(path, "/")+1:]
		if imp.Name != nil {
			local = imp.Name.Name
		}
		if local == "." && (path == "os" || path == "syscall") {
			return false
		}
		imports[local] = path
	}
	for _, group := range file.Comments {
		for _, c := range group.List {
			if strings.HasPrefix(c.Text, "//go:linkname") {
				return false
			}
		}
	}
	var names []*ast.Ident
	for _, d := range file.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil {
				names = append(names, d.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					names = append(names, s.Names...)
				case *ast.TypeSpec:
					names = append(names, s.Name)
				}
			}
		}
	}
	for _, n := range names {
		if universeDeclNames[n.Name] {
			return false
		}
		if _, closed := closedWorldDecls[n.Name]; closed && !isSelect {
			return false
		}
	}
	ok := true
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.Ident:
			if !isSelect && (n.Name == "scanScope" || n.Name == "scanArm") {
				ok = false
			}
		case *ast.SelectorExpr:
			if x, isIdent := n.X.(*ast.Ident); isIdent && envMutators[n.Sel.Name] &&
				(imports[x.Name] == "os" || imports[x.Name] == "syscall") {
				ok = false
			}
		}
		return ok
	})
	return ok
}

// scopeFieldValues reads the field values of every element of the
// package-level scope declaration (decls["scanScope"]): a single-spec var
// whose single value is a composite literal of keyed composite-literal
// elements. Keys must be plain identifiers naming a known field, each at
// most once; string fields must be string literals and includeTests the
// identifier true or false. Any other shape returns nil (fail closed).
func scopeFieldValues(decls map[string]ast.Decl) []armValues {
	gen, ok := decls["scanScope"].(*ast.GenDecl)
	if !ok || gen.Tok != token.VAR || len(gen.Specs) != 1 {
		return nil
	}
	spec, ok := gen.Specs[0].(*ast.ValueSpec)
	if !ok || len(spec.Values) != 1 {
		return nil
	}
	list, ok := spec.Values[0].(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var arms []armValues
	for _, elt := range list.Elts {
		lit, ok := elt.(*ast.CompositeLit)
		if !ok || lit.Type != nil {
			return nil
		}
		var arm armValues
		seen := make(map[string]bool)
		for _, field := range lit.Elts {
			kv, ok := field.(*ast.KeyValueExpr)
			if !ok {
				return nil
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || seen[key.Name] {
				return nil
			}
			seen[key.Name] = true
			if key.Name == "includeTests" {
				val, ok := kv.Value.(*ast.Ident)
				if !ok || (val.Name != "true" && val.Name != "false") {
					return nil
				}
				arm.includeTests = val.Name == "true"
				continue
			}
			val, ok := kv.Value.(*ast.BasicLit)
			if !ok || val.Kind != token.STRING {
				return nil
			}
			s, err := strconv.Unquote(val.Value)
			if err != nil {
				return nil
			}
			switch key.Name {
			case "pathspec":
				arm.pathspec = s
			case "prefix":
				arm.prefix = s
			case "exact":
				arm.exact = s
			default:
				return nil
			}
		}
		arms = append(arms, arm)
	}
	return arms
}

// scopeDataOK reports whether arms is exactly the pinned scan scope: the
// pathspec set equals pathspecPinLiterals and the prefix set equals
// prefixPinLiterals (no duplicates in either); no pathspec carries git
// magic (leading ":"); no prefix is a prefix of another and no exact value
// falls under a prefix; a prefix arm's pathspec is prefix+"**" and an exact
// arm's pathspec is its exact value; and includeTests is true for exactly
// the includeTestsPinPrefixes arms. Empty input fails closed.
func scopeDataOK(arms []armValues) bool {
	if len(arms) == 0 {
		return false
	}
	var pathspecs, prefixes, testPrefixes, exacts []string
	seenPathspec, seenPrefix := make(map[string]bool), make(map[string]bool)
	for _, a := range arms {
		if seenPathspec[a.pathspec] || strings.HasPrefix(a.pathspec, ":") {
			return false
		}
		seenPathspec[a.pathspec] = true
		pathspecs = append(pathspecs, a.pathspec)
		if a.prefix == "" {
			if a.pathspec != a.exact || a.includeTests {
				return false
			}
			exacts = append(exacts, a.exact)
			continue
		}
		if seenPrefix[a.prefix] || a.pathspec != a.prefix+"**" {
			return false
		}
		seenPrefix[a.prefix] = true
		prefixes = append(prefixes, a.prefix)
		if a.includeTests {
			testPrefixes = append(testPrefixes, a.prefix)
		}
	}
	for i, p := range prefixes {
		for j, q := range prefixes {
			if i != j && strings.HasPrefix(q, p) {
				return false
			}
		}
		for _, e := range exacts {
			if strings.HasPrefix(e, p) {
				return false
			}
		}
	}
	return containsAll(pathspecs, pathspecPinLiterals) && containsAll(pathspecPinLiterals, pathspecs) &&
		containsAll(prefixes, prefixPinLiterals) && containsAll(prefixPinLiterals, prefixes) &&
		containsAll(testPrefixes, includeTestsPinPrefixes) && containsAll(includeTestsPinPrefixes, testPrefixes)
}

// sharedRulesOK indexes file's top-level declarations by name (the import
// declaration keyed "import") and reports whether the shared rules hold:
//   - closed world: exactly the closedWorldDecls names, each once, with its
//     kind; type/var declarations are single-spec, single-name; functions
//     are receiver-less; anything else (init, _, const, methods, a second
//     import, an extra var/type, a bad declaration) fails;
//   - no //go:build, // +build or //go:linkname directive comment;
//   - the identifiers scanScope and scanArm occur only inside the
//     confinedIdentDecls declarations.
//
// The index is returned even when the rules fail, so a trusted text that is
// not a full closed world (canonicalDecls) can still be indexed.
func sharedRulesOK(file *ast.File) (map[string]ast.Decl, bool) {
	decls := make(map[string]ast.Decl)
	ok := true
	for _, decl := range file.Decls {
		name, kind := "", token.ILLEGAL
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil && d.Name != nil {
				name, kind = d.Name.Name, token.FUNC
			}
		case *ast.GenDecl:
			kind = d.Tok
			switch {
			case d.Tok == token.IMPORT:
				name = "import"
			case len(d.Specs) != 1:
			case d.Tok == token.TYPE:
				if ts, isType := d.Specs[0].(*ast.TypeSpec); isType {
					name = ts.Name.Name
				}
			case d.Tok == token.VAR:
				if vs, isValue := d.Specs[0].(*ast.ValueSpec); isValue && len(vs.Names) == 1 {
					name = vs.Names[0].Name
				}
			}
		}
		if want, known := closedWorldDecls[name]; !known || want != kind || decls[name] != nil {
			ok = false
			continue
		}
		decls[name] = decl
		ast.Inspect(decl, func(n ast.Node) bool {
			if id, isIdent := n.(*ast.Ident); isIdent && (id.Name == "scanScope" || id.Name == "scanArm") && !confinedIdentDecls[name] {
				ok = false
			}
			return true
		})
	}
	if len(decls) != len(closedWorldDecls) {
		ok = false
	}
	for _, group := range file.Comments {
		for _, c := range group.List {
			if strings.HasPrefix(c.Text, "//go:build") || strings.HasPrefix(c.Text, "// +build") || strings.HasPrefix(c.Text, "//go:linkname") {
				ok = false
			}
		}
	}
	return decls, ok
}

// frozenDeclsOK reports whether every declaration named in names exists in
// decls (indexed from src by sharedRulesOK) and is token-equal (declTokens)
// to the same-named declaration of canonicalDecls. A missing declaration on
// either side, an empty names list, or a canonical parse error fails closed.
func frozenDeclsOK(fset *token.FileSet, src []byte, decls map[string]ast.Decl, names []string) bool {
	canonFset := token.NewFileSet()
	canonFile, err := parser.ParseFile(canonFset, "canonical.go", canonicalDecls, parser.ParseComments)
	if err != nil || len(names) == 0 {
		return false
	}
	canonDecls, _ := sharedRulesOK(canonFile)
	for _, name := range names {
		got, want := decls[name], canonDecls[name]
		if got == nil || want == nil {
			return false
		}
		gotToks, gotOK := declTokens(fset, src, got)
		wantToks, wantOK := declTokens(canonFset, []byte(canonicalDecls), want)
		if !gotOK || !wantOK || len(gotToks) != len(wantToks) {
			return false
		}
		for i := range gotToks {
			if gotToks[i] != wantToks[i] {
				return false
			}
		}
	}
	return true
}

// declTokens tokenises the source span of decl with go/scanner, comments
// dropped, and returns its (token, literal) pairs. The span is taken from
// raw byte offsets (token.File.Offset), never from Position, so a //line
// directive cannot move it; decl.Pos() excludes the Doc comment. Every
// SEMICOLON literal is normalised to ";" so an automatically inserted "\n"
// equals an explicit ";". A bad span or a scan error fails closed.
func declTokens(fset *token.FileSet, src []byte, decl ast.Node) ([]declTok, bool) {
	tf := fset.File(decl.Pos())
	if tf == nil || !decl.End().IsValid() {
		return nil, false
	}
	start, end := tf.Offset(decl.Pos()), tf.Offset(decl.End())
	if start < 0 || end > len(src) || start >= end {
		return nil, false
	}
	span := src[start:end]
	var s scanner.Scanner
	scanErrs := 0
	s.Init(token.NewFileSet().AddFile("", -1, len(span)), span, func(token.Position, string) { scanErrs++ }, 0)
	var out []declTok
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON {
			lit = ";"
		}
		out = append(out, declTok{tok: tok, lit: lit})
	}
	return out, scanErrs == 0
}

// findFuncBody locates the *ast.FuncDecl named name at package scope (not
// a method) and returns its body, or nil if not found.
func findFuncBody(file *ast.File, name string) *ast.BlockStmt {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name == nil || fn.Name.Name != name {
			continue
		}
		return fn.Body
	}
	return nil
}

// containsAll reports whether every entry of wanted is present in haystack.
// It is deliberately NOT vacuous (033.004-T): an empty wanted list or an
// empty haystack yields false, so an emptied pin literal list or an empty
// collected literal set fails closed instead of passing. scopeDataOK
// (033.006-T) relies on this contract when it checks set equality by
// calling containsAll in both directions.
func containsAll(haystack []string, wanted []string) bool {
	if len(haystack) == 0 || len(wanted) == 0 {
		return false
	}
	set := make(map[string]bool, len(haystack))
	for _, s := range haystack {
		set[s] = true
	}
	for _, w := range wanted {
		if !set[w] {
			return false
		}
	}
	return true
}

// gitShowToplevel runs `git -C dir rev-parse --show-toplevel` and returns
// the trimmed result, or an error. Like DefaultGitRunner (D7BF9F74), the
// child runs with gitRunnerEnv's isolated environment, so an ambient
// GIT_DIR/GIT_WORK_TREE cannot redirect the pin to another repository; a
// LookPath failure is reported first, and a git that PATH resolved to a
// non-absolute path is refused before it is launched.
func gitShowToplevel(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	cmd.Env = gitRunnerEnv(os.Environ())
	if cmd.Err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", cmd.Err)
	}
	if !filepath.IsAbs(cmd.Path) {
		return "", fmt.Errorf("git rev-parse --show-toplevel: refusing non-absolute git path %q", cmd.Path)
	}
	var stdout, stderrBuf bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderrBuf.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git rev-parse --show-toplevel: %s", msg)
	}
	top := strings.TrimSpace(stdout.String())
	if top == "" {
		return "", fmt.Errorf("git rev-parse --show-toplevel returned nothing")
	}
	return top, nil
}

// SelectionPathspecPin resolves the repository root via
// `git rev-parse --show-toplevel` (anchored at startDir) and evaluates
// the pin against the real, on-disk select.go. A root-resolution
// failure fails closed (every field false), matching Python's
// selection_pathspec_pin() fail-closed-on-retrieval-failure contract.
func SelectionPathspecPin(startDir string) PathspecPin {
	root, err := gitShowToplevel(startDir)
	if err != nil {
		return PathspecPin{}
	}
	selectGoPath := filepath.Join(root, "tools", "gatecheck", "internal", "retiredarch", "select.go")
	return checkPathspecPin(selectGoPath)
}
