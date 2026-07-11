package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderImageFile reads a JPG or PNG file, resizes it, and converts it to ANSI Unicode half blocks
func RenderImageFile(path string, maxWidth int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}

	resized := resizeImage(img, maxWidth)
	return convertToANSI(resized), nil
}

func resizeImage(img image.Image, maxWidth int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= maxWidth {
		return img
	}

	aspect := float64(h) / float64(w)
	newW := maxWidth
	// Each half-block character represents two pixels vertically, so we scale height by 0.5
	newH := int(float64(newW) * aspect * 0.5)
	if newH < 1 {
		newH = 1
	}

	// Simple nearest-neighbor scaling
	resized := image.NewRGBA(image.Rect(0, 0, newW, newH*2))
	for y := 0; y < newH*2; y++ {
		for x := 0; x < newW; x++ {
			srcX := x * w / newW
			srcY := y * h / (newH * 2)
			resized.Set(x, y, img.At(srcX, srcY))
		}
	}
	return resized
}

func convertToANSI(img image.Image) string {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	var sb strings.Builder
	for y := 0; y < h; y += 2 {
		for x := 0; x < w; x++ {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r1Val := uint8(r1 >> 8)
			g1Val := uint8(g1 >> 8)
			b1Val := uint8(b1 >> 8)

			var r2Val, g2Val, b2Val uint8
			if y+1 < h {
				r2, g2, b2, _ := img.At(x, y+1).RGBA()
				r2Val = uint8(r2 >> 8)
				g2Val = uint8(g2 >> 8)
				b2Val = uint8(b2 >> 8)
			}

			fgColor := fmt.Sprintf("#%02x%02x%02x", r2Val, g2Val, b2Val)
			bgColor := fmt.Sprintf("#%02x%02x%02x", r1Val, g1Val, b1Val)

			// Half block character: ▄
			// Foreground colors the bottom half of the block
			// Background colors the top half of the block
			char := lipgloss.NewStyle().
				Foreground(lipgloss.Color(fgColor)).
				Background(lipgloss.Color(bgColor)).
				Render("▄")

			sb.WriteString(char)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
