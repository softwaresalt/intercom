// This file (pin.go) ports selection_pathspec_pin (015.011-T AC-6/AG-1,
// re-derived by 032.004-T) as a SOURCE-TEXT-ANCHORED pin: rather than
// Python's inspect.getsource(), which retrieves a running function's own
// source text from its module file, this port re-parses
// tools/gatecheck/internal/retiredarch/select.go from disk with go/parser
// and requires the pathspec/prefix literals to appear as *ast.BasicLit
// strings INSIDE the selectRepoPaths/shouldScanRepoPath function bodies.
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
// file shaped like select.go) and evaluates the pin against its
// selectRepoPaths and shouldScanRepoPath function bodies. A read or parse
// error, or a missing function, fails closed (every field false).
func checkPathspecPin(selectGoPath string) PathspecPin {
	src, err := os.ReadFile(selectGoPath)
	if err != nil {
		return PathspecPin{}
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, selectGoPath, src, 0)
	if err != nil {
		return PathspecPin{}
	}

	selectBody := findFuncBody(file, "selectRepoPaths")
	guardBody := findFuncBody(file, "shouldScanRepoPath")

	result := PathspecPin{
		SelectFound: selectBody != nil,
		GuardFound:  guardBody != nil,
	}
	if selectBody != nil {
		lits := collectStringLits(selectBody)
		result.PathspecOK = containsAll(lits, pathspecPinLiterals)
	}
	if guardBody != nil {
		lits := collectStringLits(guardBody)
		result.PrefixOK = containsAll(lits, prefixPinLiterals)
	}
	return result
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

// collectStringLits walks body and returns the unquoted string value of
// every *ast.BasicLit of kind token.STRING found anywhere inside it
// (nested expressions, calls, composite literals -- anywhere), skipping
// any literal that fails to strconv.Unquote (never treated as a match).
func collectStringLits(body *ast.BlockStmt) []string {
	var out []string
	ast.Inspect(body, func(n ast.Node) bool {
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

func containsAll(haystack []string, wanted []string) bool {
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
