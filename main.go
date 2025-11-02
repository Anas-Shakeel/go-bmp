// Go-BMP implements a bmp reader (it's a hobby project and not to be used in production!)
package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

// The BitmapFileHeader structure contains information about the type, size,
// and layout of a file that contains a DIB [device-independent bitmap].
// https://learn.microsoft.com/en-us/windows/win32/api/wingdi/ns-wingdi-bitmapfileheader

type BitmapFileHeader struct {
	Type      [2]byte // The file type: must be 0x4d42 (ASCII string "BM").
	Size      uint32  // The size, in bytes, of the bitmap file.
	Reserved1 uint16  // Reserved; must be zero.
	Reserved2 uint16  // Reserved; must be zero.
	OffBits   uint32  // Bitmap File Offset (In bytes) to Pixel Arrays
}

// The BitmapInfoHeader structure contains information about the
// dimensions and color format of DIB [device-independent bitmap].

type BitmapInfoHeader struct {
	Size            uint32 // The number of bytes required by the structure.
	Width           int32  // The width of the bitmap, in pixels.
	Height          int32  // The height of the bitmap, in pixels
	Planes          uint16 // The number of planes for the target device.
	BitCount        uint16 // The number of bits-per-pixel.
	Compression     uint32 // The type of compression
	SizeImage       uint32 // The size of the image (in bytes).
	XPixelsPerM     int32  // The horizontal resolution, in pixels-per-meter.
	YPixelsPerM     int32  // The vertical resolution, in pixels-per-meter.
	ColorsUsed      uint32 // Number of color indexes that are actually used by bitmap.
	ColorsImportant uint32 // Number of color indexes required for displaying the bitmap.
}

type Pixel struct {
	B, G, R byte
}

type BitmapImage struct {
	filename string
	BFHeader *BitmapFileHeader
	BIHeader *BitmapInfoHeader
	stride   int
	padding  int
	pixels   [][]Pixel
}

func main() {
	fmt.Println("Bitmap reader from scratch in golang!")
}

// Print a Colored Block in terminal
func coloredBlock(block string, red int, green int, blue int) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm%s\033[0m", red, green, blue, block)
}

// Returns the Pixels in bytes as BGR (Blue, Green, Red)
func (p *Pixel) bytesBGR() []byte {
	return []byte{p.B, p.G, p.R}
}

// Returns the average of all given numbers n
func average(n ...int) int {
	// Sum all numbers
	var sum int
	for _, num := range n {
		sum += num
	}

	// Divide sum by total numbers
	return sum / len(n)
}

// Creates and returns a bitmap image (24 bit uncompressed)
func createBitmap(width, height int) (*BitmapImage, error) {
	if width <= 0 {
		return nil, errors.New("width must be greater than 0")
	} else if height <= 0 {
		return nil, errors.New("height must be greater than 0")
	}

	bitsPerPixel := 24
	stride := ((width*bitsPerPixel + 31) / 32) * 4
	biSizeImage := uint32(stride * height)
	fileSize := (14 + 40 + biSizeImage) // Size of the whole bitmap file

	// NewBitmap Headers
	bfh := BitmapFileHeader{Type: [2]byte{0x42, 0x4d}, OffBits: 54, Size: fileSize}
	bih := BitmapInfoHeader{Size: 40, Width: int32(width), Height: int32(height), Planes: 1, BitCount: 24, SizeImage: biSizeImage}

	// Create the pixels 2d slice
	pixels := make([][]Pixel, height)
	for i := range height {
		pixels[i] = make([]Pixel, width)
	}

	// Create and return the bitmap
	return &BitmapImage{
		stride:   stride,
		padding:  stride - width*3, // 3 = Bytes per pixel
		BFHeader: &bfh,
		BIHeader: &bih,
		pixels:   pixels,
	}, nil

}

