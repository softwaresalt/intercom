// Package writepath reimplements the repository's write-path precondition
// gate (scripts/check-write-path-precondition.sh) on top of
// tools/gatecheck/internal/{pysem,gomask}, so the invariant -- no
// destructive filesystem write primitive exists in any non-test Go file
// under internal/** or cmd/** -- can be enforced without a Python
// interpreter.
//
// Run mirrors the bash wrapper's own top-level dispatch (bare invocation,
// --self-test, --self-test-integrity, or an unknown flag) rather than only
// the inner Python script's mode argument, so a future CLI wrapper (M1-T10)
// only needs to forward argv and the resolved --root.
//
// Detector scope is the ordered 76-selector list in Selectors below; it
// matches the wrapper's detector-scope list in
// scripts/check-write-path-precondition.sh.
//
// # Residual evasion surface (034.007-T, AC-4.7)
//
// The detector includes the 76 qualified selectors in Selectors and the
// narrowly-scoped pathsafe.Root.Resolve receiver analysis in item 2. The
// following surface is recorded here, never silently left unhandled:
//
// Closed residuals: item 1 (named import aliases) is closed by 049.004-T,
// using reject-alias-os-writefile.go. Item 7 (selector split by a newline or
// comment) is closed by 049.002-T and pinned by 049.004-T's
// reject-split-selector.go.
//
//  2. Root.Resolve is reported when its receiver is bound by pathsafe.NewRoot
//     within the same FuncDecl or FuncLit, including the supported two-name
//     declaration and assignment forms. This same-function binding is closed
//     by 049.005-T / reject-resolve-first-caller.go. Cross-function flows,
//     struct fields, package-level variables, and method values remain
//     residuals.
//  3. Dot imports of write-capable packages are reported at the import spec,
//     closed by 049.004-T / reject-dot-import-os.go. Blank imports are inert.
//     Local identifiers shadowing a package name (including a local syscall
//     identifier) remain residual.
//  4. os.Root method calls and (*os.File).Write*.
//  5. Undecidable or non-simple call shape -> rejected. A
//     syscall.CreateFile reached via a wrapper or function value, with an
//     extent that does not balance, with any argument that is not a bare
//     operand (call, composite literal, index, string or raw string), or
//     with any access, disposition or flags spelling outside the D-2'
//     predicate's exact tokens is rejected, not exempted (D-2' rules 2-7).
//     This is a false-positive surface: a future legitimate metadata-only
//     call in such a shape trips the gate and needs an explicit, reviewed
//     widening. It is never a silent hole.
//  6. Write primitives outside the selector set are not detected. E-T7 closes
//     the 50 enumerated ioutil, syscall, x/sys/windows and x/sys/unix
//     primitives in 049.007-T and fixtures reject-ioutil-write-primitives.go,
//     reject-syscall-namespace-primitives.go and
//     reject-xsys-write-primitives.go. The same-family remainder remains:
//     metadata and attribute writes; syscall.Pwrite, Ftruncate, Link and
//     Symlink; unix.Writev, Pwritev, Mknod and Mknodat; and other
//     non-enumerated primitives. This remainder is tracked as stash entry
//     C0D28448. In particular, the uncovered symbols include
//     syscall.Chmod/Fchmod/Chown/Utimes/SetFileAttributes,
//     unix.Fchmod/Chown/Fchown/Lchown/Utimes/Setxattr,
//     windows.SetFileAttributes, unix.Writev/Pwritev, syscall.Ftruncate,
//     syscall.Link/Symlink, unix.Mknod/Mknodat, and any other un-enumerated
//     write-capable symbol in those packages.
//  8. Dynamic invocation through syscall.NewLazyDLL / LazyProc.Call,
//     syscall.Syscall* and the golang.org/x/sys equivalents can reach any OS
//     write API without a write-named selector. syscall.NewLazyDLL is live in
//     production, so this surface cannot become a finding without an
//     allowance design. It is outside Unit E and deferred as stash entry
//     FE2F02FF.
//
// Residual-risk statement: this gate remains a narrow AST tripwire, not a
// complete mechanical proof. Items 2, 4, 5, 6 and 8, plus local package-name
// shadowing in item 3, remain residual surfaces. The compensating control is
// human and agent PR review against this list.
package writepath

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"
	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// Selectors is the ordered list of 76 qualified write-primitive selectors
