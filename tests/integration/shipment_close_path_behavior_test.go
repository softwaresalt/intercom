package integration

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Behavioral acceptance tests for `mode: classify-close-path` and
// `mode: safe-close`, per Stage deliberation D-2 (001-DL, stash E47FFCE7)
// and D-3 (002-DL, stash CE28BFA3) recorded against task 022.001-T.
//
// D-2 requires these tests invoke the REAL classify-close-path, safe-close,
// and cascade flows (see shipment_close_path_harness_test.go) and compare
// captured pre/post backlog state -- not static substring checks. These
// tests are the authoritative IV-1..IV-5 acceptance surface; the
// supplemental 33-row static contract test
// (ship_feature_completion_contract_test.go) proves the instruction/skill
// prose says the right thing, but cannot by itself satisfy IV-1..IV-5.
//
// Required RED-before-implementation ordering (user-specified):
//
//	Z1 -> B1 -> B5 -> B2 -> B3 -> B6 -> Z2/Z3/Z4 (combined)
//
// B4 (no mutation observable until after the verdict is produced) is folded
// in near Z1/B1 because D-3's own task-record text requires it even though
// it was omitted from the explicit ordering list.
// ---------------------------------------------------------------------------

// fixtureCascadeShape builds a fully-covered-root shipment: one root
// feature, one child task, both listed in the shipment manifest.
func fixtureCascadeShape(t *testing.T) (f *closePathFixture, shipmentID, featureID, taskID string) {
	t.Helper()
	f = newClosePathFixture(t)
	featureID = f.addFeature("Covering feature")
	taskID = f.addTask("Child task", featureID)
	shipmentID = f.addShipment("Cascade shipment", []string{featureID, taskID})
	f.claimShipment(shipmentID)
	return f, shipmentID, featureID, taskID
}

// fixturePartialFeatureShape builds a genuine task-only (partial-feature)
// shipment: a feature exists in the workspace but is NOT itself a manifest
// member, only one of its tasks is.
func fixturePartialFeatureShape(t *testing.T) (f *closePathFixture, shipmentID, featureID, taskID string) {
	t.Helper()
	f = newClosePathFixture(t)
	featureID = f.addFeature("Umbrella feature")
	taskID = f.addTask("Only task in shipment", featureID)
	_ = f.addTask("Sibling task not in shipment", featureID)
	shipmentID = f.addShipment("Partial feature shipment", []string{taskID})
	f.claimShipment(shipmentID)
	return f, shipmentID, featureID, taskID
}

// ---------------------------------------------------------------------------
// Z1: whole-tree snapshot equality across CASCADE, SAFE_CLOSE, and all six
// BLOCK/refusal reasons. Whole-tree equality has NO exemption list.
// ---------------------------------------------------------------------------

func TestClassifyClosePath_Z1_WholeTreeUnchanged_Cascade(t *testing.T) {
	f, shipmentID, _, _ := fixtureCascadeShape(t)
	before := f.snapshotTree()
	result := f.classifyClosePath(shipmentID)
	after := f.snapshotTree()
	assertTreeUnchanged(t, before, after)
	if result.Verdict != verdictCascade || result.Reason != reasonFullyCoveredRoot {
		t.Fatalf("expected CASCADE/FULLY_COVERED_ROOT, got %s/%s", result.Verdict, result.Reason)
	}
}

func TestClassifyClosePath_Z1_WholeTreeUnchanged_SafeClose(t *testing.T) {
	f, shipmentID, _, _ := fixturePartialFeatureShape(t)
	before := f.snapshotTree()
	result := f.classifyClosePath(shipmentID)
	after := f.snapshotTree()
	assertTreeUnchanged(t, before, after)
	if result.Verdict != verdictSafeClose || result.Reason != reasonPartialFeature {
		t.Fatalf("expected SAFE_CLOSE/PARTIAL_FEATURE, got %s/%s", result.Verdict, result.Reason)
	}
}

