//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package retention

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func lock(path string) (func() error, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if !info.Mode().IsRegular() {
		return nil, errors.Join(fmt.Errorf("retention lock is not a regular file"), file.Close())
	}
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return nil, errors.Join(err, file.Close())
	}
	return file.Close, nil
}