// this gate detects. The first 20 were ported verbatim from the retired
// Python SELECTORS list (including its three adversarial-review
// additions); the next six (syscall.CreateFile, syscall.Write,
// os.OpenRoot, os.Root, io.CopyN, io.CopyBuffer) were appended by 034.004-T.
// The final 50 import-path-keyed selectors were appended by 049.007-T:
// three ioutil, 14 syscall, 11 x/sys/windows and 22 x/sys/unix selectors.
// syscall.CreateFile is reported unless callAllowed proves the exact
// metadata-only reparse-probe call shape.
var Selectors = []string{
	"os.WriteFile", "os.Create", "os.OpenFile", "os.Remove", "os.RemoveAll",
	"os.Rename", "os.Mkdir", "os.MkdirAll", "os.Symlink", "os.Chmod",
	"os.Truncate", "io.Copy", "sql.Open", "bbolt.Open",
	"os.CreateTemp", "os.MkdirTemp", "os.Link", "os.Chown", "os.Lchown",
	"os.Chtimes",
	"syscall.CreateFile", "syscall.Write", "os.OpenRoot", "os.Root",
	"io.CopyN", "io.CopyBuffer",
	"ioutil.WriteFile", "ioutil.TempFile", "ioutil.TempDir",
	"syscall.WriteFile", "syscall.Open", "syscall.Unlink", "syscall.Rename",
	"syscall.Mkdir", "syscall.Rmdir", "syscall.CreateHardLink", "syscall.DeleteFile",
	"syscall.MoveFile", "syscall.RemoveDirectory", "syscall.CreateDirectory",
	"syscall.CreateSymbolicLink", "syscall.Truncate", "syscall.Creat",
	"windows.WriteFile", "windows.CreateFile", "windows.DeleteFile",
	"windows.MoveFile", "windows.MoveFileEx", "windows.CreateDirectory",
	"windows.RemoveDirectory", "windows.CreateHardLink", "windows.CreateSymbolicLink",
	"windows.SetEndOfFile", "windows.SetFileInformationByHandle",
	"unix.Open", "unix.Openat", "unix.Openat2", "unix.Creat", "unix.Write",
	"unix.Pwrite", "unix.Unlink", "unix.Unlinkat", "unix.Rename", "unix.Renameat",
	"unix.Renameat2", "unix.Mkdir", "unix.Mkdirat", "unix.Rmdir", "unix.Link",
	"unix.Linkat", "unix.Symlink", "unix.Symlinkat", "unix.Truncate",
	"unix.Ftruncate", "unix.Chmod", "unix.Fchmodat",
}

const (
	exceptionName      = "011.003-T's Constitution Check exception (internal/config/validate.go rule 7)"
	registerName       = "the consolidated risk register (internal/pathsafe package doc, root.go)"
	pathsafeImportPath = "github.com/softwaresalt/intercom-go/internal/pathsafe"
	rootResolveDisplay = "pathsafe.Root.Resolve"
)

// GitRunner runs `git ls-files -- internal/** cmd/**` rooted at root and
// returns its raw stdout bytes (for pysem.GitText decoding), or an error if
// the process could not be started or exited non-zero. It is injectable so
// tests can simulate a missing/failing git without depending on the real
// repository tree (M1-T8 AC: "an injectable gitRunner").
type GitRunner func(root string) ([]byte, error)

// DefaultGitRunner is the production GitRunner: it shells out to
// `git ls-files -- internal/** cmd/**` with root as the working directory.
func DefaultGitRunner(root string) ([]byte, error) {
	cmd := exec.Command("git", "ls-files", "--", "internal/**", "cmd/**")
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

// shouldScan reports whether relPath (a forward-slash, repo-root-relative
// path as produced by `git ls-files`) is a non-test Go source file under
// internal/** or cmd/**. Ported verbatim from should_scan().
func shouldScan(relPath string) bool {
	if strings.HasSuffix(relPath, "_test.go") {
		return false
	}
	if !strings.HasSuffix(relPath, ".go") {
		return false
	}
	return strings.HasPrefix(relPath, "internal/") || strings.HasPrefix(relPath, "cmd/")
}

// nextOccurrence returns the byte offset of the first occurrence of sel in
// line at or after start that is a qualified selector -- not preceded by a
// word rune or '.', and not followed by a word rune -- or -1 if there is
// none. It is the explicit replacement for one step of Python's
// re.compile(r'(?<![\w.])' + re.escape(sel) + r'(?![\w])').finditer(line).
func nextOccurrence(line, sel string, start int) int {
	for start <= len(line) {
		idx := strings.Index(line[start:], sel)
		if idx < 0 {
			return -1
		}
		pos := start + idx
		if !pysem.PrecededByWordOrDot(line, pos) && !pysem.FollowedByWord(line, pos+len(sel)) {
			return pos
		}
		start = pos + 1
	}
	return -1
}

// selectorOccurrences returns the byte offset of every qualified-selector
// occurrence of sel in line, in order, or nil if there is none.
func selectorOccurrences(line, sel string) []int {
	var offsets []int
	for pos := nextOccurrence(line, sel, 0); pos >= 0; pos = nextOccurrence(line, sel, pos+1) {
		offsets = append(offsets, pos)
	}
	return offsets
}

// The access-mode allowance (D-2') admits exactly one argument shape: the
// metadata-only syscall.CreateFile call that opens an existing object with
// no access rights and only FILE_FLAG_BACKUP_SEMANTICS.
const (
	allowedCreateFileName     = "CreateFile"
	allowedCreateFileArgCount = 7
	allowedCreateFileAccess   = "0"
	allowedCreateFileDispo    = "OPEN_EXISTING"
	allowedCreateFileFlags    = "FILE_FLAG_BACKUP_SEMANTICS"
)

// importPathBindings returns each file's declared import identifier mapped
// to its exact import path. Unnamed imports use the final path component,
// while declared aliases retain the identifier written in the import.
func importPathBindings(file *ast.File) (map[string]string, error) {
	bindings := make(map[string]string, len(file.Imports))
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("decode import path %s: %w", spec.Path.Value, err)
		}
		localName := path.Base(importPath)
		if spec.Name != nil {
			localName = spec.Name.Name
		}
		bindings[localName] = importPath
	}
	return bindings, nil
}

