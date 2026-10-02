// Package mergestrategy reimplements the repository's merge-strategy JSON
// evaluator (the evaluate_json function embedded in
// scripts/check-merge-strategy.sh) on top of
// tools/gatecheck/internal/pysem, so a live GitHub repository-settings API
// response (or a committed test fixture) can be evaluated for
// allow_squash_merge / allow_rebase_merge without a Python interpreter
// (plan §6, M3-T9).
//
// This package documents two accepted, scoped Principle III exceptions;
// neither is an undiscovered hole. Run accepts root to satisfy the shared
// subcommand signature, then discards it (_ = root) and reads the supplied
// payload path with os.ReadFile without checking containment. This is required
// because live-transport wrapper temp payloads are created outside the repo
// root, and matches the retired Python evaluate_json opening sys.argv[1].
// Run also evaluates only args[0] and ignores extra positional arguments,
// matching the retired Python's unconditional sys.argv[1]. That behavior is
// intentionally unchanged.
package mergestrategy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// Verdict is the fixed, three-valued vocabulary evaluate_json's own
// emit() helper prints as the first token of its single output line.
type Verdict = string

const (
	// Pass means allow_squash_merge and allow_rebase_merge are both
	// false -- the required, compliant state (P-009).
	Pass Verdict = "PASS"
	// Fail means at least one of allow_squash_merge/allow_rebase_merge
	// is true.
	Fail Verdict = "FAIL"
	// Skip means the response could not be evaluated at all (empty,
	// unparseable, missing keys, or a non-boolean value) -- never
	// silently treated as a pass or a fail.
	Skip Verdict = "SKIP"
)

// exitCodeFor reproduces emit()'s `sys.exit(0 if verdict in ("PASS",
// "SKIP") else 1)` rule.
func exitCodeFor(verdict Verdict) int {
	if verdict == Pass || verdict == Skip {
		return 0
	}
	return 1
}

// readInput reads path (or stdin, when path == "-") and applies the same
// UTF-8-validity-then-universal-newline decoding pysem.ReadText/GitText
// apply elsewhere in this port, so an invalid-UTF-8 payload is rejected
// with pysem.ErrInvalidUTF8 before any JSON decoding is attempted --
// exactly where CPython's `open(src, 'r', encoding='utf-8').read()` would
// have raised UnicodeDecodeError.
func readInput(path string, stdin io.Reader) (string, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	return pysem.GitText(data)
}

