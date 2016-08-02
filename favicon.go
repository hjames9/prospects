package common

import (
	"bytes"
	"encoding/binary"
)

const (
	ICO_CONTENT_TYPE = "image/x-icon"
)

type IcoHeader struct {
	Reserved       uint16
	ImageType      uint16
	NumberOfImages uint16
}

func (icoHeader IcoHeader) IsValid() bool {
	return icoHeader.Reserved == 0
}

func (icoHeader IcoHeader) IsIco() bool {
	return icoHeader.ImageType == 1
}

func (icoHeader IcoHeader) IsCursor() bool {
	return icoHeader.ImageType == 2
}

type ImageEntry struct {
	Width          byte
	Height         byte
	NumberOfColors byte
	Reserved       byte
	ColorPlanes    uint16
	BitsPerPixel   uint16
	ImageSize      uint32
	ImageOffset    uint32
}

type BitmapInfoHeader struct {
	HeaderSize                  uint32
	Width                       uint32
	Height                      uint32
	ColorPlanes                 uint16
	BitsPerPixel                uint16
	CompressionMethod           uint32
	ImageSize                   uint32
	HorizontalResolution        uint32
	VerticalResolution          uint32
	NumberOfColors              uint32
	NumberOfImportantColorsUsed uint32
}

func (bitmapInfoHeader BitmapInfoHeader) IsCompressed() bool {
	return bitmapInfoHeader.CompressionMethod != 0
}

type PixelData struct {
	Blue  byte
	Green byte
	Red   byte
	Alpha byte
}

type FaviconIco struct {
	Header            IcoHeader
	ImageEntries      []ImageEntry
	BitmapInfoHeaders []BitmapInfoHeader
	Pixels            [][]PixelData
}

func (favicon *FaviconIco) CreateImage(color uint32, width uint32, height uint32) {
	//ICO header
	favicon.Header.Reserved = 0
	favicon.Header.ImageType = 1
	favicon.Header.NumberOfImages = 1

	//ICO parameters
	var imageEntry ImageEntry
	imageEntry.Width = byte(width)
	imageEntry.Height = byte(height)
	imageEntry.NumberOfColors = 0
	imageEntry.Reserved = 0
	imageEntry.ColorPlanes = 1
	imageEntry.BitsPerPixel = 32
	imageEntry.ImageSize = (width * height * uint32(imageEntry.BitsPerPixel/8)) + 40
	imageEntry.ImageOffset = 22 //Size of two headers

	favicon.ImageEntries = append(favicon.ImageEntries, imageEntry)

	//Bitmap info header
	var bitmapInfoHeader BitmapInfoHeader
	bitmapInfoHeader.HeaderSize = 40
	bitmapInfoHeader.Width = width
	bitmapInfoHeader.Height = height * 2
	bitmapInfoHeader.ColorPlanes = 1
	bitmapInfoHeader.BitsPerPixel = 32
	bitmapInfoHeader.CompressionMethod = 0
	bitmapInfoHeader.ImageSize = (width * height * uint32(imageEntry.BitsPerPixel/8))
	bitmapInfoHeader.HorizontalResolution = 2834
	bitmapInfoHeader.VerticalResolution = 2834
	bitmapInfoHeader.NumberOfColors = 0
	bitmapInfoHeader.NumberOfImportantColorsUsed = 0

	favicon.BitmapInfoHeaders = append(favicon.BitmapInfoHeaders, bitmapInfoHeader)

	//Pixels
	numOfPixels := bitmapInfoHeader.ImageSize / uint32(bitmapInfoHeader.BitsPerPixel/8)
	var pixels []PixelData
	for iter := uint32(0); iter < numOfPixels; iter++ {
		var pixel PixelData
		pixel.Red = byte((color >> 24) & 0xff)
		pixel.Green = byte((color >> 16) & 0xff)
		pixel.Blue = byte((color >> 8) & 0xff)
		pixel.Alpha = byte(color & 0xff)
		pixels = append(pixels, pixel)
	}

	favicon.Pixels = append(favicon.Pixels, pixels)
}

func (favicon FaviconIco) GetImageData() []byte {
	var buffer bytes.Buffer

	binary.Write(&buffer, binary.LittleEndian, favicon.Header)

	for iter := uint16(0); iter < favicon.Header.NumberOfImages; iter++ {
		binary.Write(&buffer, binary.LittleEndian, favicon.ImageEntries[iter])
		binary.Write(&buffer, binary.LittleEndian, favicon.BitmapInfoHeaders[iter])
		binary.Write(&buffer, binary.LittleEndian, favicon.Pixels[iter])
	}

	return buffer.Bytes()
}
