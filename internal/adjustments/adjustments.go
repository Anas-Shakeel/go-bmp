// Adjusts image dimensions, orientation, or structure.
package adjustments

import (
	"errors"

	"github.com/anas-shakeel/go-bmp/internal/bmp"
)

// Crops a region in the bitmap image (0,0  is at the top-left of the image)
func Crop(b *bmp.BitmapImage, x, y, width, height int) (*bmp.BitmapImage, error) {
	// Validate bounds
	if width+x > int(b.BIHeader.Width) {
		return nil, errors.New("invalid bounds: width out of bounds")
	} else if height+y > int(b.BIHeader.Height) {
		return nil, errors.New("invalid bounds: height out of bounds")
	}

	// Copy the old bitmap (everything except pixels)
	dupBitmap := bmp.BitmapImage{
		BFHeader: b.BFHeader,
		BIHeader: b.BIHeader,
		Filename: b.Filename,
		Stride:   b.Stride,
		Padding:  b.Padding,
	}

	// Crop the bitmap
	dupBitmap.Pixels = make([][]bmp.Pixel, height)
	for row := range height { // Height | Rows
		dupBitmap.Pixels[row] = make([]bmp.Pixel, width)

		for col := range width { // Width | Columns
			dupBitmap.Pixels[row][col] = b.Pixels[row+y][col+x]
		}
	}

	// Update Metadata of dupBitmap
	dupBitmap.UpdateMeta()

	return &dupBitmap, nil
}
