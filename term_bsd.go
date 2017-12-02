// +build openbsd freebsd netbsd darwin

package main

import (
	"syscall"
)

const (
	termGetTermios = syscall.TIOCGETA
	termSetTermios = syscall.TIOCSETA
)
