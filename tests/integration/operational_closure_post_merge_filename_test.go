package integration

// TestOperationalClosurePostMergeFilenameConformance is the document<->
// artifact conformance harness for 028.002-T ("Add document-to-artifact
// closure conformance test"), shipment 025-S / feature 028-F (plan unit 1),
// per docs/plans/2026-09-18-intercom-go-ship-pipeline-contract-repair-plan.md
// section 5, row U1-T2.
//
// It PARSES the post-merge closure filename convention out of
// .github/skills/operational-closure/SKILL.md's `## Output` section --
// it does not hardcode the documented filename pattern; the only literal
// it relies on is the `*-post-merge-closure.md` glob that defines the
// test universe -- and asserts every
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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// closureOutputBacktickClosurePathPattern extracts a `docs/closure/...` path from inside
// a Markdown inline-code span.
var closureOutputBacktickClosurePathPattern = regexp.MustCompile("`(docs/closure/[^`]+)`")

// closureOutputPlaceholderTokenPattern matches a single {token} placeholder segment
// inside a documented docs/closure/ path template.
var closureOutputPlaceholderTokenPattern = regexp.MustCompile(`\{[^{}]+\}`)

// closureOutputOutputHeadingPattern matches the `## Output` heading only at the start of
// a line (and only as the whole heading), so prose or a longer heading such
// as `## Outputs` that merely contains the text is never mistaken for it.
var closureOutputOutputHeadingPattern = regexp.MustCompile(`(?m)^## Output[ \t]*\r?$`)

// closureOutputNextHeadingPattern matches the start of the next level-2 heading.
var closureOutputNextHeadingPattern = regexp.MustCompile(`(?m)^## `)

// closureOutputExtractOutputSection returns the slice of skillMD strictly between the
// line-start `## Output` heading and the next `## ` heading (or EOF if
// `## Output` is the last section).
func closureOutputExtractOutputSection(skillMD string) (string, error) {
	loc := closureOutputOutputHeadingPattern.FindStringIndex(skillMD)
	if loc == nil {
		return "", fmt.Errorf("line-start %q heading not found in operational-closure/SKILL.md", "## Output")
	}
	rest := skillMD[loc[1]:]
	if next := closureOutputNextHeadingPattern.FindStringIndex(rest); next != nil {
		return rest[:next[0]], nil
	}
	return rest, nil
}

// closureOutputPostMergeCandidatePaths returns, in order and de-duplicated, every
// docs/closure/ backtick path found on a line of section that mentions
// needle (case insensitive).
func closureOutputPostMergeCandidatePaths(section, needle string) []string {
	lowerNeedle := strings.ToLower(needle)
	var out []string
	seen := map[string]bool{}
	for _, line := range strings.Split(section, "\n") {
		if !strings.Contains(strings.ToLower(line), lowerNeedle) {
			continue
		}
		for _, m := range closureOutputBacktickClosurePathPattern.FindAllStringSubmatch(line, -1) {
			if !seen[m[1]] {
				seen[m[1]] = true
				out = append(out, m[1])
			}
		}
	}
	return out
}

// closureOutputFindFirstBacktickClosurePath returns the first docs/closure/ backtick path
// found anywhere in section, regardless of line content.
func closureOutputFindFirstBacktickClosurePath(section string) (string, bool) {
	if m := closureOutputBacktickClosurePathPattern.FindStringSubmatch(section); m != nil {
		return m[1], true
	}
	return "", false
}

// closureOutputParsePostMergeClosurePathTemplate implements the Parse rule above: prefer
// the post-merge-specific Output form; fall back to the generic form.
//
// Disambiguation: several lines may mention "post-merge" (for example a
// sub-bullet naming the committed close-evidence JSON). Only candidate paths
// that themselves contain "post-merge" describe the closure artifact; if
// more than one distinct such path remains the document is ambiguous and the
// parse is an error rather than a silent first-match.
func closureOutputParsePostMergeClosurePathTemplate(skillMD string) (string, error) {
	section, err := closureOutputExtractOutputSection(skillMD)
	if err != nil {
		return "", err
	}

	var forms []string
	for _, p := range closureOutputPostMergeCandidatePaths(section, "post-merge") {
		if strings.Contains(strings.ToLower(p), "post-merge") {
			forms = append(forms, p)
		}
	}
	switch len(forms) {
	case 1:
		return forms[0], nil
	case 0:
	default:
		return "", fmt.Errorf("ambiguous post-merge Output forms in operational-closure/SKILL.md: %q", forms)
	}
	if form, ok := closureOutputFindFirstBacktickClosurePath(section); ok {
		return form, nil
	}
	return "", errors.New("no docs/closure/ path template found in operational-closure/SKILL.md's ## Output section")
}

