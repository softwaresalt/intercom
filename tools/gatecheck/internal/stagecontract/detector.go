package stagecontract

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrNoVerdict is returned when a document has no structure the requested
// property can be evaluated against; unknown content yields no verdict.
var ErrNoVerdict = errors.New("stage contract: no verdict")

// Construct is one member of the closed construct set, mapped to a property.
type Construct struct {
	Property    string
	Description string
}

// constructs is the closed construct set. It is deliberately not extensible
// at runtime; a new member requires editing this list and the switch in Detect.
var constructs = []Construct{
	{"P1", "P-010 Stage commit grant is dedicated-Stage-artifact-branch-only"},
	{"P2", "Stage MUST NOT bullet forbidding direct commit or push to the default branch"},
	{"P3", "chore/stage-{scope-slug} create and checkout grant"},
	{"P4", "P-016 Stage spike/research worktree exception"},
	{"P5", "Role Boundary Git and PR rows"},
	{"P6", "registered fail-closed Step 1.9 gate (checklist line plus section)"},
}

// ContractDetector evaluates the closed set of Stage contract properties P1-P6.
type ContractDetector struct{}

// NewContractDetector returns the real Stage contract detector.
func NewContractDetector() *ContractDetector {
	return &ContractDetector{}
}

// Constructs returns the enumerated closed construct set.
func (*ContractDetector) Constructs() []Construct {
	return append([]Construct(nil), constructs...)
}

var (
	actorListRe   = regexp.MustCompile(`^\*\*(Stage|Ship) (MAY|MUST NOT)\*\*`)
	checklistRe   = regexp.MustCompile(`^\[ \] Step \d`)
	stepSectionRe = regexp.MustCompile(`^Step 1\.9\b`)
	defaultRefRe  = regexp.MustCompile("(?i)default branch|`main`|\\bmain\\b")
)

// Detect evaluates f.Property over the document structure of f.Content.
// Verdicts derive from parsed blocks only; quoted, fenced and historical
// blocks are excluded except the canonical Step Sequence checklist read by P6.
func (*ContractDetector) Detect(f Fixture) (Verdict, string, error) {
	blocks, err := Parse(f.Content)
	if err != nil {
		return "", "", fmt.Errorf("parse fixture %s: %w", f.ID, err)
	}
	switch f.Property {
	case "P1":
		return detectP1(blocks)
	case "P2":
		return detectP2(blocks)
	case "P3":
		return detectP3(blocks)
	case "P4":
		return detectP4(blocks)
	case "P5":
		return detectP5(blocks)
	case "P6":
		return detectP6(blocks)
	}
	return "", "", fmt.Errorf("property %q is outside the closed set: %w", f.Property, ErrNoVerdict)
}

func inPath(b Block, sub string) bool {
	for _, h := range b.HeadingPath {
		if strings.Contains(h, sub) {
			return true
		}
	}
	return false
}

// active reports whether a block may contribute to a verdict.
func active(b Block) bool {
	return !b.Quoted && !b.Historical
}

// actorItems returns the list items under the "**<actor> <modality>**" label
// within the P-010 section, and whether that labelled list exists.
func actorItems(blocks []Block, actor, modality string) ([]string, bool) {
	var items []string
	found, current := false, false
	for _, b := range blocks {
		if !active(b) || !inPath(b, "P-010") {
			continue
		}
		switch b.Kind {
		case KindParagraph:
			m := actorListRe.FindStringSubmatch(b.Text)
			current = m != nil && m[1] == actor && m[2] == modality
			found = found || current
		case KindListItem:
			if current {
				items = append(items, b.Text)
			}
		default:
			current = false
		}
	}
	return items, found
}

func hasItem(items []string, prefix string) (string, bool) {
	for _, it := range items {
		if strings.HasPrefix(it, prefix) {
			return it, true
		}
	}
	return "", false
}

func detectP1(blocks []Block) (Verdict, string, error) {
	items, found := actorItems(blocks, "Stage", "MAY")
	if !found {
		return "", "", fmt.Errorf("P1: no Stage MAY list: %w", ErrNoVerdict)
	}
	grant, ok := hasItem(items, "Commit ")
	if !ok || !strings.Contains(grant, "dedicated Stage artifact branch") || !strings.Contains(grant, "only") {
		return Violating, "P1: Stage commit grant targets the default branch instead of the dedicated Stage artifact branch", nil
	}
	return Conforming, "P1: Stage commit grant is limited to the dedicated Stage artifact branch and states never on the default branch", nil
}

