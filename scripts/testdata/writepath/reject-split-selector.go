package p

import "os"

func f() {
	os.
		WriteFile("x", nil, 0o644)
	os. /* split after the dot */ WriteFile("y", nil, 0o644)
}
