package stagecontract

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConstructSetClosed(t *testing.T) {
	got := NewContractDetector().Constructs()
	want := []string{"P1", "P2", "P3", "P4", "P5", "P6"}
	if len(got) != len(want) {
		t.Fatalf("construct set must be exactly %v, got %d members", want, len(got))
	}
	for i, c := range got {
		if c.Property != want[i] || c.Description == "" {
			t.Errorf("construct %d: want property %s with a description, got %+v", i, want[i], c)
		}
	}
}

func TestDetectNoVerdictOnUnknownContent(t *testing.T) {
	d := NewContractDetector()
	cases := map[string]Fixture{
		"unknown-property": {ID: "x", Property: "P7", Content: []byte("# Anything\n")},
		"unrelated-doc-p1": {ID: "x", Property: "P1", Content: []byte("# Notes\n\nNothing relevant here.\n")},
		"unrelated-doc-p6": {ID: "x", Property: "P6", Content: []byte("# Notes\n\n- item\n")},
	}
	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := d.Detect(f); !errors.Is(err, ErrNoVerdict) {
				t.Fatalf("want ErrNoVerdict, got %v", err)
			}
		})
	}
}

func TestDetectQuotedGrantNotViolating(t *testing.T) {
	doc := "# P-010: Agent Role Boundary\n\n" +
		"**Stage MAY** (within its legitimate scope):\n" +
		"- Commit backlog and planning artifacts on that dedicated Stage artifact branch only\n\n" +
		"> Stage MAY commit backlog and planning artifacts on the default branch.\n\n" +
		"```text\n- Commit planning artifacts on the default branch\n```\n"
	v, reason, err := NewContractDetector().Detect(Fixture{ID: "q", Property: "P1", Content: []byte(doc)})
	if err != nil || v != Conforming {
		t.Fatalf("quoted and fenced wording must not violate: verdict %q reason %q err %v", v, reason, err)
	}
}

func TestDetectP6IgnoresNonCanonicalFence(t *testing.T) {
	doc := "## Step Sequence Contract\n\n```text\n[ ] Step 1 — Triage\n[ ] Step 2 — Deliberation\n```\n\n" +
		"## Unrelated Example\n\n```text\n[ ] Step 1.9 — Stage artifact branch gate\n```\n\n" +
		"### Step 1.9: Stage Artifact Branch Gate\n\nBody.\n"
	v, reason, err := NewContractDetector().Detect(Fixture{ID: "f", Property: "P6", Content: []byte(doc)})
	if err != nil || v != Violating {
		t.Fatalf("a registration outside the canonical checklist must not count: verdict %q reason %q err %v", v, reason, err)
	}
}

func TestLiveContract(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	files := []struct {
		path       string
		properties []string
	}{
		{filepath.Join(".github", "policies", "workflow-policies.md"), []string{"P1", "P2", "P3", "P4"}},
		{filepath.Join(".github", "agents", "_stage.agent.md"), []string{"P5", "P6"}},
	}
	d := NewContractDetector()
	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(root, f.path))
		if err != nil {
			t.Fatalf("live contract file %s is missing or unreadable: %v", f.path, err)
		}
		for _, p := range f.properties {
			v, reason, err := d.Detect(Fixture{ID: f.path, Property: p, Content: content})
			if err != nil {
				t.Errorf("%s %s: detector error: %v", f.path, p, err)
				continue
			}
			if v != Conforming {
				t.Errorf("%s %s: live verdict %q: %s", f.path, p, v, reason)
			}
		}
	}
}

func detectDoc(t *testing.T, prop, doc string) (Verdict, string, error) {
	t.Helper()
	return NewContractDetector().Detect(Fixture{ID: "r", Property: prop, Content: []byte(doc)})
}

const canonicalFence = "## Step Sequence Contract\n\n```text\n[ ] Step 1 — Triage\n[ ] Step 1.9 — Gate\n[ ] Step 2 — Deliberation\n```\n\n"

func TestDetectP6HistoricalSectionHeadingIgnored(t *testing.T) {
	doc := canonicalFence + "## Historical Notes\n\n### Step 1.9: Old Gate\n\nBody.\n"
	v, reason, err := detectDoc(t, "P6", doc)
	if err != nil || v != Violating {
		t.Fatalf("historical Step 1.9 heading must not count as the live section: %q %q %v", v, reason, err)
	}
}

func TestDetectP6HistoricalFenceIgnored(t *testing.T) {
	doc := "## Step Sequence Contract\n\n### Example Checklist\n\n```text\n[ ] Step 1 — Triage\n[ ] Step 1.9 — Gate\n```\n\n### Step 1.9: Gate\n\nBody.\n"
	_, _, err := detectDoc(t, "P6", doc)
	if !errors.Is(err, ErrNoVerdict) {
		t.Fatalf("a historical fence is not the canonical checklist: want ErrNoVerdict, got %v", err)
	}
}

func TestDetectP6FirstCanonicalFenceWins(t *testing.T) {
	doc := "## Step Sequence Contract\n\n```text\n[ ] Step 1 — Triage\n[ ] Step 2 — Deliberation\n```\n\n" +
		"```text\n[ ] Step 1 — Triage\n[ ] Step 1.9 — Gate\n```\n\n### Step 1.9: Gate\n\nBody.\n"
	v, reason, err := detectDoc(t, "P6", doc)
	if err != nil || v != Violating {
		t.Fatalf("a second stale fence must not satisfy registration: %q %q %v", v, reason, err)
	}
}

