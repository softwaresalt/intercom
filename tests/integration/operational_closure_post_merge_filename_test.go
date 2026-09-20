package integration

// TestOperationalClosurePostMergeFilenameConformance is the document<->
// artifact conformance harness for 028.002-T ("Add document-to-artifact
// closure conformance test"), shipment 025-S / feature 028-F (plan unit 1),
// per docs/plans/2026-09-18-intercom-go-ship-pipeline-contract-repair-plan.md
// section 5, row U1-T2.
//
// It PARSES the post-merge closure filename convention out of
// .github/skills/operational-closure/SKILL.md's `## Output` section --
// it never hardcodes the pattern -- and asserts every
// docs/closure/*-post-merge-closure.md artifact conforms to the parsed
// form. Artifacts not matching that glob are outside the test's universe
// (AC-1.5).
//
// Parse rule (AC-1.3, reproducible on both sides of 028.001-T's edit):
// take the post-merge-specific Output form if the `## Output` section
// documents one (a line mentioning "post-merge" carrying a
// `docs/closure/...` backtick path); otherwise fall back to the first
// generic `docs/closure/...` backtick path in the section. Pre-repair,
// only the generic form exists (`docs/closure/{YYYY-MM-DD}-{slug}-closure.md`),
// which matches none of the real `*-post-merge-closure.md` artifacts --
// RED. Post-repair, the post-merge-specific form
// (`docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md`) is
// parsed and matches all of them -- GREEN.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// backtickClosurePathPattern extracts a `docs/closure/...` path from inside
// a Markdown inline-code span.
var backtickClosurePathPattern = regexp.MustCompile("`(docs/closure/[^`]+)`")

// placeholderTokenPattern matches a single {token} placeholder segment
// inside a documented docs/closure/ path template.
var placeholderTokenPattern = regexp.MustCompile(`\{[^{}]+\}`)

// extractOutputSection returns the slice of skillMD strictly between the
// `## Output` heading and the next `## ` heading (or EOF if `## Output` is
// the last section).
func extractOutputSection(t *testing.T, skillMD string) string {
	t.Helper()
	const heading = "## Output"
	startIdx := strings.Index(skillMD, heading)
	if startIdx < 0 {
		t.Fatalf("%q heading not found in operational-closure/SKILL.md", heading)
	}
	rest := skillMD[startIdx+len(heading):]
	if endIdx := strings.Index(rest, "\n## "); endIdx >= 0 {
		return rest[:endIdx]
	}
	return rest
}

// findBacktickPathOnLineContaining scans section line by line (case
// insensitive) for a line containing needle, returning the first
// docs/closure/ backtick path found on such a line.
func findBacktickPathOnLineContaining(section, needle string) (string, bool) {
	lowerNeedle := strings.ToLower(needle)
	for _, line := range strings.Split(section, "\n") {
		if !strings.Contains(strings.ToLower(line), lowerNeedle) {
			continue
		}
		if m := backtickClosurePathPattern.FindStringSubmatch(line); m != nil {
			return m[1], true
		}
	}
	return "", false
}

// findFirstBacktickClosurePath returns the first docs/closure/ backtick path
// found anywhere in section, regardless of line content.
func findFirstBacktickClosurePath(section string) (string, bool) {
	if m := backtickClosurePathPattern.FindStringSubmatch(section); m != nil {
		return m[1], true
	}
	return "", false
}

// parsePostMergeClosurePathTemplate implements the Parse rule above: prefer
// the post-merge-specific Output form; fall back to the generic form.
func parsePostMergeClosurePathTemplate(t *testing.T, skillMD string) string {
	t.Helper()
	section := extractOutputSection(t, skillMD)

	if form, ok := findBacktickPathOnLineContaining(section, "post-merge"); ok {
		return form
	}
	if form, ok := findFirstBacktickClosurePath(section); ok {
		return form
	}
	t.Fatalf("no docs/closure/ path template found in operational-closure/SKILL.md's ## Output section")
	return ""
}

// placeholderClass maps a documented {token} name to a permissive regex
// class approximating its real-world shape, without ever hardcoding the
// caller's full filename pattern.
func placeholderClass(token string) string {
	switch {
	case token == "YYYY-MM-DD":
		return `\d{4}-\d{2}-\d{2}`
	case strings.Contains(token, "slug"):
		return `[a-z0-9][a-z0-9-]*`
	case strings.Contains(token, "id"):
		return `[0-9]+(?:\.[0-9]+)*-[A-Za-z]+`
	default:
		return `[^/]+`
	}
}

// compileClosureFilenameRegex turns a documented docs/closure/ path
// template (e.g. "docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md")
// into an anchored regular expression matching the basename of a real
// closure artifact, substituting each {placeholder} with a permissive class
// and escaping every literal character in between.
func compileClosureFilenameRegex(t *testing.T, docPathTemplate string) *regexp.Regexp {
	t.Helper()
	const prefix = "docs/closure/"
	if !strings.HasPrefix(docPathTemplate, prefix) {
		t.Fatalf("documented Output path %q is not under docs/closure/", docPathTemplate)
	}
	filenameTemplate := strings.TrimPrefix(docPathTemplate, prefix)

	var b strings.Builder
	b.WriteString("^")
	last := 0
	for _, loc := range placeholderTokenPattern.FindAllStringIndex(filenameTemplate, -1) {
		b.WriteString(regexp.QuoteMeta(filenameTemplate[last:loc[0]]))
		token := filenameTemplate[loc[0]+1 : loc[1]-1]
		b.WriteString(placeholderClass(token))
		last = loc[1]
	}
	b.WriteString(regexp.QuoteMeta(filenameTemplate[last:]))
	b.WriteString("$")

	re, err := regexp.Compile(b.String())
	if err != nil {
		t.Fatalf("compiling filename regex from template %q: %v", filenameTemplate, err)
	}
	return re
}

func TestOperationalClosurePostMergeFilenameConformance(t *testing.T) {
	root := repoRoot(t)

	skillPath := filepath.Join(root, ".github", "skills", "operational-closure", "SKILL.md")
	skillBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("reading %s: %v", skillPath, err)
	}

	docPathTemplate := parsePostMergeClosurePathTemplate(t, string(skillBytes))
	re := compileClosureFilenameRegex(t, docPathTemplate)

	closureDir := filepath.Join(root, "docs", "closure")
	matches, err := filepath.Glob(filepath.Join(closureDir, "*-post-merge-closure.md"))
	if err != nil {
		t.Fatalf("globbing %s: %v", closureDir, err)
	}
	if len(matches) == 0 {
		t.Fatalf("no *-post-merge-closure.md artifacts found under %s -- test universe is empty", closureDir)
	}

	for _, m := range matches {
		base := filepath.Base(m)
		if !re.MatchString(base) {
			t.Errorf("artifact %q does not conform to documented Output form %q (compiled regex: %s)", base, docPathTemplate, re.String())
		}
	}
}