func detectP2(blocks []Block) (Verdict, string, error) {
	items, found := actorItems(blocks, "Stage", "MUST NOT")
	if !found {
		return "", "", fmt.Errorf("P2: no Stage MUST NOT list: %w", ErrNoVerdict)
	}
	bullet, ok := hasItem(items, "Commit or push directly to")
	if !ok || !strings.Contains(bullet, "main") {
		return Violating, "P2: Stage MUST NOT list lacks the commit or push directly to main bullet", nil
	}
	for _, b := range blocks {
		if !active(b) && defaultRefRe.MatchString(b.Text) {
			return Conforming, "P2: superseded default-branch wording appears only in quoted, historical and example text and is not a violation", nil
		}
	}
	return Conforming, "P2: Stage MUST NOT list keeps the commit or push directly to main bullet", nil
}

func detectP3(blocks []Block) (Verdict, string, error) {
	items, found := actorItems(blocks, "Stage", "MAY")
	if !found {
		return "", "", fmt.Errorf("P3: no Stage MAY list: %w", ErrNoVerdict)
	}
	grant, ok := hasItem(items, "Create and check out")
	if !ok || !strings.Contains(grant, "chore/stage-{scope-slug}") {
		return Violating, "P3: chore/stage-{scope-slug} create and checkout grant is missing", nil
	}
	return Conforming, "P3: chore/stage-{scope-slug} create and checkout grant is present", nil
}

func detectP4(blocks []Block) (Verdict, string, error) {
	scoped := false
	for _, b := range blocks {
		if b.Kind == KindHeading && strings.Contains(b.Text, "P-016") {
			scoped = true
		}
		if !active(b) || b.Kind != KindParagraph || !inPath(b, "P-016") {
			continue
		}
		if strings.Contains(b.Text, "Allowed exception (Stage spike/research only)") &&
			strings.Contains(b.Text, "worktree") && strings.Contains(b.Text, "time-boxed") {
			return Conforming, "P4: P-016 Stage spike/research worktree exception is intact", nil
		}
	}
	if !scoped {
		return "", "", fmt.Errorf("P4: no P-016 section: %w", ErrNoVerdict)
	}
	return Violating, "P4: P-016 Stage spike/research worktree exception paragraph is removed", nil
}

func detectP5(blocks []Block) (Verdict, string, error) {
	var git, pr *Block
	for i, b := range blocks {
		if b.Kind != KindTableRow || !active(b) || !inPath(b, "Role Boundary") || len(b.Cells) < 2 {
			continue
		}
		switch b.Cells[0] {
		case "Git":
			git = &blocks[i]
		case "PR":
			pr = &blocks[i]
		}
	}
	if git == nil && pr == nil {
		return "", "", fmt.Errorf("P5: no Role Boundary Git or PR rows: %w", ErrNoVerdict)
	}
	if pr == nil || !strings.Contains(pr.Cells[1], "operator") || !strings.Contains(pr.Cells[1], "never Stage") {
		return Violating, "P5: PR row lets Stage push the branch and open the PR", nil
	}
	if git != nil && (!strings.Contains(git.Cells[1], "dedicated Stage artifact branch") || !strings.Contains(git.Cells[1], "on that branch only")) {
		return Violating, "P5: Git row grants Stage commits beyond the dedicated Stage artifact branch", nil
	}
	return Conforming, "P5: PR row says the operator, never Stage, pushes and opens the PR", nil
}

func detectP6(blocks []Block) (Verdict, string, error) {
	checklist, registered, section := false, false, false
	for _, b := range blocks {
		if b.Kind == KindHeading && stepSectionRe.MatchString(strings.TrimSpace(b.Text)) {
			section = true
		}
		// Canonical-checklist exception: the fenced text block under the Step
		// Sequence Contract heading is read; every other fenced block stays excluded.
		if b.Kind != KindFenced || b.Lang != "text" || !inPath(b, "Step Sequence Contract") {
			continue
		}
		for _, l := range b.Lines {
			if !checklistRe.MatchString(l) {
				continue
			}
			checklist = true
			if strings.HasPrefix(l, "[ ] Step 1.9 ") {
				registered = true
			}
		}
	}
	if !checklist {
		return "", "", fmt.Errorf("P6: no Step Sequence checklist: %w", ErrNoVerdict)
	}
	switch {
	case registered && section:
		return Conforming, "P6: Step 1.9 is registered in the Step Sequence checklist and its section is present", nil
	case section:
		return Violating, "P6: Step 1.9 checklist registration line is deleted while the Step 1.9 section remains", nil
	case registered:
		return Violating, "P6: Step 1.9 section is missing while the Step Sequence checklist still registers it", nil
	}
	return Violating, "P6: Step 1.9 is neither registered in the Step Sequence checklist nor present as a section", nil
}
