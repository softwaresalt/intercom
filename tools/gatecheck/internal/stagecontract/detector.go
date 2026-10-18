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

// p4Clauses are the grant, implementation restrictions and cleanup requirement of the
// P-016 spike/research worktree exception, matched on whitespace-collapsed text.
var p4Clauses = []string{
	"Stage MAY create or use a separate worktree only for an explicit, time-boxed spike or research investigation",
	"MUST NOT be used for implementation, template/source/config mutation, shipment claim, PR preparation, or Ship execution",
	"clean up the worktree or hand off findings before Ship execution begins",
}

// p5Grant is the operator-owned push and open/approve clause of the PR row.
var p5Grant = regexp.MustCompile(`^The operator — never Stage — pushes the Stage artifact branch and opens(/approves)? the\b`)

func detectP4(blocks []Block) (Verdict, string, error) {
	scoped := false
	for _, b := range blocks {
		if b.Kind == KindHeading && active(b) && strings.Contains(b.Text, "P-016") {
			scoped = true
		}
		if !active(b) || b.Kind != KindParagraph || !inPath(b, "P-016") {
			continue
		}
		if !strings.Contains(b.Text, "Allowed exception (Stage spike/research only)") {
			continue
		}
		text := strings.Join(strings.Fields(b.Text), " ")
		for _, clause := range p4Clauses {
			if !strings.Contains(text, clause) {
				return Violating, "P4: P-016 Stage spike/research worktree exception lacks required clause: " + clause, nil
			}
		}
		return Conforming, "P4: P-016 Stage spike/research worktree exception is intact", nil
	}
	if !scoped {
		return "", "", fmt.Errorf("P4: no P-016 section: %w", ErrNoVerdict)
	}
	return Violating, "P4: P-016 Stage spike/research worktree exception paragraph is removed", nil
}

func detectP5(blocks []Block) (Verdict, string, error) {
	var gitRows, prRows []Block
	for _, b := range blocks {
		if b.Kind != KindTableRow || !active(b) || !inPath(b, "Role Boundary") || len(b.Cells) < 2 {
			continue
		}
		switch b.Cells[0] {
		case "Git":
			gitRows = append(gitRows, b)
		case "PR":
			prRows = append(prRows, b)
		}
	}
	if len(gitRows) == 0 || len(prRows) == 0 {
		return "", "", fmt.Errorf("P5: Role Boundary Git and PR rows are not both present: %w", ErrNoVerdict)
	}
	for _, pr := range prRows {
		if !p5Grant.MatchString(strings.Join(strings.Fields(pr.Cells[1]), " ")) || strings.Contains(pr.Cells[1], "to Stage") {
			return Violating, "P5: PR row lets Stage push the branch and open the PR", nil
		}
	}
	for _, git := range gitRows {
		if !strings.Contains(git.Cells[1], "dedicated Stage artifact branch") || !strings.Contains(git.Cells[1], "on that branch only") {
			return Violating, "P5: Git row grants Stage commits beyond the dedicated Stage artifact branch", nil
		}
	}
	return Conforming, "P5: PR row says the operator, never Stage, pushes and opens the PR", nil
}

// p6Clauses are the branch-before-mutation requirement and the fail-closed
// halt rule of the active Step 1.9 section, matched on whitespace-collapsed text.
var p6Clauses = []string{
	"establishes (or resumes) the dedicated Stage artifact branch",
	"before any tracked Stage artifact mutation this session",
	"the gate halts with the named token `STAGE_BRANCH_GATE_FAIL`",
	"performs no tracked mutation",
}

// p6BodyMissing returns the first required clause absent from the active
// Step 1.9 section body, or "" when the body carries every clause.
func p6BodyMissing(blocks []Block) string {
	var parts []string
	for _, b := range blocks {
		if active(b) && b.Kind != KindHeading && inPath(b, "Step 1.9") {
			parts = append(parts, b.Text)
		}
	}
	body := strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
	for _, c := range p6Clauses {
		if !strings.Contains(body, c) {
			return c
		}
	}
	return ""
}

func detectP6(blocks []Block) (Verdict, string, error) {
	checklist, registered, section := false, false, false
	for _, b := range blocks {
		if b.Kind == KindHeading && active(b) && stepSectionRe.MatchString(strings.TrimSpace(b.Text)) {
			section = true
		}
		// Canonical-checklist exception: the first non-historical text fence under
		// the Step Sequence Contract heading that holds checklist lines is read;
		// every other fenced block, including later ones, stays excluded.
		if checklist || b.Kind != KindFenced || b.Historical || b.Lang != "text" || !inPath(b, "Step Sequence Contract") {
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
		if missing := p6BodyMissing(blocks); missing != "" {
			return Violating, "P6: Step 1.9 section body lacks required clause: " + missing, nil
		}
		return Conforming, "P6: Step 1.9 is registered in the Step Sequence checklist and its section is present", nil
	case section:
		return Violating, "P6: Step 1.9 checklist registration line is deleted while the Step 1.9 section remains", nil
	case registered:
		return Violating, "P6: Step 1.9 section is missing while the Step Sequence checklist still registers it", nil
	}
	return Violating, "P6: Step 1.9 is neither registered in the Step Sequence checklist nor present as a section", nil
}