func isPathsafeNewRootCall(expr ast.Expr, bindings map[string]string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "NewRoot" {
		return false
	}
	qualifier, ok := selector.X.(*ast.Ident)
	return ok && bindings[qualifier.Name] == pathsafeImportPath
}

func bindNewRootReceiver(root, result *ast.Ident, value ast.Expr, bindings map[string]string, roots map[string]struct{}) {
	if root == nil || root.Name == "_" || result == nil || !isPathsafeNewRootCall(value, bindings) {
		return
	}
	roots[root.Name] = struct{}{}
}

func pathsafeRootResolveSelectors(file *ast.File, bindings map[string]string) map[*ast.SelectorExpr]struct{} {
	selectors := make(map[*ast.SelectorExpr]struct{})
	importsPathsafe := false
	for _, importPath := range bindings {
		if importPath == pathsafeImportPath {
			importsPathsafe = true
			break
		}
	}
	if !importsPathsafe {
		return selectors
	}

	inspectFunction := func(body *ast.BlockStmt) {
		if body == nil {
			return
		}
		roots := make(map[string]struct{})
		ast.Inspect(body, func(node ast.Node) bool {
			if node == nil {
				return true
			}
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			switch n := node.(type) {
			case *ast.AssignStmt:
				if (n.Tok != token.DEFINE && n.Tok != token.ASSIGN) || len(n.Lhs) != 2 || len(n.Rhs) != 1 {
					break
				}
				root, rootOK := n.Lhs[0].(*ast.Ident)
				result, resultOK := n.Lhs[1].(*ast.Ident)
				if rootOK && resultOK {
					bindNewRootReceiver(root, result, n.Rhs[0], bindings, roots)
				}
			case *ast.ValueSpec:
				if len(n.Names) == 2 && len(n.Values) == 1 {
					bindNewRootReceiver(n.Names[0], n.Names[1], n.Values[0], bindings, roots)
				}
			}
			return true
		})

		ast.Inspect(body, func(node ast.Node) bool {
			if node == nil {
				return true
			}
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			switch n := node.(type) {
			case *ast.CallExpr:
				selector, ok := n.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Resolve" {
					break
				}
				receiver := selector.X
				for {
					parenthesized, ok := receiver.(*ast.ParenExpr)
					if !ok {
						break
					}
					receiver = parenthesized.X
				}
				receiverIdent, ok := receiver.(*ast.Ident)
				if !ok {
					break
				}
				if _, bound := roots[receiverIdent.Name]; bound {
					selectors[selector] = struct{}{}
				}
			}
			return true
		})
	}

	for _, decl := range file.Decls {
		if function, ok := decl.(*ast.FuncDecl); ok {
			inspectFunction(function.Body)
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if function, ok := node.(*ast.FuncLit); ok {
			inspectFunction(function.Body)
		}
		return true
	})
	return selectors
}

func canonicalImportPath(qualifier string) (string, bool) {
	switch qualifier {
	case "os", "io", "syscall":
		return qualifier, true
	case "ioutil":
		return "io/ioutil", true
	case "sql":
		return "database/sql", true
	case "bbolt":
		return "go.etcd.io/bbolt", true
	case "windows":
		return "golang.org/x/sys/windows", true
	case "unix":
		return "golang.org/x/sys/unix", true
	default:
		return "", false
	}
}

