package w

import "os"

func F() error {
	return os.WriteFile("a", nil, 0o644)
}
