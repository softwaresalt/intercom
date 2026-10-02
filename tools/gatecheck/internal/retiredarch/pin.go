// This file (pin.go) ports selection_pathspec_pin (015.011-T AC-6/AG-1,
// re-derived by 032.004-T) as a SOURCE-TEXT-ANCHORED pin: rather than
// Python's inspect.getsource(), this port re-parses
// tools/gatecheck/internal/retiredarch/select.go from disk with go/parser
// and requires the pathspec/prefix literals to appear as *ast.BasicLit
// strings INSIDE the package-level scanScope declaration (033.002-T), the
// single home of the scan-scope literals that selectRepoPaths and
// shouldScanRepoPath both derive from; both functions must still exist.
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
	"scanArm":            token.TYPE,
	"scanScope":          token.VAR,
	"shouldScanRepoPath": token.FUNC,
	"engineForPath":      token.FUNC,
	"scanPath":           token.FUNC,
	"selectRepoPaths":    token.FUNC,
}

// confinedIdentDecls are the §A-CANON declaration names, the only
// declarations of select.go in which the identifiers scanScope and scanArm
// may occur (fixed from 033.005-T onward).
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
var pathspecFrozenDecls = []string{"import", "GitRunner", "DefaultGitRunner", "scanArm", "scanScope", "selectRepoPaths"}

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
// PathspecOK requires selectRepoPaths to exist, scanScope to hold every
// pathspecPinLiterals entry (presence, retained through IVL-1), the
// pathspecFrozenDecls set to be token-equal to canonicalDecls, and the
// shared rules (sharedRulesOK) to hold. PrefixOK requires
// shouldScanRepoPath to exist, scanScope to hold every prefixPinLiterals
// entry, and the shared rules to hold: a shared-rule violation clears both
// flags. A read or parse error fails closed (every field false); a missing
// function or a missing scanScope fails its half closed.
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

	var scopeLits []string
	if scope := findScanScopeDecl(file); scope != nil {
		scopeLits = collectStringLits(scope)
	}
	decls, shared := sharedRulesOK(file)

	result := PathspecPin{
		SelectFound: selectBody != nil,
		GuardFound:  guardBody != nil,
	}
	if selectBody != nil {
		result.PathspecOK = shared &&
			containsAll(scopeLits, pathspecPinLiterals) &&
			frozenDeclsOK(fset, src, decls, pathspecFrozenDecls)
	}
	if guardBody != nil {
		result.PrefixOK = shared && containsAll(scopeLits, prefixPinLiterals)
	}
	return result
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

// findScanScopeDecl is findFuncBody's sibling for the package-level
// scanScope declaration: it returns the top-level var *ast.ValueSpec that
// declares the name scanScope (only that spec, never sibling specs of a
// grouped var block), or nil if there is none.
func findScanScopeDecl(file *ast.File) *ast.ValueSpec {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name == "scanScope" {
					return vs
				}
			}
		}
	}
	return nil
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

// collectStringLits walks node and returns the unquoted string value of
// every *ast.BasicLit of kind token.STRING found anywhere inside it
// (nested expressions, calls, composite literals, any field -- anywhere),
// skipping any literal that fails to strconv.Unquote (never treated as a
// match).
func collectStringLits(node ast.Node) []string {
	var out []string
	ast.Inspect(node, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if v, err := strconv.Unquote(lit.Value); err == nil {
			out = append(out, v)
		}
		return true
	})
	return out
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
// the trimmed result, or an error.
func gitShowToplevel(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
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
