package stagecontract

import (
	"errors"
	"os"
	"path/filepath"
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
