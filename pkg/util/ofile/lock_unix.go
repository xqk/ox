// +build darwin dragonfly freebsd linux netbsd openbsd

package ofile

import (
	"io"
	"os"
	"syscall"
)

// lockCloser hides all of an os.File's methods, except for Close.
type lockCloser struct {
	f *os.File
}

// Close ...
func (l lockCloser) Close() error {
	return l.f.Close()
}

// Lock ...
func Lock(name string) (io.Closer, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	/*
		Some people tell me FcntlFlock does not exist, so use flock here
	*/
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, err
	}

	// spec := syscall.Flock_t{
	// 	Type:   syscall.F_WRLCK,
	// 	Whence: int16(os.SEEK_SET),
	// 	Start:  0,
	// 	Len:    0, // 0 means to lock the entire file.
	// 	Pid:    int32(os.Getpid()),
	// }
	// if err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, &spec); err != nil {
	// 	f.Close()
	// 	return nil, err
	// }

	return lockCloser{f}, nil
}
