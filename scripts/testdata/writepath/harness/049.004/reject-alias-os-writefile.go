package p

import fs "os"

func f() {
	_ = fs.WriteFile("x", nil, 0o644)
}
