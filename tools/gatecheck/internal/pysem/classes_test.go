package pysem

import (
	"fmt"
	"testing"
)

func runRuneCases(t *testing.T, cases []runeCase, fn func(rune) bool) {
	t.Helper()
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("U+%04X", c.CP), func(t *testing.T) {
			got := fn(runeOf(t, c))
			if got != c.Expect {
				t.Fatalf("input %q (U+%04X): got %v, want %v", c.Input, c.CP, got, c.Expect)
			}
		})
	}
}

func TestIsSpace(t *testing.T) {
	g := loadGolden(t)
	runRuneCases(t, g.IsSpace, IsSpace)
}

func TestIsWord(t *testing.T) {
	g := loadGolden(t)
	runRuneCases(t, g.IsWord, IsWord)
}

func TestIsDigit(t *testing.T) {
	g := loadGolden(t)
	runRuneCases(t, g.IsDigit, IsDigit)
}

func TestIsDecimal(t *testing.T) {
	g := loadGolden(t)
	runRuneCases(t, g.IsDecimal, IsDecimal)
}
