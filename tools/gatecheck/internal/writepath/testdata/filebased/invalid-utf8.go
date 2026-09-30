package w

import "os"

// cafÿ broken
func F() error {
	return os.WriteFile("a", nil, 0o644)
}
