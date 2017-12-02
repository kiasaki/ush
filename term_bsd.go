// +build openbsd freebsd netbsd darwin

package main

import "syscall"

const (
	termGetTermios = syscall.TIOCGETA
	termSetTermios = syscall.TIOCSETA
)

const (
	// Input flags
	icrnl  = 0x100
	inpck  = 0x010
	istrip = 0x020
	ixon   = 0x200

	// Output flags
	opost = 0x1

	// Control flags
	cs8 = 0x300

	// Local flags
	icanon = 0x100
	iecho  = 0x800
	iexten = 0x400
	isig   = 0x080
)