// closureOutputPlaceholderClass maps a documented {token} name to a permissive regex
// class approximating its real-world shape, without ever hardcoding the
// caller's full filename pattern. An unrecognized token is an error: a
// silent catch-all class would let a renamed or new placeholder pass the
// conformance check unexamined.
//
// Id tokens: matched by a "_id"/"id" name suffix instead of the former
// substring match, because any token merely containing "id" (e.g. "width")
// would otherwise be classed as an id. The class mirrors the real id shape
// (NNN or NNN.NNN, a dash, a letter suffix: 001-S, 058.001-T).
func closureOutputPlaceholderClass(token string) (string, error) {
	switch {
	case token == "YYYY-MM-DD":
		return `\d{4}-\d{2}-\d{2}`, nil
	case token == "slug" || strings.HasSuffix(token, "_slug"):
		return `[a-z0-9][a-z0-9-]*`, nil
	case token == "id" || strings.HasSuffix(token, "_id"):
		return `[0-9]+(?:\.[0-9]+)*-[A-Za-z]+`, nil
	default:
		return "", fmt.Errorf("unrecognized placeholder {%s}: add an explicit class for it", token)
	}
}

// closureOutputCompileClosureFilenameRegex turns a documented docs/closure/ path
// template (e.g. "docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md")
// into an anchored regular expression matching the basename of a real
// closure artifact, substituting each {placeholder} with a permissive class
// and escaping every literal character in between.
func closureOutputCompileClosureFilenameRegex(docPathTemplate string) (*regexp.Regexp, error) {
	const prefix = "docs/closure/"
	if !strings.HasPrefix(docPathTemplate, prefix) {
		return nil, fmt.Errorf("documented Output path %q is not under docs/closure/", docPathTemplate)
	}
	filenameTemplate := strings.TrimPrefix(docPathTemplate, prefix)

	var b strings.Builder
	b.WriteString("^")
	last := 0
	for _, loc := range closureOutputPlaceholderTokenPattern.FindAllStringIndex(filenameTemplate, -1) {
		b.WriteString(regexp.QuoteMeta(filenameTemplate[last:loc[0]]))
		token := filenameTemplate[loc[0]+1 : loc[1]-1]
		class, err := closureOutputPlaceholderClass(token)
		if err != nil {
			return nil, err
		}
		b.WriteString(class)
		last = loc[1]
	}
	b.WriteString(regexp.QuoteMeta(filenameTemplate[last:]))
	b.WriteString("$")

	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil, fmt.Errorf("compiling filename regex from template %q: %w", filenameTemplate, err)
	}
	return re, nil
}
func TestOperationalClosurePostMergeFilenameConformance(t *testing.T) {
	root := repoRoot(t)

	skillPath := filepath.Join(root, ".github", "skills", "operational-closure", "SKILL.md")
	skillBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("reading %s: %v", skillPath, err)
	}

	docPathTemplate, err := closureOutputParsePostMergeClosurePathTemplate(string(skillBytes))
	if err != nil {
		t.Fatalf("parsing documented Output form: %v", err)
	}
	re, err := closureOutputCompileClosureFilenameRegex(docPathTemplate)
	if err != nil {
		t.Fatalf("compiling documented Output form %q: %v", docPathTemplate, err)
	}

	closureDir := filepath.Join(root, "docs", "closure")
	matches, err := filepath.Glob(filepath.Join(closureDir, "*-post-merge-closure.md"))
	if err != nil {
		t.Fatalf("globbing %s: %v", closureDir, err)
	}
	// Decision (058.004-T): an empty universe stays hard-fatal -- a silent
	// zero-artifact pass would hide a renamed glob or a moved docs/closure dir.
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

// --- 058.003-T red-first tests: fatal fallback, anchored heading, ambiguity ---

