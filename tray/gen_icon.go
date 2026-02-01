//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

func main() {
	// Create a 22x22 image with a simple printer icon shape
	img := image.NewRGBA(image.Rect(0, 0, 22, 22))

	// Fill with transparent
	for y := 0; y < 22; y++ {
		for x := 0; x < 22; x++ {
			img.Set(x, y, color.Transparent)
		}
	}

	// Draw a simple printer shape (black)
	black := color.RGBA{0, 0, 0, 255}

	// Paper coming out (top)
	for y := 2; y < 7; y++ {
		for x := 6; x < 16; x++ {
			img.Set(x, y, black)
		}
	}

	// Printer body (middle)
	for y := 7; y < 15; y++ {
		for x := 3; x < 19; x++ {
			img.Set(x, y, black)
		}
	}

	// Paper tray (bottom)
	for y := 15; y < 20; y++ {
		for x := 5; x < 17; x++ {
			img.Set(x, y, black)
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)

	// Output as Go code
	fmt.Println("package tray")
	fmt.Println()
	fmt.Println("// Icon is a 22x22 printer icon for the system tray")
	fmt.Println("var Icon = []byte{")

	data := buf.Bytes()
	for i, b := range data {
		if i%12 == 0 {
			fmt.Print("\t")
		}
		fmt.Printf("0x%02x, ", b)
		if i%12 == 11 {
			fmt.Println()
		}
	}
	fmt.Println()
	fmt.Println("}")
}
