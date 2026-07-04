//go:build !windows

package file

import "os"

func OpenFile(path string) (*os.File, error) {
    return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
}

func Preallocate(f *os.File, size int64) error {
	return f.Truncate(size)
}