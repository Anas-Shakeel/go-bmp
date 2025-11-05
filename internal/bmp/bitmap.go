// bmp package implements a bitmap reader (Don't use in production!)
package bmp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/anas-shakeel/go-bmp/internal/utils"
)

type Pixel struct {
	B, G, R byte
}

type BitmapImage struct {
	Filename string
	BFHeader *BitmapFileHeader
	BIHeader *BitmapInfoHeader
	Stride   int
	Padding  int
	Pixels   [][]Pixel
}

// Returns the Pixels in bytes as BGR (Blue, Green, Red)
func (p *Pixel) BytesBGR() []byte {
	return []byte{p.B, p.G, p.R}
}

// Creates and returns a bitmap image (24 bit uncompressed)
func CreateBitmap(width, height int) (*BitmapImage, error) {
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
		Stride:   stride,
		Padding:  stride - width*3, // 3 = Bytes per pixel
		BFHeader: &bfh,
		BIHeader: &bih,
		Pixels:   pixels,
	}, nil

}

// Reads a Bitmap file
func ReadBitmap(filename string) (*BitmapImage, error) {
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
	return &BitmapImage{Filename: filename,
		BFHeader: &bfHeader,
		BIHeader: &biHeader,
		Stride:   stride,
		Padding:  padding,
		Pixels:   pixels,
	}, nil

}

// Saves the bitmap image onto local disk
func (b *BitmapImage) Save(filename string) error {
	newBitmap, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newBitmap.Close()

	width := int(b.BIHeader.Width)
	height := int(b.BIHeader.Height)
	paddingBytes := make([]byte, b.Padding)

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
			_, err := w.Write(b.Pixels[height-row-1][col].BytesBGR())
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

// Returns a Copy of the bitmap image
func (b *BitmapImage) Copy() *BitmapImage {
	newBitmap := BitmapImage{
		Filename: b.Filename,
		Stride:   b.Stride,
		Padding:  b.Padding,
		BFHeader: b.BFHeader,
		BIHeader: b.BIHeader,
	}

	width := b.BIHeader.Width
	height := b.BIHeader.Height

	// Copy over pixels too
	newBitmap.Pixels = make([][]Pixel, height)
	for row := range height {
		newBitmap.Pixels[row] = make([]Pixel, width)
		copy(newBitmap.Pixels[row], b.Pixels[row])
	}

	return &newBitmap

}

// Updates the bitmap metadata (based on pixels)
func (b *BitmapImage) UpdateMeta() {
	// Gather the info
	width := len(b.Pixels[0])
	height := len(b.Pixels)
	bitsPerPixel := 24
	stride := ((width*bitsPerPixel + 31) / 32) * 4
	sizeImage := uint32(stride * height)

	// Update the Metadata
	b.BFHeader.Size = (14 + 40 + sizeImage) // Size of the bitmap file
	b.BIHeader.Width = int32(width)
	b.BIHeader.Height = int32(height)
	b.BIHeader.SizeImage = sizeImage
	b.Stride = stride
	b.Padding = stride - width*3 // 3 = Bytes per pixel
}

// Returns an image containing a single channel of the source image.
// channel can one of (`red`, `green`, and `blue`)
func (b *BitmapImage) GetChannel(channel string) (*BitmapImage, error) {
	newBitmap := b.Copy()

	// Turn the channels to zero except requested one!
	for row := range b.BIHeader.Height {
		for col := range b.BIHeader.Width {
			switch channel {
			case "red":
				// Red Bitmap
				newBitmap.Pixels[row][col].G = 0
				newBitmap.Pixels[row][col].B = 0

			case "green":
				// Green Bitmap
				newBitmap.Pixels[row][col].R = 0
				newBitmap.Pixels[row][col].B = 0

			case "blue":
				// Blue Bitmap
				newBitmap.Pixels[row][col].R = 0
				newBitmap.Pixels[row][col].G = 0

			default:
				return nil, errors.New("invalid color channel: only red, green, and blue are supported")
			}
		}
	}

	// Save the bitmaps
	return newBitmap, nil
}

// Print the bitmap in terminal. Use for small images only
func (b *BitmapImage) PrintBitmap() {
	for _, row := range b.Pixels {
		for _, pixel := range row {
			fmt.Printf("%s", utils.ColoredBlock("  ", int(pixel.R), int(pixel.G), int(pixel.B)))
		}
		fmt.Printf("\n")
	}
}

// Print the Metadata bitmap in terminal. (in human-readable format)
func (b *BitmapImage) PrintMetadata() {
	fmt.Printf("Filename: \t%v\n", b.Filename)
	fmt.Printf("Filesize: \t%v bytes\n", b.BFHeader.Size)
	fmt.Printf("Width: \t\t%v px\n", b.BIHeader.Width)
	fmt.Printf("Height: \t%v px\n", b.BIHeader.Height)
	fmt.Printf("BitCount: \t%vbits\n", b.BIHeader.BitCount)
	fmt.Printf("PixelOffset: \t%v bytes\n", b.BFHeader.OffBits)
	fmt.Printf("PixelCount: \t%v pixels\n", b.BIHeader.Width*b.BIHeader.Height)
	fmt.Printf("Stride: \t%v bytes\n", b.Stride)
	fmt.Printf("Padding: \t%v bytes\n", b.Padding)
}