func selectorIndexForImportPath(importPath, name string) (int, bool) {
	for i, selector := range Selectors {
		qualifier, selectorName, ok := strings.Cut(selector, ".")
		if !ok || selectorName != name {
			continue
		}
		canonicalPath, ok := canonicalImportPath(qualifier)
		if ok && canonicalPath == importPath {
			return i, true
		}
	}
	return 0, false
}

func selectorPrefixIndexForImportPath(importPath string) (string, int, bool) {
	for i, selector := range Selectors {
		qualifier, _, ok := strings.Cut(selector, ".")
		if !ok {
			continue
		}
		canonicalPath, ok := canonicalImportPath(qualifier)
		if ok && canonicalPath == importPath {
			return qualifier, i, true
		}
	}
	return "", 0, false
}

func isSyscallSelector(expr ast.Expr, bindings map[string]string, name string) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != name {
		return false
	}
	qualifier, ok := selector.X.(*ast.Ident)
	return ok && selector.Sel.Pos()-qualifier.End() == 1 && bindings[qualifier.Name] == "syscall"
}

// bareOperand reports whether expr is one of the G-4 operand forms. Calls,
// literals containing strings or characters, and composite/index expressions
// are deliberately rejected.
func bareOperand(expr ast.Expr) bool {
	switch node := expr.(type) {
	case *ast.Ident:
		return true
	case *ast.BasicLit:
		return node.Kind == token.INT || node.Kind == token.FLOAT || node.Kind == token.IMAG
	case *ast.SelectorExpr:
		return bareOperand(node.X)
	case *ast.UnaryExpr:
		return bareOperand(node.X)
	case *ast.StarExpr:
		return bareOperand(node.X)
	case *ast.BinaryExpr:
		return bareOperand(node.X) && bareOperand(node.Y)
	default:
		return false
	}
}

// callAllowed is the D-2' AST allowance for the metadata-only
// syscall.CreateFile call. It accepts only the exact import-path binding,
// disposition, flags, zero access literal, argument count and G-4 operands.
func callAllowed(call *ast.CallExpr, bindings map[string]string) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != allowedCreateFileName {
		return false
	}
	qualifier, ok := selector.X.(*ast.Ident)
	if !ok || bindings[qualifier.Name] != "syscall" {
		return false
	}
	if call.Ellipsis.IsValid() || len(call.Args) != allowedCreateFileArgCount {
		return false
	}
	access, ok := call.Args[1].(*ast.BasicLit)
	if !ok || access.Kind != token.INT || access.Value != allowedCreateFileAccess {
		return false
	}
	if !isSyscallSelector(call.Args[4], bindings, allowedCreateFileDispo) ||
		!isSyscallSelector(call.Args[5], bindings, allowedCreateFileFlags) {
		return false
	}
	for i, arg := range call.Args {
		if i == 1 || i == 4 || i == 5 {
			continue
		}
		if !bareOperand(arg) {
			return false
		}
	}
	return true
}

// scanFile reads the file at filepath.Join(root, filepath.FromSlash(relPath))
// through pysem.ReadText (rejecting invalid UTF-8 and translating
// newlines), masks it via gomask.MaskGoNonCode, and scans the result.
// relPath is echoed back into finding text verbatim.
func scanFile(root, relPath string) ([]string, error) {
	text, err := pysem.ReadText(filepath.Join(root, filepath.FromSlash(relPath)))
	if err != nil {
		return nil, err
	}
	return scanSource(relPath, text)
}

func runeByteOffsets(text string) []int {
	offsets := make([]int, 0, utf8.RuneCountInString(text)+1)
	for offset := range text {
		offsets = append(offsets, offset)
	}
	return append(offsets, len(text))
}

func runeIndexForByteOffset(sourceByteOffset int, sourceRuneOffsets []int) (int, bool) {
	runeIndex := sort.SearchInts(sourceRuneOffsets, sourceByteOffset)
	if runeIndex >= len(sourceRuneOffsets) || sourceRuneOffsets[runeIndex] != sourceByteOffset {
		return 0, false
	}
	return runeIndex, true
}

func maskedOffsetForSourceByte(sourceByteOffset int, sourceRuneOffsets, maskedRuneOffsets []int) (int, bool) {
	runeIndex, ok := runeIndexForByteOffset(sourceByteOffset, sourceRuneOffsets)
	if !ok || runeIndex >= len(maskedRuneOffsets) {
		return 0, false
	}
	return maskedRuneOffsets[runeIndex], true
}

func physicalSourceOffset(file *token.File, pos token.Pos, sourceLen int) (int, bool) {
	if file == nil || file.Size() != sourceLen || !pos.IsValid() || pos < file.Pos(0) || pos > file.Pos(file.Size()) {
		return 0, false
	}
	position := file.PositionFor(pos, false)
	if !position.IsValid() || position.Offset < 0 || position.Offset > sourceLen {
		return 0, false
	}
	offset := file.Offset(pos)
	if offset != position.Offset {
		return 0, false
	}
	return offset, true
}

