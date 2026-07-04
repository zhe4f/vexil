//go:build windows

package file

import (
	"os"
	"syscall"
	"unsafe"
)

func OpenFile(path string) (*os.File, error) {
    return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
}

func Preallocate(f *os.File, size int64) error {
	// 1. 设为稀疏文件
	var bytesReturned uint32
	err := syscall.DeviceIoControl(
		syscall.Handle(f.Fd()),
		0x000900C4, // FSCTL_SET_SPARSE
		nil, 0, nil, 0,
		&bytesReturned, nil,
	)
	if err != nil {
		// 失败则回退到 Truncate（非稀疏但能保证空间）
		return f.Truncate(size)
	}

	// 2. 设置文件大小（稀疏，不实际写零）
	err = syscall.DeviceIoControl(
		syscall.Handle(f.Fd()),
		0x000980E8, // FSCTL_SET_ZERO_DATA
		(*byte)(unsafe.Pointer(&size)),
		uint32(unsafe.Sizeof(size)),
		nil, 0,
		&bytesReturned, nil,
	)
	if err != nil {
		return f.Truncate(size)
	}

	// 3. 设置文件指针到末尾
	return f.Truncate(size)
}