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
// # Residual evasion surface (034.007-T, AC-4.7)
//
// The detector is the 26 qualified selectors in Selectors. The following
// surface is recorded here, never silently left unhandled:
//
//  1. Named import aliases -- KNOWN OPEN, pending Unit E (feature 049-F,
//     shipment 039-S). An aliased import of a write-capable package
//     (import o "os" -> o.WriteFile(...), or an alias of database/sql /
//     go.etcd.io/bbolt) is not detected.
//  2. pathsafe.NewRoot receiver / Root.Resolve first-caller tracking --
//     KNOWN OPEN, pending Unit E (feature 049-F, shipment 039-S). No
//     tripwire exists. The pathsafe risk-register triggers ("forced the
//     moment a real write path exists", root.go) and feature 038-F's "once
//     a live caller exists" trigger stay awaited, not monitored.
//  3. Dot-imports, blank imports, and local identifiers shadowing a package
//     name (including a local syscall identifier, which the D-2' predicate
//     in occurrenceAllowed trusts by spelling).
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
//  6. Write primitives outside the selector set are not detected:
//     ioutil.WriteFile, ioutil.TempFile, ioutil.TempDir; syscall write and
//     namespace calls other than syscall.CreateFile and syscall.Write
//     (syscall.WriteFile, Open, Unlink, Rename, Mkdir, CreateHardLink,
//     DeleteFile); and golang.org/x/sys/windows and golang.org/x/sys/unix
//     equivalents. None occurs in internal/** or cmd/** at 9b299c8.
//     Widening the selector set is out of D-031-2's scope and is tracked as
//     stash entry 458F9385.
//  7. Selector split by a newline or comment -- KNOWN OPEN, closed by Unit
//     E's AST engine (feature 049-F). Go inserts no semicolon after ".", so
//     "syscall." + newline + "CreateFile(...)" and "os./**/WriteFile(...)"
//     are valid Go that gofmt preserves; after masking neither contains the
//     contiguous selector text, so every selector is evaded.
//
// Residual-risk statement: until feature 049-F ships, this gate is a
// qualified-selector tripwire, not a complete mechanical proof. Items 1, 2,
// 6 and 7 are known open fail-open surfaces. At 9b299c8 there are zero
// aliased write-capable imports, zero production Root.Resolve callers and
// zero item-6 primitives in internal/**/cmd/**, so items 1, 2 and 6 are not
// exploited today. None of the four is mechanically guarded. The
// compensating control is human and agent PR review against this list.
package writepath

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"
	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// Selectors is the ordered list of 26 qualified write-primitive selectors
// this gate detects. The first 20 were ported verbatim from the retired
// Python SELECTORS list (including its three adversarial-review
// additions); the last six (syscall.CreateFile, syscall.Write,
// os.OpenRoot, os.Root, io.CopyN, io.CopyBuffer) were appended by
// 034.004-T so existing finding order is unchanged. syscall.CreateFile is
// reported unless occurrenceAllowed proves the exact metadata-only
// reparse-probe call shape.
var Selectors = []string{
	"os.WriteFile", "os.Create", "os.OpenFile", "os.Remove", "os.RemoveAll",
	"os.Rename", "os.Mkdir", "os.MkdirAll", "os.Symlink", "os.Chmod",
	"os.Truncate", "io.Copy", "sql.Open", "bbolt.Open",
	"os.CreateTemp", "os.MkdirTemp", "os.Link", "os.Chown", "os.Lchown",
	"os.Chtimes",
	"syscall.CreateFile", "syscall.Write", "os.OpenRoot", "os.Root",
	"io.CopyN", "io.CopyBuffer",
}

