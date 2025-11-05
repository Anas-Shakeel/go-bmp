// BMP-specific structs and types
package bmp

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