func TestDetectP5MissingRowNoVerdict(t *testing.T) {
	head := "## Role Boundary (NON-NEGOTIABLE)\n\n| Category | Allowed | Forbidden |\n|---|---|---|\n"
	gitRow := "| Git | dedicated Stage artifact branch; on that branch only | x |\n"
	prRow := "| PR | The operator — never Stage — pushes the Stage artifact branch and opens/approves the resulting merge-commit staging PR | x |\n"
	for name, doc := range map[string]string{"git-only": head + gitRow, "pr-only": head + prRow} {
		t.Run(name, func(t *testing.T) {
			if _, reason, err := detectDoc(t, "P5", doc); !errors.Is(err, ErrNoVerdict) {
				t.Fatalf("want ErrNoVerdict, got reason %q err %v", reason, err)
			}
		})
	}
}

func TestDetectP5DuplicateRowViolates(t *testing.T) {
	doc := "## Role Boundary (NON-NEGOTIABLE)\n\n| Category | Allowed | Forbidden |\n|---|---|---|\n" +
		"| Git | dedicated Stage artifact branch; on that branch only | x |\n" +
		"| PR | Stage pushes and opens the PR | x |\n" +
		"| PR | The operator — never Stage — pushes the Stage artifact branch and opens/approves the resulting merge-commit staging PR | x |\n"
	v, reason, err := detectDoc(t, "P5", doc)
	if err != nil || v != Violating {
		t.Fatalf("any violating duplicate row must violate: %q %q %v", v, reason, err)
	}
}

func TestDetectP5CodeSpanPipe(t *testing.T) {
	doc := "## Role Boundary (NON-NEGOTIABLE)\n\n| Category | Allowed | Forbidden |\n|---|---|---|\n" +
		"| Git | dedicated Stage artifact branch (`a|b`); on that branch only | x |\n" +
		"| PR | The operator — never Stage — pushes the Stage artifact branch and opens/approves the resulting merge-commit staging PR | x |\n"
	v, reason, err := detectDoc(t, "P5", doc)
	if err != nil || v != Conforming {
		t.Fatalf("pipe inside a code span must not split the cell: %q %q %v", v, reason, err)
	}
}

func TestDetectP4HistoricalHeadingNotScoped(t *testing.T) {
	doc := "# Policies\n\n## Historical Examples\n\n### P-016: Old\n\nBody.\n"
	if _, _, err := detectDoc(t, "P4", doc); !errors.Is(err, ErrNoVerdict) {
		t.Fatalf("historical P-016 heading must not scope P4: got %v", err)
	}
}

func TestDetectP3LazyContinuationKeepsLaterItems(t *testing.T) {
	doc := "# P-010: Agent Role Boundary\n\n**Stage MAY** (within scope):\n" +
		"- Commit backlog and planning artifacts on that dedicated Stage artifact branch\nonly, never on the default branch\n" +
		"- Create and check out a dedicated branch (`chore/stage-{scope-slug}`)\n\n" +
		"**Stage MUST NOT**:\n- Commit or push directly to main\n"
	for _, p := range []string{"P1", "P2", "P3"} {
		v, reason, err := detectDoc(t, p, doc)
		if err != nil || v != Conforming {
			t.Errorf("%s: %q %q %v", p, v, reason, err)
		}
	}
}

const p4Intact = "## P-016: No Parallel Branch/Worktree Execution\n\n" +
	"**Allowed exception (Stage spike/research only)**: Stage MAY create or use a separate worktree only for an explicit, time-boxed spike or research investigation during staging. " +
	"That worktree MUST NOT be used for implementation, template/source/config mutation, shipment claim, PR preparation, or Ship execution. " +
	"Stage MUST record the spike/research context and clean up the worktree or hand off findings before Ship execution begins.\n"

func TestDetectP4ExceptionClauses(t *testing.T) {
	cases := map[string]struct {
		doc  string
		want Verdict
	}{
		"intact":               {p4Intact, Conforming},
		"grant-revoked":        {strings.Replace(p4Intact, "Stage MAY create or use", "Stage MUST NOT create or use", 1), Violating},
		"restrictions-deleted": {strings.Replace(p4Intact, "That worktree MUST NOT be used for implementation, template/source/config mutation, shipment claim, PR preparation, or Ship execution. ", "", 1), Violating},
		"restrictions-partial": {strings.Replace(p4Intact, "shipment claim, PR preparation, or ", "", 1), Violating},
		"cleanup-deleted":      {strings.Replace(p4Intact, "Stage MUST record the spike/research context and clean up the worktree or hand off findings before Ship execution begins.", "", 1), Violating},
		"cleanup-partial":      {strings.Replace(p4Intact, " or hand off findings before Ship execution begins", "", 1), Violating},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			v, reason, err := detectDoc(t, "P4", c.doc)
			if err != nil || v != c.want {
				t.Fatalf("want %q, got %q (%s) err %v", c.want, v, reason, err)
			}
		})
	}
}

func TestDetectP5SplitAuthorityViolates(t *testing.T) {
	doc := "## Role Boundary (NON-NEGOTIABLE)\n\n| Category | Allowed | Forbidden |\n|---|---|---|\n" +
		"| Git | dedicated Stage artifact branch; on that branch only | x |\n" +
		"| PR | The operator — never Stage — pushes the Stage artifact branch and delegates opening the staging PR to Stage | x |\n"
	v, reason, err := detectDoc(t, "P5", doc)
	if err != nil || v != Violating {
		t.Fatalf("split push/open authority must violate: %q %q %v", v, reason, err)
	}
}
