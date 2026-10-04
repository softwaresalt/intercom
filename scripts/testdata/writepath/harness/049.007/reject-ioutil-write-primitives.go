package p

import (
	. "golang.org/x/sys/unix"
	"io/ioutil"
)

func ioutilPrimitives() {
	ioutil.WriteFile("file", nil, 0o644)
	ioutil.TempFile("", "pattern")
	ioutil.TempDir("", "pattern")
}
