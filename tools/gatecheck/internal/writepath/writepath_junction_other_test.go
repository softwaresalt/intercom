//go:build !windows

package writepath

import "errors"

// nativeJunction exists so the shared helpers compile everywhere; linkDir only
// calls it on Windows, where junctions need no privilege.
func nativeJunction(target, link string) error {
	return errors.New("junctions are only supported on Windows")
}