const (
	exceptionName = "011.003-T's Constitution Check exception (internal/config/validate.go rule 7)"
	registerName  = "the consolidated risk register (internal/pathsafe package doc, root.go)"
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

// findSelector reports whether line contains sel as a qualified selector,
// i.e. whether the first enumerated occurrence exists. It matches
// pattern.search's truthiness semantics and serves as the per-line fast
// path ahead of full enumeration.
func findSelector(line, sel string) bool {
	return nextOccurrence(line, sel, 0) >= 0
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

// advanceCursor checks that line is exactly text[cur:cur+len(line)] and
// returns the offset just past the single line boundary that follows it
// ("\r\n" as two bytes, otherwise the boundary rune's UTF-8 width, or
// nothing at the end of text). It reports false, failing closed, when the
// line does not match the text at cur or the boundary cannot be decoded.
func advanceCursor(text string, cur int, line string) (int, bool) {
	end := cur + len(line)
	if end > len(text) || text[cur:end] != line {
		return 0, false
	}
	if end == len(text) {
		return end, true
	}
	if strings.HasPrefix(text[end:], "\r\n") {
		return end + 2, true
	}
	r, width := utf8.DecodeRuneInString(text[end:])
	if r == utf8.RuneError && width <= 1 {
		return 0, false
	}
	return end + width, true
}

// extentKind classifies the masked text that follows a selector occurrence.
type extentKind int

const (
	// extentNonCall: the selector is not followed (after whitespace) by '('.
	extentNonCall extentKind = iota
	// extentBalanced: the call's parentheses close with every bracket matched.
	extentBalanced
	// extentUnbalanced: the text ends before the call closes (undecidable).
	extentUnbalanced
	// extentMismatched: a closer does not match the innermost opener
	// (undecidable).
	extentMismatched
)

// extent is the call extent of one selector occurrence. segments holds the
// raw interior text split at commas at bracket depth exactly 1 (the call's
// own parentheses); it is nil unless kind is extentBalanced.
type extent struct {
	kind     extentKind
	segments []string
}

// extractExtent extracts the call extent starting at offset after (just
// past a selector occurrence) in the whole masked text: it skips Go
// whitespace, newlines included, requires '(', and walks to the balanced
// ')' while tracking a bracket stack over (), [] and {}. Byte-wise scanning
// is safe because ASCII bytes never occur inside a multi-byte UTF-8
// sequence.
func extractExtent(text string, after int) extent {
	i := after
	for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\r' || text[i] == '\n') {
		i++
	}
	if i >= len(text) || text[i] != '(' {
		return extent{kind: extentNonCall}
	}
	var stack []byte
	var segments []string
	segStart := i + 1
	for j := i; j < len(text); j++ {
		switch c := text[j]; c {
		case '(', '[', '{':
			stack = append(stack, c)
		case ')', ']', '}':
			if stack[len(stack)-1] != bracketOpener[c] {
				return extent{kind: extentMismatched}
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return extent{kind: extentBalanced, segments: append(segments, text[segStart:j])}
			}
		case ',':
			if len(stack) == 1 {
				segments = append(segments, text[segStart:j])
				segStart = j + 1
			}
		}
	}
	return extent{kind: extentUnbalanced}
}

// bracketOpener maps each closing bracket to its opener.
var bracketOpener = map[byte]byte{')': '(', ']': '[', '}': '{'}

// The access-mode allowance (D-2') admits exactly one argument shape: the
// metadata-only syscall.CreateFile call that opens an existing object with
// no access rights and only FILE_FLAG_BACKUP_SEMANTICS.
const (
	allowedCreateFileSelector = "syscall.CreateFile"
	allowedCreateFileArgCount = 7
	allowedCreateFileAccess   = "0"
	allowedCreateFileDispo    = "syscall.OPEN_EXISTING"
	allowedCreateFileFlags    = "syscall.FILE_FLAG_BACKUP_SEMANTICS"
)

// occurrenceAllowed is the D-2' access-mode allowance: it reports whether a
// single selector occurrence, classified by its call extent, is exempt from
// being reported. Every rule is an exact-token match, so the predicate can
// only err toward reporting (fail closed). Rejected shapes include nested
// brackets (call arguments, composite literals, index expressions), visible
// raw strings, blanked string or rune arguments (empty segments), any
// argument count other than 7, any access spelling other than the token 0,
// any disposition other than OPEN_EXISTING, and any flag expression other
// than the bare FILE_FLAG_BACKUP_SEMANTICS.
func occurrenceAllowed(sel string, ext extent) bool {
	if sel != allowedCreateFileSelector || ext.kind != extentBalanced {
		return false
	}
	for _, seg := range ext.segments {
		if strings.ContainsAny(seg, "()[]{}`\"") {
			return false
		}
	}
	args := ext.segments
	if n := len(args); n > 0 && strings.TrimSpace(args[n-1]) == "" {
		args = args[:n-1]
	}
	if len(args) != allowedCreateFileArgCount {
		return false
	}
	for _, arg := range args {
		if strings.TrimSpace(arg) == "" {
			return false
		}
	}
	return strings.TrimSpace(args[1]) == allowedCreateFileAccess &&
		strings.TrimSpace(args[4]) == allowedCreateFileDispo &&
		strings.TrimSpace(args[5]) == allowedCreateFileFlags
}

// lineReportsSelector is the per-line, per-selector evaluator: it reports
// whether line (starting at byte offset lineStart of the whole maskedText)
// yields a finding for sel, i.e. whether any qualified occurrence of sel on
// the line is not allowed. When allowance is false (the line cursor lost
// sync) every occurrence is reported, failing closed.
func lineReportsSelector(maskedText, line string, lineStart int, allowance bool, sel string) bool {
	if !findSelector(line, sel) {
		return false
	}
	if !allowance {
		return true
	}
	for _, pos := range selectorOccurrences(line, sel) {
		if !occurrenceAllowed(sel, extractExtent(maskedText, lineStart+pos+len(sel))) {
			return true
		}
	}
	return false
}

// scanText scans already-decoded, newline-translated, masked text for
// every selector hit, returning one finding string per (line, selector)
// pair in Selectors order, formatted exactly as Python's
// f"{relPath}:{line_no}: write primitive {sel!r} found" (via pysem.Repr).
// relPath is echoed back into the finding text verbatim and must already
// be in the caller's desired display form (forward-slash, repo-relative).
//
// It keeps a byte cursor into maskedText alongside the line loop so each
// occurrence's call extent can be extracted from the whole text; if the
// cursor ever loses sync the allowance is disabled for the rest of the
// text, so every occurrence is reported (fail closed).
func scanText(relPath, maskedText string) []string {
	var findings []string
	cur := 0
	allowance := true
	for i, line := range pysem.SplitLines(maskedText) {
		lineNo := i + 1
		lineStart := cur
		if allowance {
			next, ok := advanceCursor(maskedText, cur, line)
			allowance = ok
			cur = next
		}
		for _, sel := range Selectors {
			if lineReportsSelector(maskedText, line, lineStart, allowance, sel) {
				findings = append(findings, fmt.Sprintf("%s:%d: write primitive %s found", relPath, lineNo, pysem.Repr(sel)))
			}
		}
	}
	return findings
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
		next, ok := advanceCursor(masked, cursor, line)
		if !ok {
			return nil, fmt.Errorf("map masked lines in %q: lost synchronization at byte %d", relPath, cursor)
		}
		cursor = next
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
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		qualifier, ok := selector.X.(*ast.Ident)
		if !ok || qualifier.Name+"."+selector.Sel.Name != allowedCreateFileSelector {
			return true
		}

		tf := fset.File(selector.Pos())
		if tf == nil || tf.Size() != len(src) || selector.End() < tf.Pos(0) || selector.End() > tf.Pos(tf.Size()) {
			return true
		}
		sourceEndOffset := tf.Offset(selector.End())
		maskedEndOffset, ok := maskedOffsetForSourceByte(sourceEndOffset, sourceRuneOffsets, maskedRuneOffsets)
		if ok && occurrenceAllowed(allowedCreateFileSelector, extractExtent(masked, maskedEndOffset)) {
			allowedCalls[selector] = struct{}{}
		}
		return true
	})

	type findingKey struct {
		line          int
		selectorIndex int
	}
	findings := make(map[findingKey]struct{})
	addFinding := func(maskedOffset, selectorIndex int) error {
		line, ok := lineNumberAtOffset(lineStarts, maskedOffset)
		if !ok {
			return fmt.Errorf("map finding in %q: invalid masked byte offset %d", relPath, maskedOffset)
		}
		findings[findingKey{line: line, selectorIndex: selectorIndex}] = struct{}{}
		return nil
	}

	var scanErr error
	ast.Inspect(file, func(node ast.Node) bool {
		if scanErr != nil {
			return false
		}
		switch n := node.(type) {
		case *ast.SelectorExpr:
			qualifier, ok := n.X.(*ast.Ident)
			if !ok {
				return true
			}
			selectorIndex, ok := selectorIndexes[qualifier.Name+"."+n.Sel.Name]
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
			scanErr = addFinding(maskedOffset, selectorIndex)

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
					if err := addFinding(maskedStart+occurrence, selectorIndex); err != nil {
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

	ordered := make([]findingKey, 0, len(findings))
	for finding := range findings {
		ordered = append(ordered, finding)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].line != ordered[j].line {
			return ordered[i].line < ordered[j].line
		}
		return ordered[i].selectorIndex < ordered[j].selectorIndex
	})

	result := make([]string, 0, len(ordered))
	for _, finding := range ordered {
		selector := Selectors[finding.selectorIndex]
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

// runRepoScan performs the repo-mode invariant scan: enumerate every
// tracked internal/**, cmd/** Go source (via git, injectable), scan each,
// and fail closed on a git error, a per-file read/decode error (ED-2), any
// finding, or an empty selection (ED-7).
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
		fs, err := scanFile(root, rel)
		if err != nil {
			return Result{Stderr: errorLine(rel, err), Code: 1}
		}
		findings = append(findings, fs...)
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
