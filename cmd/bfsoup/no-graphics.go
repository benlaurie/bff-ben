//go:build !graphics
// +build !graphics

package main

func graphics(_ *[65536]uint8) {
	// Sleep forever
	select {}
}