// Reads a Bitmap file
func readBitmap(filename string) (*BitmapImage, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read File Header
	var bfHeader BitmapFileHeader
	binary.Read(file, binary.LittleEndian, &bfHeader)

	// TODO: Verify that this is a .BMP file by checking bitmap id (0x424d)
	if bfHeader.Type[0] != 0x42 && bfHeader.Type[1] != 0x4d {
		return nil, errors.New("invalid file: provided file is not a bitmap")
	}

	// READ Info Header OR (more commonly) DIB Header!
	var biHeader BitmapInfoHeader
	binary.Read(file, binary.LittleEndian, &biHeader)

	// Support only 24bit uncompressed Bitmaps (common)
	if biHeader.BitCount != 24 || biHeader.Compression != 0 {
		return nil, errors.New("unsupported BMP format: only 24-bit uncompressed is supported")
	}

	width := int(biHeader.Width)
	height := int(biHeader.Height)
	topDown := false // Pixels are stored TopDown?
	if height < 0 {
		topDown = true
		height = -height // Abs(olute) Height
		biHeader.Height = int32(height)
	}
	bytesPerPixel := biHeader.BitCount / 8

	stride := ((width*int(bytesPerPixel) + 3) / 4) * 4 // Total bytes in a row (incl. padding)
	padding := stride - width*int(bytesPerPixel)       // padding-bytes for each row

	// Initialize a 2D Slice to store pixels
	pixels := make([][]Pixel, height)
	for i := range height {
		pixels[i] = make([]Pixel, width)
	}

	// Seek to Pixel Array (OffBits)
	_, err = file.Seek(int64(bfHeader.OffBits), io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Populate the 2D slice with image pixels
	for i := range height {
		rowIndex := height - i - 1
		if topDown {
			rowIndex = i
		}

		// Read pixels of current row (excluding padding)
		err := binary.Read(file, binary.LittleEndian, pixels[rowIndex])
		if err != nil {
			return nil, err
		}

		// Seek over padding bytes
		_, err = file.Seek(int64(padding), io.SeekCurrent)
		if err != nil {
			return nil, err
		}
	}

	// Return the (pointer to) Bitmap Image
	return &BitmapImage{filename: filename,
		BFHeader: &bfHeader,
		BIHeader: &biHeader,
		stride:   stride,
		padding:  padding,
		pixels:   pixels,
	}, nil

}

// Saves the bitmap image onto local disk
func (b *BitmapImage) save(filename string) error {
	newBitmap, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newBitmap.Close()

	width := int(b.BIHeader.Width)
	height := int(b.BIHeader.Height)
	paddingBytes := make([]byte, b.padding)

	// Create a buffer (to reduce syscalls)
	w := bufio.NewWriter(newBitmap)

	// Write File Header
	err = binary.Write(w, binary.LittleEndian, b.BFHeader)
	if err != nil {
		return err
	}
	// Write Info Header
	err = binary.Write(w, binary.LittleEndian, b.BIHeader)
	if err != nil {
		return err
	}

	// Write the pixels (BottomUp: last row first)
	for row := range height {
		for col := range width {
			_, err := w.Write(b.pixels[height-row-1][col].bytesBGR())
			if err != nil {
				return err
			}
		}
		_, err := w.Write(paddingBytes) // Padding bytes
		if err != nil {
			return err
		}
	}

	w.Flush() // Write buffer to disk

	return nil
}

// Returns a copy of the bitmap image
func (b *BitmapImage) copy() *BitmapImage {
	newBitmap := BitmapImage{
		filename: b.filename,
		stride:   b.stride,
		padding:  b.padding,
		BFHeader: b.BFHeader,
		BIHeader: b.BIHeader,
	}

	width := b.BIHeader.Width
	height := b.BIHeader.Height

	// Copy over pixels too
	newBitmap.pixels = make([][]Pixel, height)
	for row := range height {
		newBitmap.pixels[row] = make([]Pixel, width)
		copy(newBitmap.pixels[row], b.pixels[row])
	}

	return &newBitmap

}

// Updates the bitmap metadata (based on pixels)
func (b *BitmapImage) updateMeta() {
	// Gather the info
	width := len(b.pixels[0])
	height := len(b.pixels)
	bitsPerPixel := 24
	stride := ((width*bitsPerPixel + 31) / 32) * 4
	sizeImage := uint32(stride * height)

	// Update the Metadata
	b.BFHeader.Size = (14 + 40 + sizeImage) // Size of the bitmap file
	b.BIHeader.Width = int32(width)
	b.BIHeader.Height = int32(height)
	b.BIHeader.SizeImage = sizeImage
	b.stride = stride
	b.padding = stride - width*3 // 3 = Bytes per pixel
}

// Crops a region in the bitmap image (0,0  is at the top-left of the image)
func (b *BitmapImage) crop(x, y, width, height int) (*BitmapImage, error) {
	// Validate bounds
	if width+x > int(b.BIHeader.Width) {
		return nil, errors.New("invalid bounds: width out of bounds")
	} else if height+y > int(b.BIHeader.Height) {
		return nil, errors.New("invalid bounds: height out of bounds")
	}

	// Copy the old bitmap (everything except pixels)
	dupBitmap := BitmapImage{
		BFHeader: b.BFHeader,
		BIHeader: b.BIHeader,
		filename: b.filename,
		stride:   b.stride,
		padding:  b.padding,
	}

	// Crop the bitmap
	dupBitmap.pixels = make([][]Pixel, height)
	for row := range height { // Height | Rows
		dupBitmap.pixels[row] = make([]Pixel, width)

		for col := range width { // Width | Columns
			dupBitmap.pixels[row][col] = b.pixels[row+y][col+x]
		}
	}

	// Update Metadata of dupBitmap
	dupBitmap.updateMeta()

	return &dupBitmap, nil
}

// Inverts (negates) the bitmap image
func (b *BitmapImage) invert() {
	width := int(b.BIHeader.Width)
	height := int(b.BIHeader.Height)

	// Iterate rows
	for row := range height {
		// Iterate pixels in row
		for col := range width {
			b.pixels[row][col].R = (255 - b.pixels[row][col].R)
			b.pixels[row][col].G = (255 - b.pixels[row][col].G)
			b.pixels[row][col].B = (255 - b.pixels[row][col].B)
		}
	}
}

// Converts a bitmap to Black-and-White
func (b *BitmapImage) grayscale() {
	width := int(b.BIHeader.Width)
	height := int(b.BIHeader.Height)

	// // Iterate rows
	for row := range height {
		// Iterate pixels in row
		for col := range width {
			// Find the average value for pixel
			pixel := b.pixels[row][col]
			avg := average(int(pixel.R), int(pixel.G), int(pixel.B))

			b.pixels[row][col].R = byte(avg)
			b.pixels[row][col].G = byte(avg)
			b.pixels[row][col].B = byte(avg)
		}
	}

}

// Converts a bitmap to Black-and-White (with ITU-R 601-2 Luma Transform)
func (b *BitmapImage) grayscaleLuma() {
	width := int(b.BIHeader.Width)
	height := int(b.BIHeader.Height)

	// Iterate rows
	for row := range height {
		// Iterate pixels in row
		for col := range width {
			p := b.pixels[row][col]
			L := int(p.R)*299/1000 + int(p.G)*587/1000 + int(p.B)*114/1000

			b.pixels[row][col].R = byte(L)
			b.pixels[row][col].G = byte(L)
			b.pixels[row][col].B = byte(L)
		}
	}

}

// Adjusts image brightness by a factor float.
// An enhancement factor of 0.0 gives a black image.
// A factor of 1.0 gives the original image.
// And a factor of 1.5 gives image 50% brighter!
func (b *BitmapImage) brightness(factor float64) {
	for row := range b.BIHeader.Height {
		for col := range b.BIHeader.Width {
			p := b.pixels[row][col]

			b.pixels[row][col].R = byte(math.Min(math.Max(float64(p.R)*factor, 0), 255))
			b.pixels[row][col].G = byte(math.Min(math.Max(float64(p.G)*factor, 0), 255))
			b.pixels[row][col].B = byte(math.Min(math.Max(float64(p.B)*factor, 0), 255))
		}
	}
}

// Returns an image containing a single channel of the source image.
// channel can one of (`red`, `green`, and `blue`)
func (b *BitmapImage) getChannel(channel string) (*BitmapImage, error) {
	newBitmap := b.copy()

	// Turn the channels to zero except requested one!
	for row := range b.BIHeader.Height {
		for col := range b.BIHeader.Width {
			switch channel {
			case "red":
				// Red Bitmap
				newBitmap.pixels[row][col].G = 0
				newBitmap.pixels[row][col].B = 0

			case "green":
				// Green Bitmap
				newBitmap.pixels[row][col].R = 0
				newBitmap.pixels[row][col].B = 0

			case "blue":
				// Blue Bitmap
				newBitmap.pixels[row][col].R = 0
				newBitmap.pixels[row][col].G = 0

			default:
				return nil, errors.New("invalid color channel: only red, green, and blue are supported")
			}
		}
	}

	// Save the bitmaps
	return newBitmap, nil
}

// Print the bitmap in terminal. Use for small images only
func (b *BitmapImage) printBitmap() {
	for _, row := range b.pixels {
		for _, pixel := range row {
			fmt.Printf("%s", coloredBlock("  ", int(pixel.R), int(pixel.G), int(pixel.B)))
		}
		fmt.Printf("\n")
	}
}
