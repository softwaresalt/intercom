package p

import "github.com/softwaresalt/intercom-go/internal/pathsafe"

func shortDeclaration() error {
	root, err := pathsafe.NewRoot(".")
	if err != nil {
		return err
	}
	_, err = root.Resolve("short")
	return err
}

func plainAssignment() error {
	var root pathsafe.Root
	var err error
	root, err = pathsafe.NewRoot(".")
	if err != nil {
		return err
	}
	_, err = root.Resolve("assigned")
	return err
}

func valueSpecification() error {
	var r, err = pathsafe.NewRoot(".")
	if err != nil {
		return err
	}
	_, err = r.Resolve("value-spec")
	return err
}
