// +build linux

package term

import "syscall"

const (
	termGetTermios = syscall.TCGETS
	termSetTermios = syscall.TCSETS
)