func lineNumberAtOffset(lineStarts []int, offset int) (int, bool) {
	line := sort.Search(len(lineStarts), func(i int) bool {
		return lineStarts[i] > offset
	}) - 1
	if line < 0 {
		return 0, false
	}
	return line + 1, true
}

// scanSource scans unmasked, decoded Go source and fails closed on any parse
// error. The whole-file mask supplies the legacy line boundaries and exposes
// only struct-tag raw strings for the residual textual selector check.
func scanSource(relPath, src string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, relPath, src, parser.ParseComments|parser.AllErrors|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse Go source %q: %w", relPath, err)
	}

	importBindings, err := importPathBindings(file)
	if err != nil {
		return nil, fmt.Errorf("read imports in Go source %q: %w", relPath, err)
	}
	rootResolveSelectors := pathsafeRootResolveSelectors(file, importBindings)
	masked := gomask.MaskGoNonCode(src)
	sourceRuneOffsets := runeByteOffsets(src)
	maskedRuneOffsets := runeByteOffsets(masked)
	if len(sourceRuneOffsets) != len(maskedRuneOffsets) {
		return nil, fmt.Errorf("mask Go source %q: rune count changed", relPath)
	}

	lines := pysem.SplitLines(masked)
	lineStarts := make([]int, len(lines))
	cursor := 0
	for i, line := range lines {
		lineStarts[i] = cursor
		lineEnd := cursor + len(line)
		if lineEnd > len(masked) || masked[cursor:lineEnd] != line {
			return nil, fmt.Errorf("map masked lines in %q: lost synchronization at byte %d", relPath, cursor)
		}
		cursor = lineEnd
		if cursor == len(masked) {
			continue
		}
		if strings.HasPrefix(masked[cursor:], "\r\n") {
			cursor += 2
			continue
		}
		r, width := utf8.DecodeRuneInString(masked[cursor:])
		if r == utf8.RuneError && width <= 1 {
			return nil, fmt.Errorf("map masked lines in %q: lost synchronization at byte %d", relPath, cursor)
		}
		cursor += width
	}
	if cursor != len(masked) {
		return nil, fmt.Errorf("map masked lines in %q: stopped at byte %d of %d", relPath, cursor, len(masked))
	}

	selectorIndexes := make(map[string]int, len(Selectors))
	for i, selector := range Selectors {
		selectorIndexes[selector] = i
	}

	allowedCalls := make(map[*ast.SelectorExpr]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !callAllowed(call, importBindings) {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if ok {
			allowedCalls[selector] = struct{}{}
		}
		return true
	})

	type findingKey struct {
		line          int
		selectorIndex int
		display       string
	}
	findings := make(map[findingKey]struct{})
	addFinding := func(maskedOffset, selectorIndex int, display string) error {
		line, ok := lineNumberAtOffset(lineStarts, maskedOffset)
		if !ok {
			return fmt.Errorf("map finding in %q: invalid masked byte offset %d", relPath, maskedOffset)
		}
		findings[findingKey{line: line, selectorIndex: selectorIndex, display: display}] = struct{}{}
		return nil
	}

	var scanErr error
	ast.Inspect(file, func(node ast.Node) bool {
		if scanErr != nil {
			return false
		}
		switch n := node.(type) {
		case *ast.SelectorExpr:
			if _, rootResolve := rootResolveSelectors[n]; rootResolve {
				tf := fset.File(n.Pos())
				sourceOffset, ok := physicalSourceOffset(tf, n.Pos(), len(src))
				if !ok {
					scanErr = fmt.Errorf("map Root.Resolve position in %q: invalid token position", relPath)
					return false
				}
				maskedOffset, ok := maskedOffsetForSourceByte(sourceOffset, sourceRuneOffsets, maskedRuneOffsets)
				if !ok {
					scanErr = fmt.Errorf("map Root.Resolve position in %q: invalid source byte offset %d", relPath, sourceOffset)
					return false
				}
				scanErr = addFinding(maskedOffset, len(Selectors), rootResolveDisplay)
				return true
			}
			qualifier, ok := n.X.(*ast.Ident)
			if !ok {
				return true
			}
			selectorIndex, ok := selectorIndexes[qualifier.Name+"."+n.Sel.Name]
			importPath, bound := importBindings[qualifier.Name]
			if bound {
				if canonicalIndex, matched := selectorIndexForImportPath(importPath, n.Sel.Name); matched {
					if !ok {
						selectorIndex = canonicalIndex
						ok = true
					}
				}
			}
			if !ok {
				return true
			}
			if _, allowed := allowedCalls[n]; allowed {
				return true
			}
			tf := fset.File(n.Pos())
			sourceOffset, ok := physicalSourceOffset(tf, n.Pos(), len(src))
			if !ok {
				scanErr = fmt.Errorf("map selector position in %q: invalid token position", relPath)
				return false
			}
			maskedOffset, ok := maskedOffsetForSourceByte(sourceOffset, sourceRuneOffsets, maskedRuneOffsets)
			if !ok {
				scanErr = fmt.Errorf("map selector position in %q: invalid source byte offset %d", relPath, sourceOffset)
				return false
			}
			scanErr = addFinding(maskedOffset, selectorIndex, "")

		case *ast.BasicLit:
			if n.Kind != token.STRING || len(n.Value) < 2 || n.Value[0] != '`' || n.Value[len(n.Value)-1] != '`' {
				return true
			}
			tf := fset.File(n.Pos())
			sourceStart, ok := physicalSourceOffset(tf, n.Pos(), len(src))
			if !ok {
				scanErr = fmt.Errorf("map raw string in %q: invalid start position", relPath)
				return false
			}
			sourceEnd, ok := physicalSourceOffset(tf, n.End(), len(src))
			if !ok || sourceEnd-sourceStart < 2 || src[sourceStart] != '`' || src[sourceEnd-1] != '`' {
				scanErr = fmt.Errorf("map raw string in %q: invalid end position", relPath)
				return false
			}
			startRune, ok := runeIndexForByteOffset(sourceStart+1, sourceRuneOffsets)
			if !ok {
				scanErr = fmt.Errorf("map raw string in %q: invalid interior start", relPath)
				return false
			}
			endRune, ok := runeIndexForByteOffset(sourceEnd-1, sourceRuneOffsets)
			if !ok || endRune < startRune {
				scanErr = fmt.Errorf("map raw string in %q: invalid interior end", relPath)
				return false
			}
			maskedStart := maskedRuneOffsets[startRune]
			maskedEnd := maskedRuneOffsets[endRune]
			sourceInterior := src[sourceStart+1 : sourceEnd-1]
			maskedInterior := masked[maskedStart:maskedEnd]
			if sourceInterior != maskedInterior {
				return true
			}
			for selectorIndex, selector := range Selectors {
				for _, occurrence := range selectorOccurrences(maskedInterior, selector) {
					if err := addFinding(maskedStart+occurrence, selectorIndex, ""); err != nil {
						scanErr = err
						return false
					}
				}
			}
		}
		return true
	})
	if scanErr != nil {
		return nil, scanErr
	}

	for _, spec := range file.Imports {
		if spec.Name == nil || spec.Name.Name != "." {
			continue
		}
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("decode import path %s: %w", spec.Path.Value, err)
		}
		prefix, selectorIndex, ok := selectorPrefixIndexForImportPath(importPath)
		if !ok {
			continue
		}
		tf := fset.File(spec.Pos())
		sourceOffset, ok := physicalSourceOffset(tf, spec.Pos(), len(src))
		if !ok {
			return nil, fmt.Errorf("map dot import position in %q: invalid token position", relPath)
		}
		maskedOffset, ok := maskedOffsetForSourceByte(sourceOffset, sourceRuneOffsets, maskedRuneOffsets)
		if !ok {
			return nil, fmt.Errorf("map dot import position in %q: invalid source byte offset %d", relPath, sourceOffset)
		}
		if err := addFinding(maskedOffset, selectorIndex, prefix+".*"); err != nil {
			return nil, err
		}
	}

	ordered := make([]findingKey, 0, len(findings))
	for finding := range findings {
		ordered = append(ordered, finding)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].line != ordered[j].line {
			return ordered[i].line < ordered[j].line
		}
		if ordered[i].selectorIndex != ordered[j].selectorIndex {
			return ordered[i].selectorIndex < ordered[j].selectorIndex
		}
		return ordered[i].display < ordered[j].display
	})

	result := make([]string, 0, len(ordered))
	for _, finding := range ordered {
		selector := finding.display
		if selector == "" {
			selector = Selectors[finding.selectorIndex]
		}
		result = append(result, fmt.Sprintf("%s:%d: write primitive %s found", relPath, finding.line, pysem.Repr(selector)))
	}
	return result, nil
}

