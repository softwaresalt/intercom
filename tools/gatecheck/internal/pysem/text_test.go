package pysem

import (
	"fmt"
	"slices"
	"testing"
)

func TestStrip(t *testing.T) {
	g := loadGolden(t)
	for i, c := range g.Strip {
		c := c
		t.Run(fmt.Sprintf("case-%02d", i), func(t *testing.T) {
			got := Strip(c.Input)
			if got != c.Expect {
				t.Fatalf("Strip(%q) = %q, want %q", c.Input, got, c.Expect)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	g := loadGolden(t)
	for i, c := range g.SplitLines {
		c := c
		t.Run(fmt.Sprintf("case-%02d", i), func(t *testing.T) {
			got := SplitLines(c.Input)
			if !slices.Equal(got, c.Expect) {
				t.Fatalf("SplitLines(%q) = %#v, want %#v", c.Input, got, c.Expect)
			}
		})
	}
}

func TestRepr(t *testing.T) {
	g := loadGolden(t)
	for i, c := range g.Repr {
		c := c
		t.Run(fmt.Sprintf("case-%02d", i), func(t *testing.T) {
			got := Repr(c.Input)
			if got != c.Expect {
				t.Fatalf("Repr(%q) = %s, want %s", c.Input, got, c.Expect)
			}
		})
	}
}

func TestWordBoundary(t *testing.T) {
	g := loadGolden(t)
	for i, c := range g.WordBoundary {
		c := c
		t.Run(fmt.Sprintf("case-%02d", i), func(t *testing.T) {
			got := WordBoundary(c.Text, c.Offset)
			if got != c.Expect {
				t.Fatalf("WordBoundary(%q, %d) = %v, want %v", c.Text, c.Offset, got, c.Expect)
			}
		})
	}
}
