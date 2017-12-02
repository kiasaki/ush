// +build linux

package main

import "syscall"

const (
	termGetTermios = syscall.TCGETS
	termSetTermios = syscall.TCSETS
)

const (
	// Input flags
	icrnl  = syscall.ICRNL
	inpck  = syscall.INPCK
	istrip = syscall.ISTRIP
	ixon   = syscall.IXON

	// Output flags
	opost = syscall.OPOST

	// Control flags
	cs8 = syscall.CS8

	// Local flags
	icanon = syscall.ICANON
	iecho  = syscall.ECHO
	iexten = syscall.IEXTEN
	isig   = syscall.ISIG
)
