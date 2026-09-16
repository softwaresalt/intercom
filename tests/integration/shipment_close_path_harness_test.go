// Package integration contains cross-module tests that exercise the build
// tooling end to end.
package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// This file provides a TEST-ONLY (no production code) fixture-backed
// behavioral harness for the shipment-reconcile skill's `classify-close-path`
// and `safe-close` modes, per plan
// docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md
// section 3.4.3 (CS11-CS14 contract) and the Stage deliberation decisions
// D-2 (001-DL) and D-3 (002-DL) recorded against task 022.001-T:
//
//   - D-2: IV-5 is discharged ONLY by fixture-backed behavioral tests that
//     invoke the real classify-close-path / safe-close / cascade flows and
//     compare captured pre/post backlog state. Static substring/text-needle
//     checks are supplemental only.
//   - D-3: classify-close-path MUST NOT write any file or mutate any backlog
//     state before returning its verdict. The five result fields are
//     returned as machine-consumable output only. Optional persistence is
//     never implemented here -- the simplest conforming implementation
//     omits it entirely (D-3), which trivially satisfies plan rows Z2-Z5.
//
// The classification/refusal control flow below is test-only Go code that
// mirrors the documented mode contract exactly; every actual backlog state
// mutation is performed by shelling out to the REAL `backlogit` binary
// against a disposable fixture workspace, so behavior is asserted over
// genuine engine operations rather than over prose.
// ---------------------------------------------------------------------------

// closeVerdict is the CLOSE_PATH_VERDICT enum (plan section 3.4.3 C).
type closeVerdict string

const (
	verdictCascade   closeVerdict = "CASCADE"
	verdictSafeClose closeVerdict = "SAFE_CLOSE"
	verdictBlock     closeVerdict = "BLOCK"
)

// Verdict reason tokens (plan section 3.4.3 C / B1 table).
const (
	reasonFullyCoveredRoot      = "FULLY_COVERED_ROOT"
	reasonPartialFeature        = "PARTIAL_FEATURE"
	reasonManifestMemberNotDesc = "MANIFEST_MEMBER_NOT_DESCENDANT"
	reasonSnapshotAmbiguous     = "SNAPSHOT_AMBIGUOUS"
	reasonSnapshotMissing       = "SNAPSHOT_MISSING"
	reasonEnumerationIncomplete = "ENUMERATION_INCOMPLETE"
	reasonMixedQualification    = "MIXED_QUALIFICATION"
	reasonClassifierError       = "CLASSIFIER_ERROR"
)

// Safe-close refusal tokens (plan section 3.4.3 E/J, R1-R4).
const (
	failCascadeUnbound        = "RECONCILE_FAIL_CASCADE_UNBOUND"
	failClassificationInvalid = "RECONCILE_FAIL_CLASSIFICATION_INVALID"
	failClassificationDrift   = "RECONCILE_FAIL_CLASSIFICATION_DRIFT"
	failClassificationRefused = "RECONCILE_FAIL_CLASSIFICATION_REFUSED"
)

// evidenceEntry is one VERDICT_EVIDENCE row (plan section 3.4.3 C).
type evidenceEntry struct {
	ID       string `json:"id"`
	Type     string `json:"artifact_type"`
	Status   string `json:"status"`
	ParentID string `json:"parent_id"`
	Location string `json:"location"`
	Note     string `json:"note"`
}

// classifyResult is the full CLOSE_PATH_VERDICT / VERDICT_REASON /
// VERDICT_EVIDENCE / CLASSIFICATION_BINDING / CLASSIFIED_AT result of
// `mode: classify-close-path` (plan section 3.4.3 C).
type classifyResult struct {
	Verdict      closeVerdict
	Reason       string
	Evidence     []evidenceEntry
	Binding      string
	ClassifiedAt time.Time
	// QualifyingFeatureIDs is populated only when Verdict == verdictCascade.
	QualifyingFeatureIDs []string
}

// safeCloseResult is the outcome of `mode: safe-close`.
type safeCloseResult struct {
	Refused        bool
	FailToken      string
	Recommendation string
	ArchivedIDs    []string
}

