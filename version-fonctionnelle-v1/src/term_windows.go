//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsole = kernel32.NewProc("GetConsoleMode")
	procSetConsole = kernel32.NewProc("SetConsoleMode")

	savedInputMode uint32
	rawInputActive bool
)

const (
	enableEchoInput       = 0x0004
	enableLineInput       = 0x0002
	enableVirtualTerminal = 0x0004
)

// enableANSI active l'affichage des couleurs dans la console Windows.
// Retourne false si la sortie n'est pas une console (redirection, tests).
func enableANSI() bool {
	handle := uintptr(syscall.Stdout)
	var mode uint32
	if r, _, _ := procGetConsole.Call(handle, uintptr(unsafe.Pointer(&mode))); r == 0 {
		return false
	}
	r, _, _ := procSetConsole.Call(handle, uintptr(mode|enableVirtualTerminal))
	return r != 0
}

// setRawInput active (on = true) ou désactive la lecture des touches en direct,
// sans avoir à appuyer sur Entrée. Retourne false si l'entrée n'est pas une console.
func setRawInput(on bool) bool {
	handle := uintptr(syscall.Stdin)
	if !on {
		if rawInputActive {
			procSetConsole.Call(handle, uintptr(savedInputMode))
			rawInputActive = false
		}
		return true
	}

	var mode uint32
	if r, _, _ := procGetConsole.Call(handle, uintptr(unsafe.Pointer(&mode))); r == 0 {
		return false
	}
	if !rawInputActive {
		savedInputMode = mode
	}
	raw := mode &^ (enableLineInput | enableEchoInput)
	if r, _, _ := procSetConsole.Call(handle, uintptr(raw)); r == 0 {
		return false
	}
	rawInputActive = true
	return true
}