// Result carries the ordered stdout/stderr text and process-style exit
// code a single mode produced, mirroring writepath_golden.json's
// stream_captures shape.
type Result struct {
	Stdout string
	Stderr string
	Code   int
}

// errorLine formats a one-line, "::error::"-prefixed ED-2 message: the
// replacement for what would otherwise be an uncaught Python exception
// (a git failure, a file read failure, or invalid UTF-8) propagating as a
// traceback. context names the file or operation that failed.
func errorLine(context string, err error) string {
	return fmt.Sprintf("::error::%s: %v\n", context, err)
}

// containedRegularFile reports whether rel, a repo-relative '/'-separated path
// exactly as `git ls-files` prints it, names a regular file reached without
// following a link (D44D8BDF, writepath arm). It is a semantic copy of
// retiredarch's containedRegularFile (retiredarch U3, 990AFA71). It is copied,
// not shared, because each gate package keeps its own surface (D-BA-3,
// 5A8EC1BC). The walk uses os.Lstat one component at a time. Every
// intermediate component must be a real directory (no ModeSymlink and no
// ModeIrregular, so a Windows junction is rejected), and the final component
// must be regular. An empty, "." or ".." component, or on Windows a component
// containing '\' or ':', is a containment violation. Only fs.ErrNotExist
// returns ok=true, so a deleted tracked file keeps the existing read-error
// text. Any other Lstat error fails closed. The *fs.PathError wrapper is
// stripped so CI logs show only repo-relative prefixes (SEC-4). Stdlib only:
// no filepath.Abs and no filepath.EvalSymlinks.
func containedRegularFile(root, rel string) (bool, string) {
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false, fmt.Sprintf("invalid path component %q in %q", part, rel)
		}
		if runtime.GOOS == "windows" && strings.ContainsAny(part, `\:`) {
			return false, fmt.Sprintf("invalid path component %q in %q", part, rel)
		}
	}
	current := root
	for i, part := range parts {
		current = filepath.Join(current, part)
		prefix := strings.Join(parts[:i+1], "/")
		info, err := os.Lstat(current)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return true, ""
			}
			var pathErr *fs.PathError
			if errors.As(err, &pathErr) {
				err = pathErr.Err
			}
			return false, fmt.Sprintf("lstat %s: %v", prefix, err)
		}
		mode := info.Mode()
		if i < len(parts)-1 {
			if !mode.IsDir() || mode&(fs.ModeSymlink|fs.ModeIrregular) != 0 {
				return false, fmt.Sprintf("%s is not a real directory (mode %v)", prefix, mode)
			}
			continue
		}
		if !mode.IsRegular() {
			return false, fmt.Sprintf("%s is not a regular file (mode %v)", prefix, mode)
		}
	}
	return true, ""
}