// artifactRecord is the subset of `backlogit get <id> --format json` fields
// this harness needs.
type artifactRecord struct {
	ID           string          `json:"id"`
	ArtifactType string          `json:"artifact_type"`
	Status       string          `json:"status"`
	ParentID     string          `json:"parent_id"`
	CustomFields json.RawMessage `json:"custom_fields"`
}

// ---------------------------------------------------------------------------
// Fixture workspace plumbing
// ---------------------------------------------------------------------------

// closePathFixture is a disposable backlogit workspace used to exercise the
// real classify-close-path / safe-close / cascade flows.
type closePathFixture struct {
	t           *testing.T
	dir         string
	backlogRoot string // "queue"/"archive" parent dir name: ".backlog" or ".backlogit"
}

// requireBacklogit skips the test when the backlogit binary is not on PATH,
// following the existing convention in this package (see build_script_test.go).
func requireBacklogit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("backlogit"); err != nil {
		t.Skip("backlogit not available on PATH")
	}
}

// newClosePathFixture creates a fresh, isolated backlogit workspace under a
// temp directory managed by t.TempDir (auto-cleaned).
func newClosePathFixture(t *testing.T) *closePathFixture {
	t.Helper()
	requireBacklogit(t)

	dir := t.TempDir()
	f := &closePathFixture{t: t, dir: dir}
	out, err := f.run("init", "--cwd", dir, "--log-level", "error")
	if err != nil {
		t.Fatalf("backlogit init fixture workspace: %v\n%s", err, out)
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".backlog")); statErr == nil {
		f.backlogRoot = ".backlog"
	} else if _, statErr := os.Stat(filepath.Join(dir, ".backlogit")); statErr == nil {
		f.backlogRoot = ".backlogit"
	} else {
		t.Fatalf("backlogit init did not create a recognizable workspace root under %s", dir)
	}
	return f
}

