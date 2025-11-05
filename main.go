// Go-BMP implements a bmp reader (it's a hobby project and not to be used in production!)
package main

import (
	"log"

	"github.com/anas-shakeel/go-bmp/internal/bmp"
	"github.com/anas-shakeel/go-bmp/internal/filters"
)

func main() {
	bitmap, err := bmp.ReadBitmap("./images/dot.bmp")
	if err != nil {
		log.Fatal(err)
	}
	filters.Invert(bitmap)
	bitmap.PrintBitmap()
}