// runRepoScan performs the repo-mode invariant scan: enumerate every
// tracked internal/**, cmd/** Go source (via git, injectable), scan each,
// and fail closed on a git error, a per-file read/decode error (ED-2), a
// selected path that is not a contained regular file (D44D8BDF: a symlink or
// junction is never followed), any finding, or an empty selection (ED-7).
func runRepoScan(root string, git GitRunner) Result {
	out, err := git(root)
	if err != nil {
		return Result{Stderr: errorLine("git ls-files", err), Code: 1}
	}
	listing, err := pysem.GitText(out)
	if err != nil {
		return Result{Stderr: errorLine("git ls-files output", err), Code: 1}
	}

	var relPaths []string
	for _, p := range pysem.SplitLines(listing) {
		if p == "" {
			continue
		}
		if shouldScan(p) {
			relPaths = append(relPaths, p)
		}
	}

	if len(relPaths) == 0 {
		return Result{
			Stderr: "::error::write-path repo scan matched no files under internal/** or cmd/** (ED-7)\n",
			Code:   1,
		}
	}

	var findings []string
	for _, rel := range relPaths {
		if ok, reason := containedRegularFile(root, rel); !ok {
			return Result{Stderr: errorLine(rel, fmt.Errorf("not a contained regular file: %s", reason)), Code: 1}
		}
		fileFindings, err := scanFile(root, rel)
		if err != nil {
			return Result{Stderr: errorLine(rel, err), Code: 1}
		}
		findings = append(findings, fileFindings...)
	}

	if len(findings) > 0 {
		var b strings.Builder
		for _, f := range findings {
			b.WriteString(f)
			b.WriteByte('\n')
		}
		_, _ = fmt.Fprintf(&b,
			"::error::a destructive filesystem write primitive was found under internal/** or cmd/**. "+
				"This trips the write-path precondition recorded in %s and tracked in %s. "+
				"RETIREMENT PROCEDURE: re-evaluate every finding in the risk register against this new "+
				"write call site, land a mitigation (or an explicit, re-justified acceptance) in the SAME "+
				"change, and update both the register and the Constitution Check exception to reflect the "+
				"arrival of a real write path.\n",
			exceptionName, registerName,
		)
		return Result{Stderr: b.String(), Code: 1}
	}

	return Result{Code: 0}
}

