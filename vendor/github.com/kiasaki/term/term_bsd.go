// +build openbsd freebsd netbsd darwin

package term

import (
	"syscall"
)

const (
	termGetTermios = syscall.TIOCGETA
	termSetTermios = syscall.TIOCSETA
)