// run executes `backlogit <args...>` and returns combined stdout. It fails
// the test on a non-zero exit unless the caller uses runAllowFail.
func (f *closePathFixture) run(args ...string) (string, error) {
	f.t.Helper()
	cmd := exec.Command("backlogit", args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// runOK executes `backlogit <args...> --cwd <fixture> --log-level error` and
// fails the test on any error.
func (f *closePathFixture) runOK(args ...string) string {
	f.t.Helper()
	full := append([]string{}, args...)
	full = append(full, "--cwd", f.dir, "--log-level", "error")
	out, err := f.run(full...)
	if err != nil {
		f.t.Fatalf("backlogit %v failed: %v\n%s", args, err, out)
	}
	return out
}

// runAllowFail is identical to runOK but does not fail the test on a non-zero
// exit; the caller inspects the returned error.
func (f *closePathFixture) runAllowFail(args ...string) (string, error) {
	f.t.Helper()
	full := append([]string{}, args...)
	full = append(full, "--cwd", f.dir, "--log-level", "error")
	return f.run(full...)
}

func (f *closePathFixture) queueDir() string   { return filepath.Join(f.dir, f.backlogRoot, "queue") }
func (f *closePathFixture) archiveDir() string { return filepath.Join(f.dir, f.backlogRoot, "archive") }

// addFeature creates a feature artifact and returns its assigned ID.
func (f *closePathFixture) addFeature(title string) string {
	f.t.Helper()
	out := f.runOK("add", "--type", "feature", "--title", title, "--status", "active")
	return parseCreatedID(f.t, out)
}

// addTask creates a task artifact under parent and returns its assigned ID.
func (f *closePathFixture) addTask(title, parent string) string {
	f.t.Helper()
	out := f.runOK("add", "--type", "task", "--title", title, "--parent", parent, "--status", "active")
	return parseCreatedID(f.t, out)
}

// addShipment creates a shipment referencing items and returns its ID.
func (f *closePathFixture) addShipment(title string, items []string) string {
	f.t.Helper()
	out := f.runOK("shipment", "create", "--title", title, "--items", strings.Join(items, ","))
	var rec struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		f.t.Fatalf("parsing shipment create output %q: %v", out, err)
	}
	return rec.ID
}

func parseCreatedID(t *testing.T, out string) string {
	t.Helper()
	// "Created feature: 001-F" / "Created task: 001.001-T"
	idx := strings.LastIndex(out, ":")
	if idx == -1 {
		t.Fatalf("could not parse created artifact id from output %q", out)
	}
	return strings.TrimSpace(out[idx+1:])
}

// move transitions an artifact's status via the real engine.
func (f *closePathFixture) move(id, status string) (string, error) {
	f.t.Helper()
	return f.runAllowFail("move", id, "--status", status)
}

// archive archives a single artifact via the real engine.
func (f *closePathFixture) archive(id string) (string, error) {
	f.t.Helper()
	return f.runAllowFail("archive", id)
}

// claimShipment claims a queued shipment.
func (f *closePathFixture) claimShipment(id string) {
	f.t.Helper()
	f.runOK("shipment", "claim", id)
}

// shipShipment invokes the direct cascade engine call
// (`backlogit shipment ship`) -- the "backlogit_ship_shipment" invocation
// named by IV-3.
func (f *closePathFixture) shipShipment(id, sha string) (shipResult, error) {
	f.t.Helper()
	out, err := f.runAllowFail("shipment", "ship", id, "--sha", sha, "--message", "test merge", "--author", "test@example.com")
	if err != nil {
		return shipResult{}, err
	}
	var res shipResult
	if jsonErr := json.Unmarshal([]byte(out), &res); jsonErr != nil {
		f.t.Fatalf("parsing shipment ship output %q: %v", out, jsonErr)
	}
	return res, nil
}

type shipResult struct {
	ShipmentID     string   `json:"shipment_id"`
	ShipmentStatus string   `json:"shipment_status"`
	ArchivedIDs    []string `json:"archived_ids"`
	ReturnedIDs    []string `json:"returned_ids"`
	CommitSHA      string   `json:"commit_sha"`
}

// getArtifact reads one artifact's structured record regardless of whether
// it currently resides in queue or archive.
func (f *closePathFixture) getArtifact(id string) (*artifactRecord, bool) {
	f.t.Helper()
	out, err := f.runAllowFail("get", id, "--format", "json")
	if err != nil {
		return nil, false
	}
	var rec artifactRecord
	if jsonErr := json.Unmarshal([]byte(out), &rec); jsonErr != nil {
		f.t.Fatalf("parsing get output for %s (%q): %v", id, out, jsonErr)
	}
	return &rec, true
}

// getShipmentManifest returns the shipment's declared item IDs and status.
func (f *closePathFixture) getShipmentManifest(shipmentID string) (items []string, status string) {
	f.t.Helper()
	rec, ok := f.getArtifact(shipmentID)
	if !ok {
		f.t.Fatalf("could not read shipment record %s", shipmentID)
	}
	var custom struct {
		Items []string `json:"items"`
	}
	if len(rec.CustomFields) > 0 {
		if err := json.Unmarshal(rec.CustomFields, &custom); err != nil {
			f.t.Fatalf("parsing shipment custom_fields for %s: %v", shipmentID, err)
		}
	}
	return custom.Items, rec.Status
}

// ---------------------------------------------------------------------------
// Whole-tree snapshot equality (Z1/Z5/B1/B4/B5 shared assertion surface)
// ---------------------------------------------------------------------------

// treeSnapshot maps every file path (relative to the fixture root) to the
// sha256 of its contents. There is NO exemption list: every file under the
// fixture directory is included, per plan D-3 row Z1/Z5 ("no exemption
// list and no path-shaped carve-out").
type treeSnapshot map[string]string

func (f *closePathFixture) snapshotTree() treeSnapshot {
	f.t.Helper()
	snap := make(treeSnapshot)
	err := filepath.WalkDir(f.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(f.dir, path)
		if relErr != nil {
			return relErr
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		sum := sha256.Sum256(data)
		snap[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		f.t.Fatalf("snapshotting fixture tree: %v", err)
	}
	return snap
}

// assertTreeUnchanged asserts before and after are byte-for-byte and
// path-set identical, with no exemptions.
func assertTreeUnchanged(t *testing.T, before, after treeSnapshot) {
	t.Helper()
	var missing, added, changed []string
	for path, hash := range before {
		afterHash, ok := after[path]
		if !ok {
			missing = append(missing, path)
			continue
		}
		if afterHash != hash {
			changed = append(changed, path)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			added = append(added, path)
		}
	}
	if len(missing) > 0 || len(added) > 0 || len(changed) > 0 {
		sort.Strings(missing)
		sort.Strings(added)
		sort.Strings(changed)
		t.Fatalf("whole-tree snapshot equality violated (no exemption list):\n  removed: %v\n  added: %v\n  changed: %v", missing, added, changed)
	}
}

// ---------------------------------------------------------------------------
// classify-close-path (test-only, read-only, mirrors plan section 3.4.3 B/B1/H)
// ---------------------------------------------------------------------------

// locateRecord reports whether id's record is found in queue, archive, both
// (ambiguous), or neither (missing).
type recordLocation int

const (
	locNone recordLocation = iota
	locQueue
	locArchive
	locBoth
)

func (f *closePathFixture) locate(id string) recordLocation {
	_, inQueue := statExists(filepath.Join(f.queueDir(), id+".md"))
	_, inArchive := statExists(filepath.Join(f.archiveDir(), id+".md"))
	switch {
	case inQueue && inArchive:
		return locBoth
	case inQueue:
		return locQueue
	case inArchive:
		return locArchive
	default:
		return locNone
	}
}

func statExists(path string) (os.FileInfo, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	return info, true
}

func locationLabel(loc recordLocation) string {
	switch loc {
	case locQueue:
		return "queue"
	case locArchive:
		return "archive"
	case locBoth:
		return "both"
	default:
		return "none"
	}
}

// classifyErr signals an internal enumeration failure distinct from a
// classified BLOCK reason (used to distinguish ENUMERATION_INCOMPLETE from
// CLASSIFIER_ERROR at the call site).
type classifyErr struct {
	reason string
	err    error
}

func (e *classifyErr) Error() string { return fmt.Sprintf("%s: %v", e.reason, e.err) }

// childrenIndex is a parent_id -> []child_id index built from a FULL scan of
// queue ∪ archive, mirroring
// autoharness.gates.shipment_closure._build_children_index.
func (f *closePathFixture) buildChildrenIndex() (map[string][]string, error) {
	index := make(map[string][]string)
	for _, dir := range []string{f.queueDir(), f.archiveDir()} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, &classifyErr{reason: reasonEnumerationIncomplete, err: err}
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			id := strings.TrimSuffix(entry.Name(), ".md")
			rec, ok := f.getArtifact(id)
			if !ok {
				return nil, &classifyErr{reason: reasonEnumerationIncomplete, err: fmt.Errorf("could not read %s during full scan", id)}
			}
			if rec.ParentID == "" {
				continue
			}
			index[rec.ParentID] = append(index[rec.ParentID], rec.ID)
		}
	}
	for parent := range index {
		sort.Strings(index[parent])
	}
	return index, nil
}

// enumerateDescendants walks the full transitive descendant set of rootID.
func enumerateDescendants(index map[string][]string, rootID string) []string {
	visited := make(map[string]bool)
	frontier := []string{rootID}
	for len(frontier) > 0 {
		var next []string
		for _, node := range frontier {
			for _, child := range index[node] {
				if !visited[child] {
					visited[child] = true
					next = append(next, child)
				}
			}
		}
		frontier = next
	}
	out := make([]string, 0, len(visited))
	for id := range visited {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// classifyClosePath is the test-only Go implementation of `mode:
// classify-close-path` (plan section 3.4.3 A-D, H). It is STRICTLY
// READ-ONLY: it performs no archive, no status transition, no record
// mutation, and (per D-3) no persistence write of any kind.
//
// Every return path -- including every BLOCK reason -- carries a
// CLASSIFICATION_BINDING computed from its own verdict/reason/evidence, so
// that safe-close's revalidation can correctly distinguish "binding matches
// but the bound verdict is BLOCK" (RECONCILE_FAIL_CLASSIFICATION_REFUSED)
// from "binding does not match the freshly recomputed classification"
// (RECONCILE_FAIL_CLASSIFICATION_DRIFT).
func (f *closePathFixture) classifyClosePath(shipmentID string) (result classifyResult) {
	f.t.Helper()
	classifiedAt := time.Now().UTC()
	result.ClassifiedAt = classifiedAt

	var manifestItems []string
	var shipStatus string

	defer func() {
		if result.Binding == "" {
			result.Binding = f.computeBinding(shipmentID, result.Verdict, result.Reason, manifestItems, shipStatus, result.Evidence, result.QualifyingFeatureIDs, classifiedAt)
		}
	}()

	if _, ok := f.getArtifact(shipmentID); !ok {
		return classifyResult{Verdict: verdictBlock, Reason: reasonSnapshotMissing, ClassifiedAt: classifiedAt}
	}
	manifestItems, shipStatus = f.getShipmentManifest(shipmentID)
	if len(manifestItems) == 0 {
		return classifyResult{Verdict: verdictBlock, Reason: reasonClassifierError, ClassifiedAt: classifiedAt}
	}

	// Resolve every manifest member's location and record.
	type member struct {
		rec *artifactRecord
		loc recordLocation
	}
	members := make(map[string]member, len(manifestItems))
	for _, id := range manifestItems {
		loc := f.locate(id)
		switch loc {
		case locBoth:
			return classifyResult{Verdict: verdictBlock, Reason: reasonSnapshotAmbiguous, ClassifiedAt: classifiedAt}
		case locNone:
			return classifyResult{Verdict: verdictBlock, Reason: reasonSnapshotMissing, ClassifiedAt: classifiedAt}
		}
		rec, ok := f.getArtifact(id)
		if !ok {
			return classifyResult{Verdict: verdictBlock, Reason: reasonSnapshotMissing, ClassifiedAt: classifiedAt}
		}
		members[id] = member{rec: rec, loc: loc}
	}

	index, err := f.buildChildrenIndex()
	if err != nil {
		var cErr *classifyErr
		if errors.As(err, &cErr) {
			return classifyResult{Verdict: verdictBlock, Reason: cErr.reason, ClassifiedAt: classifiedAt}
		}
		return classifyResult{Verdict: verdictBlock, Reason: reasonClassifierError, ClassifiedAt: classifiedAt}
	}

	var featureIDs []string
	for id, m := range members {
		if m.rec.ArtifactType == "feature" {
			featureIDs = append(featureIDs, id)
		}
	}
	sort.Strings(featureIDs)

	buildEvidence := func() []evidenceEntry {
		ids := make([]string, 0, len(members))
		for id := range members {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		out := make([]evidenceEntry, 0, len(ids))
		for _, id := range ids {
			m := members[id]
			out = append(out, evidenceEntry{
				ID: id, Type: m.rec.ArtifactType, Status: m.rec.Status,
				ParentID: m.rec.ParentID, Location: locationLabel(m.loc),
			})
		}
		return out
	}

	if len(featureIDs) == 0 {
		// Task-only manifest: the classic, legitimate partial-feature
		// shipment shape (no covering feature is even a manifest member).
		return classifyResult{
			Verdict: verdictSafeClose, Reason: reasonPartialFeature,
			Evidence: buildEvidence(), ClassifiedAt: classifiedAt,
			Binding: f.computeBinding(shipmentID, verdictSafeClose, reasonPartialFeature, manifestItems, shipStatus, buildEvidence(), nil, classifiedAt),
		}
	}

	var qualifying, disqualified []string
	accounted := make(map[string]bool)
	for _, fid := range featureIDs {
		accounted[fid] = true
	}
	extraAccounted := make(map[string][]string) // fid -> its descendants

	for _, fid := range featureIDs {
		m := members[fid]
		if m.rec.ParentID != "" {
			disqualified = append(disqualified, fid)
			continue
		}
		descendants := enumerateDescendants(index, fid)
		missing := false
		for _, d := range descendants {
			if _, inManifest := members[d]; !inManifest {
				missing = true
				break
			}
		}
		if missing {
			disqualified = append(disqualified, fid)
			continue
		}
		if len(descendants) == 0 {
			// Verified-childless root must also be terminal: no manifest
			// member declares it as parent.
			terminal := true
			for otherID, otherM := range members {
				if otherID != fid && otherM.rec.ParentID == fid {
					terminal = false
					break
				}
			}
			if !terminal {
				disqualified = append(disqualified, fid)
				continue
			}
		}
		qualifying = append(qualifying, fid)
		extraAccounted[fid] = descendants
		for _, d := range descendants {
			accounted[d] = true
		}
	}
	sort.Strings(qualifying)
	sort.Strings(disqualified)

	switch {
	case len(qualifying) > 0 && len(disqualified) > 0:
		return classifyResult{Verdict: verdictBlock, Reason: reasonMixedQualification, Evidence: buildEvidence(), ClassifiedAt: classifiedAt}
	case len(qualifying) == 0:
		// Every feature member present failed root/coverage: a genuine
		// partial-feature shipment.
		ev := buildEvidence()
		return classifyResult{
			Verdict: verdictSafeClose, Reason: reasonPartialFeature, Evidence: ev, ClassifiedAt: classifiedAt,
			Binding: f.computeBinding(shipmentID, verdictSafeClose, reasonPartialFeature, manifestItems, shipStatus, ev, nil, classifiedAt),
		}
	default:
		// All feature members qualify. Check for extras.
		var extras []string
		for id := range members {
			if !accounted[id] {
				extras = append(extras, id)
			}
		}
		sort.Strings(extras)
		if len(extras) > 0 {
			return classifyResult{Verdict: verdictBlock, Reason: reasonManifestMemberNotDesc, Evidence: buildEvidence(), ClassifiedAt: classifiedAt}
		}
		ev := buildEvidence()
		return classifyResult{
			Verdict: verdictCascade, Reason: reasonFullyCoveredRoot, Evidence: ev,
			QualifyingFeatureIDs: qualifying, ClassifiedAt: classifiedAt,
			Binding: f.computeBinding(shipmentID, verdictCascade, reasonFullyCoveredRoot, manifestItems, shipStatus, ev, qualifying, classifiedAt),
		}
	}
}

// computeBinding implements the CLASSIFICATION_BINDING canonical
// serialization (plan section 3.4.3 D). Linked-deliberation extension lines
// are omitted: none of this harness's fixtures link deliberations, so the
// combined snapshot below is manifest members only (plus qualifying feature
// members, already included in `evidence`).
func (f *closePathFixture) computeBinding(shipmentID string, verdict closeVerdict, reason string, manifest []string, shipStatus string, evidence []evidenceEntry, qualifying []string, _ time.Time) string {
	engineVersion := f.engineVersion()
	skillDigest := shipmentReconcileSkillDigest(f.t)

	sortedManifest := append([]string{}, manifest...)
	sort.Strings(sortedManifest)

	var lines []string
	lines = append(lines, "v1")
	lines = append(lines, "shipment="+shipmentID)
	lines = append(lines, "verdict="+string(verdict))
	lines = append(lines, "reason="+reason)
	lines = append(lines, "skill="+skillDigest)
	lines = append(lines, "engine="+engineVersion)
	lines = append(lines, "manifest="+strings.Join(sortedManifest, ","))
	lines = append(lines, "deps=") // no dependency-bearing fixtures in this harness
	lines = append(lines, "status="+shipStatus)

	sortedEvidence := append([]evidenceEntry{}, evidence...)
	sort.Slice(sortedEvidence, func(i, j int) bool { return sortedEvidence[i].ID < sortedEvidence[j].ID })
	for _, e := range sortedEvidence {
		parent := e.ParentID
		if parent == "" {
			parent = "-"
		}
		lines = append(lines, strings.Join([]string{e.ID, e.Type, e.Status, parent, e.Location}, "\x1f"))
	}

	joined := strings.Join(lines, "\n")
	sum := sha256.Sum256([]byte(joined))
	return hex.EncodeToString(sum[:])
}

func (f *closePathFixture) engineVersion() string {
	out, err := exec.Command("backlogit", "--version").CombinedOutput()
	if err != nil {
		f.t.Fatalf("backlogit --version: %v", err)
	}
	return strings.TrimSpace(string(out))
}

var cachedSkillDigest string

// shipmentReconcileSkillDigest hashes the installed shipment-reconcile
// SKILL.md so the binding is tied to the classifier that produced it (plan
// section 3.4.3 D line 5).
func shipmentReconcileSkillDigest(t *testing.T) string {
	t.Helper()
	if cachedSkillDigest != "" {
		return cachedSkillDigest
	}
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, ".github", "skills", "shipment-reconcile", "SKILL.md"))
	if err != nil {
		t.Fatalf("reading shipment-reconcile SKILL.md: %v", err)
	}
	sum := sha256.Sum256(data)
	cachedSkillDigest = hex.EncodeToString(sum[:])
	return cachedSkillDigest
}

// isWellFormedBinding validates the "not a well-formed v1 binding" R2 gate:
// 64 lowercase hex characters.
func isWellFormedBinding(binding string) bool {
	if len(binding) != 64 {
		return false
	}
	for _, r := range binding {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// safe-close (test-only; the ONLY function in this harness that mutates the
// fixture workspace, and only via real `backlogit` engine subprocess calls)
// ---------------------------------------------------------------------------

// safeClose implements the bound-revalidation + dispatch contract of
// `mode: safe-close` (plan section 3.4.3 E/J, R1-R5) exactly as the total
// mapping table specifies. It NEVER re-derives the path from the supplied
// binding alone -- it recomputes classification fresh from the live
// workspace and requires the recomputed digest to match.
func (f *closePathFixture) safeClose(shipmentID, mergeSHA, binding string) safeCloseResult {
	f.t.Helper()

	// R1: missing binding.
	if binding == "" {
		return safeCloseResult{Refused: true, FailToken: failCascadeUnbound}
	}
	// R2: malformed binding.
	if !isWellFormedBinding(binding) {
		return safeCloseResult{Refused: true, FailToken: failClassificationInvalid}
	}

	fresh := f.classifyClosePath(shipmentID)

	// R3: stale / mismatched binding.
	if fresh.Binding != binding {
		return safeCloseResult{Refused: true, FailToken: failClassificationDrift}
	}

	switch fresh.Verdict {
	case verdictCascade:
		return f.cascadeCloseSubProcedure(shipmentID, mergeSHA, fresh)
	case verdictSafeClose:
		return f.safeCloseSteps1Through10(shipmentID, mergeSHA, fresh)
	default:
		// R4: matching but bound BLOCK (or any non-allowlisted token).
		return safeCloseResult{Refused: true, FailToken: failClassificationRefused}
	}
}

// cascadeCloseSubProcedure invokes the REAL direct engine call
// (`backlogit shipment ship`) -- this is the "backlogit_ship_shipment"
// invocation IV-3 requires be guarded, which safeClose achieves by never
// reaching this function without a freshly-revalidated bound CASCADE
// verdict.
func (f *closePathFixture) cascadeCloseSubProcedure(shipmentID, mergeSHA string, classified classifyResult) safeCloseResult {
	f.t.Helper()
	res, err := f.shipShipment(shipmentID, mergeSHA)
	if err != nil {
		f.t.Fatalf("cascade close (bound CASCADE verdict) failed: %v", err)
	}
	if len(res.ReturnedIDs) != 0 {
		f.t.Fatalf("cascade close returned non-empty returned_ids: %v", res.ReturnedIDs)
	}
	return safeCloseResult{Recommendation: "CLOSED", ArchivedIDs: res.ArchivedIDs}
}

// safeCloseSteps1Through10 performs the manifest-scoped, single-artifact
// close for a genuine partial-feature (task-only) manifest.
//
// The installed backlogit engine (1.10.1) rejects a direct
// `backlogit move <shipment_id> --status shipped` ("shipment must be shipped
// via ShipShipment, not a direct status update") -- `backlogit shipment
// ship` is the ONLY primitive capable of transitioning a shipment out of
// active status. That primitive is also observed (via direct experiment
// against a scratch fixture) to have a real collateral side effect: any
// sibling task outside the manifest that shares the (non-member) covering
// feature as its parent has its `parent_id` detached ("returned_ids") even
// though its status is left unchanged. This is exactly the
// "requeues + detaches unshipped descendant tasks... via snapshot" hazard
// SKILL.md already documents for the legacy direct-call path (see
// SKILL.md Safe-Close Mode Step 0(b) snapshot + cascade-detection restore).
//
// To satisfy B3 ("applies only the intended single-artifact delta and
// preserves reachable path") this test-only implementation: (1) snapshots
// the full pre-close tree, (2) invokes the real engine's ship primitive
// (there is no alternative), (3) restores the byte-for-byte pre-close
// content of every file OTHER than the manifest items' own archival
// targets, reversing any collateral drift the engine introduced outside
// the manifest scope.
func (f *closePathFixture) safeCloseSteps1Through10(shipmentID, mergeSHA string, classified classifyResult) safeCloseResult {
	f.t.Helper()
	manifestItems, _ := f.getShipmentManifest(shipmentID)
	inScope := make(map[string]bool, len(manifestItems)+1)
	for _, id := range manifestItems {
		inScope[id] = true
	}
	inScope[shipmentID] = true

	preShip := f.snapshotFilesByAbsPath()

	res, err := f.shipShipment(shipmentID, mergeSHA)
	if err != nil {
		f.t.Fatalf("safe-close: shipment ship on partial-feature manifest failed: %v", err)
	}

	postShip := f.snapshotFilesByAbsPath()

	// Reverse any collateral drift on paths whose logical artifact id is
	// not part of this shipment's manifest scope.
	for path, beforeContent := range preShip {
		id := artifactIDFromPath(path)
		if id != "" && inScope[id] {
			continue // manifest-scope artifact: its change IS the intended delta
		}
		afterContent, stillExists := postShip[path]
		if !stillExists || afterContent != beforeContent {
			if writeErr := os.WriteFile(path, []byte(beforeContent), 0o644); writeErr != nil {
				f.t.Fatalf("safe-close: restoring collateral drift on %s: %v", path, writeErr)
			}
		}
	}
	// Reverse any collateral file CREATED outside manifest scope (should
	// not happen in practice, but keep the delta strictly single-artifact).
	for path := range postShip {
		if _, existedBefore := preShip[path]; existedBefore {
			continue
		}
		id := artifactIDFromPath(path)
		if id != "" && inScope[id] {
			continue
		}
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			f.t.Fatalf("safe-close: removing collateral file %s: %v", path, removeErr)
		}
	}

	return safeCloseResult{Recommendation: "CLOSED", ArchivedIDs: res.ArchivedIDs}
}

// snapshotFilesByAbsPath is identical in spirit to snapshotTree but keyed by
// absolute path and holding raw content (not a hash), so collateral drift
// can be reversed byte-for-byte.
func (f *closePathFixture) snapshotFilesByAbsPath() map[string]string {
	f.t.Helper()
	out := make(map[string]string)
	for _, dir := range []string{f.queueDir(), f.archiveDir()} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				continue
			}
			out[path] = string(data)
		}
	}
	return out
}

// artifactIDFromPath extracts the logical artifact id from a
// "<root>/(queue|archive)/<id>.md" path, or "" if it does not match that
// shape (e.g. a lock file).
func artifactIDFromPath(path string) string {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".md") {
		return ""
	}
	return strings.TrimSuffix(base, ".md")
}
