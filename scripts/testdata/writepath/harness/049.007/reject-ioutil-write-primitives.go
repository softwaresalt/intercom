package p

import (
	"io/ioutil"

	. "golang.org/x/sys/unix"
)

// Keep the fixture calls on their expected physical lines.
func ioutilPrimitives() {
	ioutil.WriteFile("file", nil, 0o644)
	ioutil.TempFile("", "pattern")
	ioutil.TempDir("", "pattern")
}
