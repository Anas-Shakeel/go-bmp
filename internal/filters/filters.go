// Filters perform color manipulation and per-pixel operations
package filters

import (
	"errors"
	"math"

	"github.com/anas-shakeel/go-bmp/internal/bmp"
	"github.com/anas-shakeel/go-bmp/internal/utils"
)

// Inverts (negates) the bitmap image
func Invert(b *bmp.BitmapImage) {
	width := int(b.BIHeader.Width)
	height := int(b.BIHeader.Height)

	// Iterate rows
	for row := range height {
		// Iterate pixels in row
		for col := range width {
			b.Pixels[row][col].R = (255 - b.Pixels[row][col].R)
			b.Pixels[row][col].G = (255 - b.Pixels[row][col].G)
			b.Pixels[row][col].B = (255 - b.Pixels[row][col].B)
		}
	}
}

// Converts a bitmap to Black-and-White
func Grayscale(b *bmp.BitmapImage) {
	width := int(b.BIHeader.Width)
	height := int(b.BIHeader.Height)

	// // Iterate rows
	for row := range height {
		// Iterate pixels in row
		for col := range width {
			// Find the average value for pixel
			pixel := b.Pixels[row][col]
			avg := utils.Average(int(pixel.R), int(pixel.G), int(pixel.B))

			b.Pixels[row][col].R = byte(avg)
			b.Pixels[row][col].G = byte(avg)
			b.Pixels[row][col].B = byte(avg)
		}
	}

}

// Converts a bitmap to Black-and-White (with ITU-R 601-2 Luma Transform)
func GrayscaleLuma(b *bmp.BitmapImage) {
	width := int(b.BIHeader.Width)
	height := int(b.BIHeader.Height)

	// Iterate rows
	for row := range height {
		// Iterate pixels in row
		for col := range width {
			p := b.Pixels[row][col]
			L := int(p.R)*299/1000 + int(p.G)*587/1000 + int(p.B)*114/1000

			b.Pixels[row][col].R = byte(L)
			b.Pixels[row][col].G = byte(L)
			b.Pixels[row][col].B = byte(L)
		}
	}

}

// Adjusts the Brightness of a Bitmap in-place.
//
// method can be "add" (adds value to each channel) or "multiply" (multiplies each channel by value).
// Pixel values are clipped to [0, 255].
func Brightness(b *bmp.BitmapImage, factor float64, method string) error {
	type Operation func(x, y float64) float64
	var operation Operation

	// Select an operation of brightness (additive or multiplicative)
	switch method {
	case "add":
		operation = func(x, y float64) float64 {
			return x + y
		}
	case "multiply":
		operation = func(x, y float64) float64 {
			return x * y
		}
	default:
		return errors.New("invalid method: method must be add or multiply")
	}

	// Apply brightness (or darkness)
	for row := range b.BIHeader.Height {
		for col := range b.BIHeader.Width {
			p := b.Pixels[row][col]

			b.Pixels[row][col].R = byte(math.Min(math.Max(operation(float64(p.R), factor), 0), 255))
			b.Pixels[row][col].G = byte(math.Min(math.Max(operation(float64(p.G), factor), 0), 255))
			b.Pixels[row][col].B = byte(math.Min(math.Max(operation(float64(p.B), factor), 0), 255))
		}
	}

	return nil
}

// Adjusts the Contrast of a Bitmap in-place.
// factor > 1.0 increases Contrast, factor < 1.0 decreases it.
func Contrast(b *bmp.BitmapImage, factor float64) {
	// Compute mean for each channel
	var sumR, sumG, sumB int
	for row := range b.BIHeader.Height {
		for col := range b.BIHeader.Width {
			sumR += int(b.Pixels[row][col].R)
			sumG += int(b.Pixels[row][col].G)
			sumB += int(b.Pixels[row][col].B)
		}
	}
	totalPixels := int(b.BIHeader.Width * b.BIHeader.Height)
	meanR := float64(sumR / totalPixels) // Average of all R pixels
	meanG := float64(sumG / totalPixels) // Average of all G pixels
	meanB := float64(sumB / totalPixels) // Average of all B pixels

	// Apply contrast
	for row := range b.BIHeader.Height {
		for col := range b.BIHeader.Width {
			p := b.Pixels[row][col]

			b.Pixels[row][col].R = byte(math.Min(math.Max(float64(p.R)*factor+(1-factor)*meanR, 0), 255))
			b.Pixels[row][col].G = byte(math.Min(math.Max(float64(p.G)*factor+(1-factor)*meanG, 0), 255))
			b.Pixels[row][col].B = byte(math.Min(math.Max(float64(p.B)*factor+(1-factor)*meanB, 0), 255))
		}
	}

}
