package stagecontract

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeDetector struct {
	verdicts map[string]Verdict
	reasons  map[string]string
	errs     map[string]error
	seen     []string
}

func (d *fakeDetector) Detect(f Fixture) (Verdict, string, error) {
	d.seen = append(d.seen, f.ID)
	if err := d.errs[f.ID]; err != nil {
		return "", "", err
	}
	return d.verdicts[f.ID], d.reasons[f.ID], nil
}

const twoFixtureManifest = `{"fixtures":[
{"id":"a-ok","file":"a.md","property":"P1","verdict":"conforming","reason":"reason a"},
{"id":"b-bad","file":"b.md","property":"P2","verdict":"violating","reason":"reason b"}]}`

func writeSuite(t *testing.T, manifestBody string) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range []string{"a.md", "b.md"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("content "+f), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	p := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(p, []byte(manifestBody), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func matchingDetector() *fakeDetector {
	return &fakeDetector{
		verdicts: map[string]Verdict{"a-ok": Conforming, "b-bad": Violating},
		reasons:  map[string]string{"a-ok": "reason a", "b-bad": "reason b"},
	}
}

func TestRunSuiteOneCasePerFixture(t *testing.T) {
	d := matchingDetector()
	res, err := RunSuite(writeSuite(t, twoFixtureManifest), d)
	if err != nil {
		t.Fatalf("RunSuite: %v", err)
	}
	if len(res) != 2 || len(d.seen) != 2 {
		t.Fatalf("want 2 cases, got %d results, %d detector calls", len(res), len(d.seen))
	}
	for _, r := range res {
		if f := r.Failure(); f != "" {
			t.Errorf("unexpected failure: %s", f)
		}
	}
}

func TestRunSuiteVerdictMismatchReport(t *testing.T) {
	d := matchingDetector()
	d.verdicts["b-bad"] = Conforming
	res, err := RunSuite(writeSuite(t, twoFixtureManifest), d)
	if err != nil {
		t.Fatalf("RunSuite: %v", err)
	}
	var report string
	for _, r := range res {
		if r.ID == "b-bad" {
			report = r.Failure()
		}
	}
	for _, want := range []string{"b-bad", "violating", "conforming"} {
		if !strings.Contains(report, want) {
			t.Errorf("report %q missing %q", report, want)
		}
	}
}

func TestRunSuiteReasonMismatchFails(t *testing.T) {
	d := matchingDetector()
	d.reasons["a-ok"] = "different reason"
	res, err := RunSuite(writeSuite(t, twoFixtureManifest), d)
	if err != nil {
		t.Fatalf("RunSuite: %v", err)
	}
	if res[0].Failure() == "" {
		t.Fatal("reason mismatch must fail the case")
	}
}

func TestRunSuiteDetectorErrorRecorded(t *testing.T) {
	d := matchingDetector()
	d.errs = map[string]error{"a-ok": ErrNotImplemented}
	res, err := RunSuite(writeSuite(t, twoFixtureManifest), d)
	if err != nil {
		t.Fatalf("RunSuite: %v", err)
	}
	if !errors.Is(res[0].Err, ErrNotImplemented) {
		t.Fatalf("want ErrNotImplemented recorded, got %v", res[0].Err)
	}
	if res[0].Failure() == "" {
		t.Fatal("detector error must fail the case")
	}
}

func TestRunSuiteBadManifestFails(t *testing.T) {
	cases := map[string]string{
		"malformed":       `{"fixtures": [ not json`,
		"truncated":       twoFixtureManifest[:60],
		"empty":           ``,
		"empty-corpus":    `{"fixtures":[]}`,
		"missing-fixture": `{"fixtures":[{"id":"x","file":"nope.md","property":"P1","verdict":"conforming","reason":"r"}]}`,
		"bad-verdict":     `{"fixtures":[{"id":"x","file":"a.md","property":"P1","verdict":"maybe","reason":"r"}]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := RunSuite(writeSuite(t, body), matchingDetector()); err == nil {
				t.Fatal("bad manifest must fail, got nil error")
			}
		})
	}
}

func TestRunSuiteUnreadableManifestFails(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "manifest.json")
	if _, err := RunSuite(missing, matchingDetector()); err == nil {
		t.Fatal("unreadable manifest must fail, got nil error")
	}
}

func TestStubDetectorReturnsErrNotImplemented(t *testing.T) {
	_, _, err := StubDetector{}.Detect(Fixture{ID: "x"})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("want ErrNotImplemented, got %v", err)
	}
}

func TestFixtureCorpus(t *testing.T) {
	res, err := RunSuite(filepath.Join("testdata", "manifest.json"), StubDetector{})
	if err != nil {
		t.Fatalf("RunSuite: %v", err)
	}
	for _, r := range res {
		if errors.Is(r.Err, ErrNotImplemented) {
			t.Skip("detector is still the ErrNotImplemented stub; corpus enforcement is enabled by 020.006-T")
		}
	}
	for _, r := range res {
		if f := r.Failure(); f != "" {
			t.Error(f)
		}
	}
}