// runFixtureSelfTest performs ONLY the fixture self-test: every
// scripts/testdata/writepath/*.go fixture is scanned and checked against
// its accept-/reject- filename prefix. It is shared, verbatim, by both the
// "self-test" and "self-test-integrity" top-level modes; whether the repo
// scan also runs afterward is Run's concern, not this function's.
func runFixtureSelfTest(root string) Result {
	fixtureDir := filepath.Join(root, "scripts", "testdata", "writepath")
	fixtureDirPosix := filepath.ToSlash(fixtureDir)

	info, statErr := os.Stat(fixtureDir)
	if statErr != nil || !info.IsDir() {
		return Result{Stderr: fmt.Sprintf("fixture dir not found: %s\n", fixtureDirPosix), Code: 1}
	}

	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		return Result{Stderr: fmt.Sprintf("fixture dir not found: %s\n", fixtureDirPosix), Code: 1}
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".go") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	if len(names) == 0 {
		return Result{Stderr: fmt.Sprintf("no fixtures discovered under %s\n", fixtureDirPosix), Code: 1}
	}

	var stdout strings.Builder
	var failures []string
	for _, name := range names {
		relPath := "scripts/testdata/writepath/" + name
		findings, err := scanFile(root, relPath)
		if err != nil {
			return Result{Stderr: errorLine(relPath, err), Code: 1}
		}
		rejected := len(findings) > 0
		switch {
		case strings.HasPrefix(name, "reject-"):
			if rejected {
				_, _ = fmt.Fprintf(&stdout, "PASS %s: rejected as expected\n", name)
			} else {
				failures = append(failures, fmt.Sprintf("%s: expected rejection (write primitive), got clean", name))
			}
		case strings.HasPrefix(name, "accept-"):
			if rejected {
				failures = append(failures, fmt.Sprintf("%s: expected clean, got findings: %s", name, strings.Join(findings, "; ")))
			} else {
				_, _ = fmt.Fprintf(&stdout, "PASS %s: clean as expected\n", name)
			}
		default:
			failures = append(failures, fmt.Sprintf("%s: fixture filename must start with 'accept-' or 'reject-'", name))
		}
	}

	if len(failures) > 0 {
		var b strings.Builder
		for _, f := range failures {
			b.WriteString("FAIL " + f + "\n")
		}
		return Result{Stdout: stdout.String(), Stderr: b.String(), Code: 1}
	}

	return Result{Stdout: stdout.String(), Code: 0}
}

// Run dispatches the top-level CLI-style modes exactly like
// scripts/check-write-path-precondition.sh's case statement:
//
//	flag == ""                      -> repo-mode scan only
//	flag == "--self-test"           -> fixture self-test, then repo scan,
//	                                    then a trailing banner on success
//	flag == "--self-test-integrity" -> fixture self-test only, then a
//	                                    trailing banner on success
//	anything else                   -> usage to stderr, exit 2
//
// It writes ordered stdout/stderr bytes to the given writers and returns
// the process-style exit code (M1-T8; M1-T10 wires this into
// `gatecheck write-path`).
func Run(flag, root string, git GitRunner, stdout, stderr io.Writer) int {
	switch flag {
	case "":
		res := runRepoScan(root, git)
		_, _ = io.WriteString(stdout, res.Stdout)
		_, _ = io.WriteString(stderr, res.Stderr)
		return res.Code

	case "--self-test":
		res := runFixtureSelfTest(root)
		_, _ = io.WriteString(stdout, res.Stdout)
		_, _ = io.WriteString(stderr, res.Stderr)
		if res.Code != 0 {
			return res.Code
		}
		repoRes := runRepoScan(root, git)
		_, _ = io.WriteString(stdout, repoRes.Stdout)
		_, _ = io.WriteString(stderr, repoRes.Stderr)
		if repoRes.Code != 0 {
			return repoRes.Code
		}
		_, _ = io.WriteString(stdout, "self-test passed: fixtures matched expectations and the tracked tree is clean\n")
		return 0

	case "--self-test-integrity":
		res := runFixtureSelfTest(root)
		_, _ = io.WriteString(stdout, res.Stdout)
		_, _ = io.WriteString(stderr, res.Stderr)
		if res.Code != 0 {
			return res.Code
		}
		_, _ = io.WriteString(stdout, "self-test-integrity passed: fixtures matched expectations (repo scan skipped, 030.001-T)\n")
		return 0

	default:
		_, _ = io.WriteString(stderr, "usage: scripts/check-write-path-precondition.sh [--self-test|--self-test-integrity]\n")
		return 2
	}
}
