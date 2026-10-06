//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package retention

import "fmt"

func lock(string) (func() error, error) {
	return nil, fmt.Errorf("retention policy writes require a supported Unix host")
}
