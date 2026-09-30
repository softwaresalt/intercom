package pysem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReadText(t *testing.T) {
	g := loadGolden(t)
	dir := t.TempDir()
	for i, c := range g.ReadText {
		c := c
		t.Run(fmt.Sprintf("case-%02d-%s", i, c.Name), func(t *testing.T) {
			raw := make([]byte, len(c.RawUTF8Bytes))
			for j, b := range c.RawUTF8Bytes {
				raw[j] = byte(b)
			}
			path := filepath.Join(dir, c.Name)
			if err := os.WriteFile(path, raw, 0o600); err != nil {
				t.Fatalf("write fixture %s: %v", path, err)
			}
			got, err := ReadText(path)
			if c.InvalidUTF8 {
				if !errors.Is(err, ErrInvalidUTF8) {
					t.Fatalf("ReadText(%s) error = %v, want ErrInvalidUTF8", c.Name, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadText(%s) unexpected error: %v", c.Name, err)
			}
			if c.Expect == nil {
				t.Fatalf("golden case %s has nil expect but invalid_utf8 is false", c.Name)
			}
			if got != *c.Expect {
				t.Fatalf("ReadText(%s) = %q, want %q", c.Name, got, *c.Expect)
			}
		})
	}
}

func TestGitText(t *testing.T) {
	g := loadGolden(t)
	for i, c := range g.ReadText {
		c := c
		t.Run(fmt.Sprintf("case-%02d-%s", i, c.Name), func(t *testing.T) {
			raw := make([]byte, len(c.RawUTF8Bytes))
			for j, b := range c.RawUTF8Bytes {
				raw[j] = byte(b)
			}
			got, err := GitText(raw)
			if c.InvalidUTF8 {
				if !errors.Is(err, ErrInvalidUTF8) {
					t.Fatalf("GitText(%s) error = %v, want ErrInvalidUTF8", c.Name, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("GitText(%s) unexpected error: %v", c.Name, err)
			}
			if c.Expect == nil {
				t.Fatalf("golden case %s has nil expect but invalid_utf8 is false", c.Name)
			}
			if got != *c.Expect {
				t.Fatalf("GitText(%s) = %q, want %q", c.Name, got, *c.Expect)
			}
		})
	}
}

func TestPrecededByWordOrDot(t *testing.T) {
	g := loadGolden(t)
	for i, c := range g.PrecededByWordOrDot {
		c := c
		t.Run(fmt.Sprintf("case-%02d", i), func(t *testing.T) {
			got := PrecededByWordOrDot(c.Text, c.Offset)
			if got != c.Expect {
				t.Fatalf("PrecededByWordOrDot(%q, %d) = %v, want %v", c.Text, c.Offset, got, c.Expect)
			}
		})
	}
}

func TestFollowedByWord(t *testing.T) {
	g := loadGolden(t)
	for i, c := range g.FollowedByWord {
		c := c
		t.Run(fmt.Sprintf("case-%02d", i), func(t *testing.T) {
			got := FollowedByWord(c.Text, c.Offset)
			if got != c.Expect {
				t.Fatalf("FollowedByWord(%q, %d) = %v, want %v", c.Text, c.Offset, got, c.Expect)
			}
		})
	}
}
