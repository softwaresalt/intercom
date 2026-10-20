package writepath

import "testing"

// 070.001-T: the first production pathsafe.Root.Resolve caller (the process
// workspace-root resolution) fired the 049.005-T tripwire. The register's
// RETIREMENT PROCEDURE was followed in the same change (see internal/pathsafe
// root.go, "070.001-T" entry), and the reviewed call site is recorded in
// reviewedResolveCallers. Every other caller must still fire.

const reviewedResolveSource = `package app
import "github.com/softwaresalt/intercom-go/internal/pathsafe"
func resolve(dir, rel string) (string, error) {
	base, err := pathsafe.NewRoot(dir)
	if err != nil {
		return "", err
	}
	return base.Resolve(rel)
}
`

func TestReviewedResolveCaller_IsNotAFinding(t *testing.T) {
	got, err := scanSource("internal/app/workspace.go", reviewedResolveSource)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("the reviewed Root.Resolve call site produced findings: %q", got)
	}
}

// TestReviewedResolveCaller_AllowanceIsExactlyOneCall pins the reviewed count:
// internal/app/workspace.go was reviewed with ONE Root.Resolve call, so a
// second call there fires the gate and needs its own review.
func TestReviewedResolveCaller_AllowanceIsExactlyOneCall(t *testing.T) {
	const twoCalls = `package app
import "github.com/softwaresalt/intercom-go/internal/pathsafe"
func resolve(dir, rel string) (string, error) {
	base, err := pathsafe.NewRoot(dir)
	if err != nil {
		return "", err
	}
	if _, err := base.Resolve(rel); err != nil {
		return "", err
	}
	return base.Resolve(rel)
}
`
	got, err := scanSource("internal/app/workspace.go", twoCalls)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("two Root.Resolve calls in the reviewed file produced %d findings: %q; want 1 (the second call)", len(got), got)
	}
}

func TestReviewedResolveCaller_AnyOtherFileStillFires(t *testing.T) {
	for _, rel := range []string{
		"internal/app/other.go",
		"internal/app/sub/workspace.go",
		"cmd/intercom/workspace.go",
		"internal/app/workspace.go.bak.go",
	} {
		got, err := scanSource(rel, reviewedResolveSource)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("%s: %d findings for an unreviewed Root.Resolve caller: %q; want 1", rel, len(got), got)
		}
	}
}

func TestReviewedResolveCaller_DoesNotExemptWritePrimitives(t *testing.T) {
	const src = `package app
import "os"
func write() { _ = os.WriteFile("x", nil, 0o600) }
`
	got, err := scanSource("internal/app/workspace.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("a write primitive in the reviewed file produced %d findings: %q; want 1 (the allowance covers Root.Resolve only)", len(got), got)
	}
}
