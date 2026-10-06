package p

import "github.com/softwaresalt/intercom-go/internal/pathsafe"

type unrelated struct{}

func (unrelated) Resolve(string) {}

type pathsafeCarrier struct {
	root pathsafe.Root
}

func f() {
	var receiver unrelated
	receiver.Resolve("unrelated")
}
