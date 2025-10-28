// Go-BMP implements a bmp reader (it's a hobby project and not to be used in production!)
package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
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
	bitmap, err := readBitmap("./images/bmp_24.bmp")
	if err != nil {
		log.Fatal(err)
	}

	bitmap.printBitmap()
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

// Print the bitmap in terminal. Use for small images only
func (b *BitmapImage) printBitmap() {
	for _, row := range b.pixels {
		for _, pixel := range row {
			fmt.Printf("%s", coloredBlock("  ", int(pixel.R), int(pixel.G), int(pixel.B)))
		}
		fmt.Printf("\n")
	}
}

// Print a Colored Block in terminal
func coloredBlock(block string, red int, green int, blue int) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm%s\033[0m", red, green, blue, block)
}

// Returns the Pixels in bytes as BGR (Blue, Green, Red)
func (p *Pixel) bytesBGR() []byte {
	return []byte{p.B, p.G, p.R}
}
