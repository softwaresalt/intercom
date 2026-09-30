package pysem

import (
	"fmt"
	"testing"
)

func TestIsUpper(t *testing.T) {
	g := loadGolden(t)
	runRuneCases(t, g.IsUpper, IsUpper)
}

func TestIsLower(t *testing.T) {
	g := loadGolden(t)
	runRuneCases(t, g.IsLower, IsLower)
}

func TestLower(t *testing.T) {
	g := loadGolden(t)
	for i, c := range g.Lower {
		c := c
		t.Run(fmt.Sprintf("case-%02d", i), func(t *testing.T) {
			got := Lower(c.Input)
			if got != c.Expect {
				t.Fatalf("Lower(%q) = %q, want %q", c.Input, got, c.Expect)
			}
		})
	}
}