// pyTypeName derives the Python type name (NoneType, bool, str, int,
// float, list, dict) evaluate_json's own f-string `{type(x).__name__}`
// would have produced for a decoded JSON value, given a Go `any` decoded
// via json.Decoder with UseNumber(). A json.Number is classified as
// "float" when its literal text contains '.', 'e' or 'E' (matching
// Python's own json.loads float-vs-int literal distinction), otherwise
// "int".
func pyTypeName(v any) string {
	switch val := v.(type) {
	case nil:
		return "NoneType"
	case bool:
		return "bool"
	case string:
		return "str"
	case json.Number:
		if strings.ContainsAny(string(val), ".eE") {
			return "float"
		}
		return "int"
	case []any:
		return "list"
	case map[string]any:
		return "dict"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// evaluate is Evaluate's internal implementation, additionally
// distinguishing a read/decode-input failure (readErr non-empty, ED-2 --
// printed as a raw ::error:: line, never wrapped in "<VERDICT> <reason>")
// from a normal verdict outcome (readErr empty).
func evaluate(path string, stdin io.Reader) (verdict Verdict, reason string, exitCode int, readErr string) {
	text, err := readInput(path, stdin)
	if err != nil {
		if errors.Is(err, pysem.ErrInvalidUTF8) {
			return "", "", 1, fmt.Sprintf("::error::could not read %s as UTF-8 text (invalid UTF-8 bytes)", path)
		}
		return "", "", 1, fmt.Sprintf("::error::could not read %s: %s", path, err)
	}

	if pysem.Strip(text) == "" {
		v := Skip
		return v, "empty API response (unauthorized or unreachable)", exitCodeFor(v), ""
	}

	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	var data any
	if decErr := dec.Decode(&data); decErr != nil {
		v := Skip
		return v, fmt.Sprintf("could not parse API response as JSON: %s", decErr), exitCodeFor(v), ""
	}
	// M3-T9's named red test (ED-8): a second Decode on the SAME decoder
	// must return io.EOF for a clean single JSON value. Anything else --
	// trailing garbage after a valid object, or a second concatenated
	// object/value -- takes the parse-error SKIP path, matching Python's
	// own "Extra data" JSONDecodeError. json.Decoder.Decode alone (without
	// this second call) would silently IGNORE trailing data, turning an
	// input Python treats as SKIP into a fail-open PASS/FAIL -- exactly
	// the P1 finding this second Decode call closes.
	var extra any
	if decErr := dec.Decode(&extra); decErr != io.EOF {
		v := Skip
		return v, "could not parse API response as JSON: Extra data", exitCodeFor(v), ""
	}

	obj, ok := data.(map[string]any)
	if !ok {
		v := Skip
		return v, "allow_squash_merge/allow_rebase_merge absent from API response (unauthorized token or field not exposed to this credential)", exitCodeFor(v), ""
	}
	squash, squashOK := obj["allow_squash_merge"]
	rebase, rebaseOK := obj["allow_rebase_merge"]
	if !squashOK || !rebaseOK {
		v := Skip
		return v, "allow_squash_merge/allow_rebase_merge absent from API response (unauthorized token or field not exposed to this credential)", exitCodeFor(v), ""
	}

	squashBool, squashIsBool := squash.(bool)
	rebaseBool, rebaseIsBool := rebase.(bool)

	if squashIsBool && squashBool && rebaseIsBool && rebaseBool {
		v := Fail
		return v, "both allow_squash_merge and allow_rebase_merge are true", exitCodeFor(v), ""
	}
	if squashIsBool && squashBool {
		v := Fail
		return v, "allow_squash_merge is true", exitCodeFor(v), ""
	}
	if rebaseIsBool && rebaseBool {
		v := Fail
		return v, "allow_rebase_merge is true", exitCodeFor(v), ""
	}
	if !squashIsBool || !rebaseIsBool {
		v := Skip
		return v, fmt.Sprintf(
			"allow_squash_merge/allow_rebase_merge present but not boolean (got %s/%s); malformed response is never a pass",
			pyTypeName(squash), pyTypeName(rebase),
		), exitCodeFor(v), ""
	}
	v := Pass
	return v, "allow_squash_merge and allow_rebase_merge are both false", exitCodeFor(v), ""
}

// Evaluate ports evaluate_json (plan §6, M3-T9): it reads path (or stdin
// when path == "-"), decodes it into an exact-key map[string]any (never a
// struct, which would match keys case-insensitively and silently accept
// e.g. "Allow_Squash_Merge"), and returns the verdict, reason and process
// exit code exactly as evaluate_json's own emit() would have produced --
// except for a read/decode-input failure, which is reported via readErr
// (ED-2: a synthetic ::error:: line replacing Python's uncaught-exception
// traceback) instead of the PASS/FAIL/SKIP vocabulary.
func Evaluate(path string, stdin io.Reader) (verdict Verdict, reason string, exitCode int, readErr string) {
	return evaluate(path, stdin)
}

// Run is mergestrategy's top-level CLI entry point, implementing
// `gatecheck merge-strategy-evaluate <path|->` (plan §6, M3-T9). It takes
// the single path argument (root is accepted only to satisfy the shared
// subcommandFunc signature every tools/gatecheck subcommand implements;
// merge-strategy-evaluate itself, like the Python engine before it, never
// reads anything relative to the repository root).
func Run(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	_ = root
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, "usage: gatecheck merge-strategy-evaluate <path|->")
		return 2
	}
	path := args[0]
	verdict, reason, code, readErr := evaluate(path, stdin)
	if readErr != "" {
		_, _ = fmt.Fprintf(stderr, "%s\n", readErr)
		return code
	}
	_, _ = fmt.Fprintf(stdout, "%s %s\n", verdict, reason)
	return code
}
