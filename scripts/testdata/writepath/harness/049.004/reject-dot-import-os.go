package p

import . "os"

func f() {
	_ = WriteFile("x", nil, 0o644)
}
