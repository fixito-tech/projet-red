//go:build !windows

package main

// enableANSI : les terminaux Linux et macOS gèrent déjà les couleurs ANSI.
func enableANSI() bool {
	return true
}

// setRawInput : la lecture des touches en direct n'est gérée que sous Windows ;
// ailleurs, la carte se joue avec Entrée.
func setRawInput(on bool) bool {
	return false
}
