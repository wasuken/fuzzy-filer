package main

import (
	"syscall"
	"unsafe"
)

// SelectedPath は選択されたパス情報
type SelectedPath struct {
	Path string
}

// InjectToShell はパスをシェルの入力バッファに注入
func InjectToShell(text string) error {
	fd := int(syscall.Stdin)

	// TIOCSTI: Simulate Terminal Input
	// 各文字をターミナルに注入する
	for _, ch := range text {
		b := byte(ch)
		_, _, errno := syscall.Syscall(
			syscall.SYS_IOCTL,
			uintptr(fd),
			uintptr(syscall.TIOCSTI),
			uintptr(unsafe.Pointer(&b)),
		)
		if errno != 0 {
			return errno
		}
	}

	return nil
}
