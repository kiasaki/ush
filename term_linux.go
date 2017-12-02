// +build linux

package main

import "syscall"

const (
	termGetTermios = syscall.TCGETS
	termSetTermios = syscall.TCSETS
)