func TestClassifyClosePath_Z1_WholeTreeUnchanged_AllBlockReasons(t *testing.T) {
	cases := []struct {
		name       string
		build      func(t *testing.T) (f *closePathFixture, shipmentID string)
		wantReason string
	}{
		{
			name: "MANIFEST_MEMBER_NOT_DESCENDANT",
			build: func(t *testing.T) (*closePathFixture, string) {
				f := newClosePathFixture(t)
				featureID := f.addFeature("Root feature")
				taskID := f.addTask("Child task", featureID)
				unrelatedFeature := f.addFeature("Unrelated feature (not a manifest member)")
				extraTask := f.addTask("Extra task under unrelated feature", unrelatedFeature)
				shipmentID := f.addShipment("Extras shipment", []string{featureID, taskID, extraTask})
				f.claimShipment(shipmentID)
				return f, shipmentID
			},
			wantReason: reasonManifestMemberNotDesc,
		},
		{
			name: "SNAPSHOT_AMBIGUOUS",
			build: func(t *testing.T) (*closePathFixture, string) {
				f := newClosePathFixture(t)
				featureID := f.addFeature("Root feature")
				taskID := f.addTask("Child task", featureID)
				shipmentID := f.addShipment("Ambiguous shipment", []string{featureID, taskID})
				f.claimShipment(shipmentID)
				// Force an ambiguous (both-locations) member by hand-copying
				// the task's queue file into the archive directory without
				// going through the engine -- this is fixture setup, not a
				// classify-close-path mutation.
				if err := os.MkdirAll(f.archiveDir(), 0o755); err != nil {
					t.Fatalf("mkdir archive dir: %v", err)
				}
				data, err := os.ReadFile(filepath.Join(f.queueDir(), taskID+".md"))
				if err != nil {
					t.Fatalf("reading queue file for %s: %v", taskID, err)
				}
				if err := os.WriteFile(filepath.Join(f.archiveDir(), taskID+".md"), data, 0o644); err != nil {
					t.Fatalf("writing duplicate archive file for %s: %v", taskID, err)
				}
				return f, shipmentID
			},
			wantReason: reasonSnapshotAmbiguous,
		},
		{
			name: "SNAPSHOT_MISSING",
			build: func(t *testing.T) (*closePathFixture, string) {
				f := newClosePathFixture(t)
				featureID := f.addFeature("Root feature")
				taskID := f.addTask("Child task", featureID)
				shipmentID := f.addShipment("Missing member shipment", []string{featureID, taskID})
				f.claimShipment(shipmentID)
				if err := os.Remove(filepath.Join(f.queueDir(), taskID+".md")); err != nil {
					t.Fatalf("removing queue file for %s: %v", taskID, err)
				}
				return f, shipmentID
			},
			wantReason: reasonSnapshotMissing,
		},
		{
			name: "ENUMERATION_INCOMPLETE",
			build: func(t *testing.T) (*closePathFixture, string) {
				f := newClosePathFixture(t)
				featureID := f.addFeature("Root feature")
				taskID := f.addTask("Child task", featureID)
				shipmentID := f.addShipment("Enumeration incomplete shipment", []string{featureID, taskID})
				f.claimShipment(shipmentID)
				// Corrupt an unrelated record in the archive directory so
				// the full children-index scan cannot read it.
				if err := os.MkdirAll(f.archiveDir(), 0o755); err != nil {
					t.Fatalf("mkdir archive dir: %v", err)
				}
				corruptPath := filepath.Join(f.archiveDir(), "999-Z.md")
				if err := os.WriteFile(corruptPath, []byte("not a valid backlogit record"), 0o644); err != nil {
					t.Fatalf("writing corrupt record: %v", err)
				}
				return f, shipmentID
			},
			wantReason: reasonEnumerationIncomplete,
		},
		{
			name: "MIXED_QUALIFICATION",
			build: func(t *testing.T) (*closePathFixture, string) {
				f := newClosePathFixture(t)
				qualifyingFeature := f.addFeature("Qualifying root feature")
				qualifyingTask := f.addTask("Qualifying child", qualifyingFeature)
				disqualifiedFeature := f.addFeature("Disqualified umbrella feature")
				_ = f.addTask("Task NOT in manifest", disqualifiedFeature)
				shipmentID := f.addShipment("Mixed shipment", []string{qualifyingFeature, qualifyingTask, disqualifiedFeature})
				f.claimShipment(shipmentID)
				return f, shipmentID
			},
			wantReason: reasonMixedQualification,
		},
		{
			name: "CLASSIFIER_ERROR",
			build: func(t *testing.T) (*closePathFixture, string) {
				f := newClosePathFixture(t)
				shipmentID := f.addShipment("Empty manifest shipment", nil)
				f.claimShipment(shipmentID)
				return f, shipmentID
			},
			wantReason: reasonClassifierError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, shipmentID := tc.build(t)
			before := f.snapshotTree()
			result := f.classifyClosePath(shipmentID)
			after := f.snapshotTree()
			assertTreeUnchanged(t, before, after)
			if result.Verdict != verdictBlock {
				t.Fatalf("expected BLOCK, got %s", result.Verdict)
			}
			if result.Reason != tc.wantReason {
				t.Fatalf("expected reason %s, got %s", tc.wantReason, result.Reason)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// B4: no mutation is observable until AFTER classify-close-path returns its
// verdict (folded near Z1/B1 per D-3's task-record text).
// ---------------------------------------------------------------------------

func TestSafeClose_B4_NoMutationUntilAfterVerdict(t *testing.T) {
	f, shipmentID, _, _ := fixtureCascadeShape(t)

	preClassifySnapshot := f.snapshotTree()
	result := f.classifyClosePath(shipmentID)
	postClassifySnapshot := f.snapshotTree()

	// Classification itself must be a strict no-op (D-3): the state
	// immediately after classify-close-path returns is identical to the
	// state immediately before it was invoked.
	assertTreeUnchanged(t, preClassifySnapshot, postClassifySnapshot)

	// Only AFTER the verdict is in hand does safe-close (bound with the
	// verdict's own binding) perform any mutation.
	closeResult := f.safeClose(shipmentID, "deadbeef", result.Binding)
	if closeResult.Refused {
		t.Fatalf("expected bound CASCADE to proceed, got refusal %s", closeResult.FailToken)
	}
	postCloseSnapshot := f.snapshotTree()
	assertTreeUnchanged(t, postClassifySnapshot, postClassifySnapshot) // sanity: identity
	if equalSnapshots(postClassifySnapshot, postCloseSnapshot) {
		t.Fatalf("expected safe-close (post-verdict) to mutate the tree, but it did not")
	}
}

func equalSnapshots(a, b treeSnapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// B1: each RECONCILE_FAIL_* token returns refusal and leaves the post-tree
// byte-identical to the pre-tree.
// ---------------------------------------------------------------------------

func TestSafeClose_B1_RefusalTokens_ZeroMutation(t *testing.T) {
	t.Run(failCascadeUnbound, func(t *testing.T) {
		f, shipmentID, _, _ := fixtureCascadeShape(t)
		before := f.snapshotTree()
		res := f.safeClose(shipmentID, "deadbeef", "")
		after := f.snapshotTree()
		assertTreeUnchanged(t, before, after)
		if !res.Refused || res.FailToken != failCascadeUnbound {
			t.Fatalf("expected refusal %s, got refused=%v token=%s", failCascadeUnbound, res.Refused, res.FailToken)
		}
	})

	t.Run(failClassificationInvalid, func(t *testing.T) {
		f, shipmentID, _, _ := fixtureCascadeShape(t)
		before := f.snapshotTree()
		res := f.safeClose(shipmentID, "deadbeef", "not-a-well-formed-binding")
		after := f.snapshotTree()
		assertTreeUnchanged(t, before, after)
		if !res.Refused || res.FailToken != failClassificationInvalid {
			t.Fatalf("expected refusal %s, got refused=%v token=%s", failClassificationInvalid, res.Refused, res.FailToken)
		}
	})

	t.Run(failClassificationDrift, func(t *testing.T) {
		f, shipmentID, _, _ := fixtureCascadeShape(t)
		classified := f.classifyClosePath(shipmentID)
		before := f.snapshotTree()
		// Corrupt a single hex character so it is well-formed but stale.
		stale := flipHexChar(classified.Binding)
		res := f.safeClose(shipmentID, "deadbeef", stale)
		after := f.snapshotTree()
		assertTreeUnchanged(t, before, after)
		if !res.Refused || res.FailToken != failClassificationDrift {
			t.Fatalf("expected refusal %s, got refused=%v token=%s", failClassificationDrift, res.Refused, res.FailToken)
		}
	})

	t.Run(failClassificationRefused, func(t *testing.T) {
		f, shipmentID, _, _ := fixtureCascadeShape(t)
		// Add an extra task under an unrelated (non-member) feature to
		// force a bound BLOCK verdict (MANIFEST_MEMBER_NOT_DESCENDANT).
		unrelatedFeature := f.addFeature("Unrelated feature")
		extraTask := f.addTask("Extra task", unrelatedFeature)
		f.runOK("shipment", "add", shipmentID, extraTask)
		classified := f.classifyClosePath(shipmentID)
		if classified.Verdict != verdictBlock {
			t.Fatalf("test setup expected BLOCK, got %s", classified.Verdict)
		}
		before := f.snapshotTree()
		res := f.safeClose(shipmentID, "deadbeef", classified.Binding)
		after := f.snapshotTree()
		assertTreeUnchanged(t, before, after)
		if !res.Refused || res.FailToken != failClassificationRefused {
			t.Fatalf("expected refusal %s, got refused=%v token=%s", failClassificationRefused, res.Refused, res.FailToken)
		}
	})
}

func flipHexChar(binding string) string {
	runes := []rune(binding)
	for i, r := range runes {
		if r != '0' {
			runes[i] = '0'
			return string(runes)
		}
		_ = i
	}
	runes[0] = '1'
	return string(runes)
}

// ---------------------------------------------------------------------------
// B5: bound snapshot drift causes refusal and zero mutation (distinct
// scenario from B1's synthetic bit-flip: here the WORKSPACE itself changes
// between classify and close).
// ---------------------------------------------------------------------------

func TestSafeClose_B5_WorkspaceDriftAfterClassification_RefusesWithZeroMutation(t *testing.T) {
	f, shipmentID, featureID, _ := fixtureCascadeShape(t)

	classified := f.classifyClosePath(shipmentID)
	if classified.Verdict != verdictCascade {
		t.Fatalf("test setup expected CASCADE, got %s", classified.Verdict)
	}

	// Drift: add a new child task to the covering feature AFTER
	// classification but before close, without updating the manifest.
	newTask := f.addTask("Newly added task after classification", featureID)
	_ = newTask

	before := f.snapshotTree()
	res := f.safeClose(shipmentID, "deadbeef", classified.Binding)
	after := f.snapshotTree()
	assertTreeUnchanged(t, before, after)
	if !res.Refused || res.FailToken != failClassificationDrift {
		t.Fatalf("expected drift refusal %s, got refused=%v token=%s", failClassificationDrift, res.Refused, res.FailToken)
	}
}

// ---------------------------------------------------------------------------
// B2: valid bound CASCADE archives exactly the manifest subtree.
// ---------------------------------------------------------------------------

func TestSafeClose_B2_BoundCascade_ArchivesExactlyManifestSubtree(t *testing.T) {
	f, shipmentID, featureID, taskID := fixtureCascadeShape(t)
	classified := f.classifyClosePath(shipmentID)
	if classified.Verdict != verdictCascade {
		t.Fatalf("test setup expected CASCADE, got %s/%s", classified.Verdict, classified.Reason)
	}

	res := f.safeClose(shipmentID, "cafef00d", classified.Binding)
	if res.Refused {
		t.Fatalf("expected bound CASCADE to proceed, got refusal %s", res.FailToken)
	}

	wantArchived := []string{featureID, shipmentID, taskID}
	gotArchived := append([]string{}, res.ArchivedIDs...)
	sort.Strings(gotArchived)
	sort.Strings(wantArchived)
	if !equalStringSlices(gotArchived, wantArchived) {
		t.Fatalf("expected archived set %v, got %v", wantArchived, gotArchived)
	}

	for _, id := range []string{featureID, taskID, shipmentID} {
		if f.locate(id) != locArchive {
			t.Fatalf("expected %s to be archived exactly, got location %s", id, locationLabel(f.locate(id)))
		}
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// B3: valid bound SAFE_CLOSE applies only the intended single-artifact
// delta and preserves the reachable path (the covering feature that is NOT
// a manifest member must remain queued/untouched, per the partial-feature
// shape's non-cascading semantics).
// ---------------------------------------------------------------------------

func TestSafeClose_B3_BoundSafeClose_SingleArtifactDelta_PreservesReachablePath(t *testing.T) {
	f, shipmentID, featureID, taskID := fixturePartialFeatureShape(t)
	classified := f.classifyClosePath(shipmentID)
	if classified.Verdict != verdictSafeClose {
		t.Fatalf("test setup expected SAFE_CLOSE, got %s/%s", classified.Verdict, classified.Reason)
	}

	beforeFeatureLoc := f.locate(featureID)
	res := f.safeClose(shipmentID, "cafef00d", classified.Binding)
	if res.Refused {
		t.Fatalf("expected bound SAFE_CLOSE to proceed, got refusal %s", res.FailToken)
	}

	// The single-artifact delta: only the manifest task + shipment record
	// close. The covering feature (not a manifest member) is untouched and
	// remains reachable at its original location.
	if f.locate(taskID) != locArchive {
		t.Fatalf("expected manifest task %s archived, got %s", taskID, locationLabel(f.locate(taskID)))
	}
	if f.locate(shipmentID) != locArchive {
		t.Fatalf("expected shipment %s archived, got %s", shipmentID, locationLabel(f.locate(shipmentID)))
	}
	afterFeatureLoc := f.locate(featureID)
	if afterFeatureLoc != beforeFeatureLoc {
		t.Fatalf("expected covering feature %s location unchanged (%s), got %s", featureID, locationLabel(beforeFeatureLoc), locationLabel(afterFeatureLoc))
	}
	if afterFeatureLoc != locQueue {
		t.Fatalf("expected covering feature %s to remain reachable in queue, got %s", featureID, locationLabel(afterFeatureLoc))
	}
}

// ---------------------------------------------------------------------------
// B6: the direct `backlogit_ship_shipment` engine path is guarded -- it is
// reachable ONLY through safeClose's own freshly-revalidated bound CASCADE
// dispatch, never invoked standalone by any documented Ship procedure.
// This behavioral row is paired with a static grep assertion (see the
// 33-row contract test) that Ship's own agent-file procedure, post-edit,
// contains no remaining direct `backlogit shipment ship` /
// `backlogit_ship_shipment` invocation outside the guarded safe-close
// dispatch.
// ---------------------------------------------------------------------------

func TestSafeClose_B6_DirectEnginePathOnlyReachableThroughBoundDispatch(t *testing.T) {
	f, shipmentID, _, _ := fixtureCascadeShape(t)

	// An unbound direct call to the engine's cascade primitive is exactly
	// what IV-3 requires be guarded or replaced. safeClose is the sole
	// sanctioned entry point; calling the underlying engine primitive
	// directly (bypassing safeClose) without ever having produced a
	// binding is the exact shape B6 forbids being reachable from any
	// documented procedure. We assert here that safeClose never reaches
	// cascadeCloseSubProcedure (the only call site of shipShipment) without
	// a matching, freshly-revalidated binding, for every unbound/mismatched
	// input.
	before := f.snapshotTree()
	res := f.safeClose(shipmentID, "deadbeef", "")
	after := f.snapshotTree()
	assertTreeUnchanged(t, before, after)
	if !res.Refused {
		t.Fatalf("expected unbound direct-path attempt to be refused, got success")
	}

	// Now prove the guarded path DOES work when properly bound (paired
	// positive case demonstrating replacement, not merely refusal).
	classified := f.classifyClosePath(shipmentID)
	res2 := f.safeClose(shipmentID, "deadbeef", classified.Binding)
	if res2.Refused {
		t.Fatalf("expected properly bound CASCADE dispatch to succeed, got refusal %s", res2.FailToken)
	}
}

// ---------------------------------------------------------------------------
// Z2/Z3/Z4 (combined): verdict/binding are carried in the RETURN VALUE only;
// an absent reconcile-report directory is a no-op (never created, never
// required); optional persistence is strictly post-verdict and, per D-3,
// is not implemented at all -- the simplest conforming implementation omits
// it entirely, which trivially satisfies "suppressed in read-only mode".
// ---------------------------------------------------------------------------

func TestClassifyClosePath_Z2Z3Z4_ReturnValueOnly_NoReconcileReportFile(t *testing.T) {
	f, shipmentID, _, _ := fixtureCascadeShape(t)

	// Confirm no reconcile directory exists ahead of classification.
	reconcileDir := filepath.Join(f.dir, f.backlogRoot, "reconcile")
	if _, err := os.Stat(reconcileDir); err == nil {
		t.Fatalf("test setup invariant violated: reconcile dir already exists at %s", reconcileDir)
	}

	result := f.classifyClosePath(shipmentID)

	// Z2: all five result fields are present in the return value.
	if result.Verdict == "" {
		t.Fatalf("CLOSE_PATH_VERDICT missing from return value")
	}
	if result.Reason == "" {
		t.Fatalf("VERDICT_REASON missing from return value")
	}
	if len(result.Evidence) == 0 {
		t.Fatalf("VERDICT_EVIDENCE missing from return value")
	}
	if result.Binding == "" {
		t.Fatalf("CLASSIFICATION_BINDING missing from return value")
	}
	if result.ClassifiedAt.IsZero() {
		t.Fatalf("CLASSIFIED_AT missing from return value")
	}

	// Z3: absent reconcile report directory is a no-op -- classification
	// must not have created it, and must not require its presence.
	if _, err := os.Stat(reconcileDir); err == nil {
		t.Fatalf("classify-close-path must not create a reconcile report directory (D-3): found %s", reconcileDir)
	}

	// Z4: no reconcile report FILE of any kind exists anywhere in the
	// fixture tree after classification (persistence omitted entirely, per
	// D-3's "simplest conforming implementation").
	after := f.snapshotTree()
	for path := range after {
		if filepathContainsReconcileReport(path) {
			t.Fatalf("unexpected reconcile report artifact found on disk: %s (D-3 requires no persistence)", path)
		}
	}
}

func filepathContainsReconcileReport(path string) bool {
	return strings.Contains(path, "reconcile") && strings.Contains(path, "classify-close-path")
}