func TestPlaceholderClass_UnrecognizedTokenIsRejected(t *testing.T) {
	if _, err := closureOutputPlaceholderClass("mystery"); err == nil {
		t.Fatal("closureOutputPlaceholderClass(\"mystery\") succeeded; an unrecognized placeholder must be fatal, not silently permissive")
	}
	for _, token := range []string{"YYYY-MM-DD", "slug", "shipment_id", "feature_id"} {
		if _, err := closureOutputPlaceholderClass(token); err != nil {
			t.Errorf("closureOutputPlaceholderClass(%q) = %v, want a recognized class", token, err)
		}
	}
	// "width" merely contains "id"; the tightened rule must not treat it as an id.
	if _, err := closureOutputPlaceholderClass("width"); err == nil {
		t.Fatal("closureOutputPlaceholderClass(\"width\") succeeded; only *_id tokens are ids")
	}
}

func TestPlaceholderClass_SlugTokenMatchIsExactOrSuffix(t *testing.T) {
	for _, token := range []string{"slug", "post_merge_slug"} {
		if _, err := closureOutputPlaceholderClass(token); err != nil {
			t.Errorf("closureOutputPlaceholderClass(%q) = %v, want a recognized class", token, err)
		}
	}
	for _, token := range []string{"slugger", "slugify", "preslug", "my_slugger"} {
		if _, err := closureOutputPlaceholderClass(token); err == nil {
			t.Errorf("closureOutputPlaceholderClass(%q) succeeded; only slug or *_slug tokens are slugs", token)
		}
	}
}

func TestCompileClosureFilenameRegex_UnrecognizedPlaceholderIsFatal(t *testing.T) {
	if _, err := closureOutputCompileClosureFilenameRegex("docs/closure/{shipment_id}-{mystery}-post-merge-closure.md"); err == nil {
		t.Fatal("closureOutputCompileClosureFilenameRegex accepted an unrecognized placeholder")
	}
}

func TestExtractOutputSection_HeadingAnchoredToLineStart(t *testing.T) {
	// "## Output" mid-line (and as a longer heading) must not be matched.
	md := "intro text mentioning ## Output inline\n## Outputs and more\n## Output\n* `docs/closure/a.md`\n## Next\nx\n"
	got, err := closureOutputExtractOutputSection(md)
	if err != nil {
		t.Fatalf("closureOutputExtractOutputSection: %v", err)
	}
	if !strings.Contains(got, "docs/closure/a.md") || strings.Contains(got, "intro") || strings.Contains(got, "Next") {
		t.Fatalf("section = %q, want only the body of the line-start '## Output' heading", got)
	}
	if _, err := closureOutputExtractOutputSection("prose with ## Output mid-line\n"); err == nil {
		t.Fatal("mid-line '## Output' was accepted as a heading")
	}
}

func TestParsePostMergeClosurePathTemplate_Disambiguation(t *testing.T) {
	section := "* **post-merge** mode: `docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md`\n" +
		"  * post-merge evidence at `docs/closure/evidence/{shipment_id}-{feature_id}-close-evidence.json`\n"
	got, err := closureOutputParsePostMergeClosurePathTemplate("## Output\n\n" + section)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if want := "docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md"; got != want {
		t.Fatalf("got %q, want %q (the evidence path must not win)", got, want)
	}

	ambiguous := "## Output\n* post-merge: `docs/closure/{a}-post-merge-closure.md`\n* post-merge alt: `docs/closure/{b}-post-merge-closure.md`\n"
	if _, err := closureOutputParsePostMergeClosurePathTemplate(ambiguous); err == nil {
		t.Fatal("two distinct post-merge forms must be reported as ambiguous")
	}
}

// TestOperationalClosureGenericOutputFormStaysDocumented guards that the
// generic pre-merge/post-deploy Output form remains documented in the skill
// (058.004-T). The post-merge parse above falls back to the generic form when
// no post-merge-specific one exists; if the generic form silently vanished,
// that fallback (and the RED/GREEN reasoning in the header) would rot.
func TestOperationalClosureGenericOutputFormStaysDocumented(t *testing.T) {
	root := repoRoot(t)
	skillPath := filepath.Join(root, ".github", "skills", "operational-closure", "SKILL.md")
	skillBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("reading %s: %v", skillPath, err)
	}
	section, err := closureOutputExtractOutputSection(string(skillBytes))
	if err != nil {
		t.Fatalf("extracting Output section: %v", err)
	}
	const want = "docs/closure/{YYYY-MM-DD}-{slug}-closure.md"
	for _, line := range strings.Split(section, "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "pre-merge") && strings.Contains(lower, "post-deploy") && strings.Contains(line, "`"+want+"`") {
			return
		}
	}
	t.Fatalf("## Output no longer documents the generic pre-merge/post-deploy form `%s` on a line naming both modes", want)
}
