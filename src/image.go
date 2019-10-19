package main

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"
	"time"
)

func startCompareImage(path string) (bool, string) {

	// Open file
	file, _ := os.Open(path)
	defer file.Close()

	// get the color mean
	pixelsImage, err := getPixels(path)

	if err != nil {
		panic(err)
	}

	width := len(pixelsImage[0])
	height := len(pixelsImage)

	// calculate the mean
	now := time.Now().UnixNano()
	mean := computeMeanSlow(height, width, 0, 0, pixelsImage)
	fmt.Println("mmhhh: ", time.Now().UnixNano()-now)

	hex := fmt.Sprintf("%X%X%X", int(mean.R/255), int(mean.G/255), int(mean.B/255))
	temp, err := strconv.ParseUint(hex, 16, 32)

	if err != nil {
		panic(err)
	}

	finalMean := uint32(temp)
	fmt.Println(finalMean)

	// create groups of means
	var nbGroups uint32

	nbGroups = 10000
	finalMean /= (16777215 / nbGroups)

	var size int32

	// create size value
	if height < 500 && width < 500 {
		size = 0
	} else if height > 1000 && width > 1000 {
		size = 2
	} else {
		size = 1
	}

	// get images
	images := getImages(finalMean, size)

	for i := 0; i < len(images); i++ {
		exists := compareImage(pixelsImage, images[i].Path, width, height)
		if exists {
			return true, images[i].Path
		}
	}

	id := getID("image")

	ext := strings.Split(path, ".")

	extension := ext[len(ext)-1]

	if extension == "jpeg" {
		extension = "jpg"
	}

	newName := fmt.Sprintf("./public/%d.%s", id, extension)

	addExistingImage(Image{Path: newName, Color: finalMean, Size: size, ID: id})

	// if gets here, its false
	return false, newName
}

func compareImage(pixelsImage1 [][]Pixel, path2 string, width1 int, height1 int) bool {

	// Get the pixel arrays of the two images
	pixelsImage2, _ := getPixels(path2)

	width2 := len(pixelsImage2[0])
	height2 := len(pixelsImage2)

	// Check if the dimension is equal or not
	if width1 != width2 && height1 != height2 {
		return false
	}

	// The two images are the same size
	var nbPixelsEquivalent int
	var counter int

	// Size of the square pixel groups of which the mean will be computed (can be tweaked)
	pixelSize := 40

	for i := 0; i < height1-height1%pixelSize; i += pixelSize {
		for j := 0; j < width1-width1%pixelSize; j += pixelSize {
			counter++
			mean1 := computeMean(pixelSize, i, j, pixelsImage1)
			mean2 := computeMean(pixelSize, i, j, pixelsImage2)
			if areTheSamePixels(mean1.R, mean1.G, mean1.B, mean2.R, mean2.G, mean2.B) {
				nbPixelsEquivalent++
			}
		}
	}

	result := float32(nbPixelsEquivalent) / float32(counter) * 10000
	// fmt.Println("The two images have a resemblance of", (result)/100, "%")

	if result > 0.94 {
		return true
	}
	return false
}

// Get the bi-dimensional pixel array
func getPixels(filePath string) ([][]Pixel, error) {
	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		panic(err)
		os.Exit(1)
	}

	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	pixels := make([][]Pixel, height)

	for y := 0; y < height; y++ {
		row := make([]Pixel, width)
		for x := 0; x < width; x++ {

			row[x] = rgbaToPixel(img.At(x, y).RGBA())

		}
		pixels[y] = row
	}

	return pixels, nil
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{r, g, b, a}
}

// Pixel struct example
type Pixel struct {
	R uint32
	G uint32
	B uint32
	A uint32
}

func areTheSamePixels(r1 uint32, g1 uint32, b1 uint32, r2 uint32, g2 uint32, b2 uint32) bool {
	if abs(r1-r2) < 10 && abs(g1-g2) < 10 && abs(b1-b2) < 10 {
		return true
	}
	return false

}

func abs(value uint32) uint32 {
	if value < 0 {
		return -value
	}
	return value
}

func computeMean(side int, x int, y int, image [][]Pixel) Pixel {
	var sumR uint32 = 0
	var sumG uint32 = 0
	var sumB uint32 = 0
	var sumA uint32 = 0

	var count uint32 = 1

	xWidth := x + side
	yHeight := y + side

	// Avoid out of bounds
	if xWidth > len(image) {
		xWidth = len(image) - 1
	}
	if yHeight > len(image[0]) {
		yHeight = len(image[0]) - 1
	}

	for i := x; i < xWidth; i++ {
		for j := 0; j < yHeight; j++ {
			// On fait une pt1 de diagonale et balec
			sumR += image[i][j].R
			sumG += image[i][j].G
			sumB += image[i][j].B
			sumA += image[i][j].A
			count++
		}
	}

	/*for i := x; i < xWidth; i++ {
		// up-left
		sumR += image[x+i][y+i].R
		sumG += image[x+i][y+i].G
		sumB += image[x+i][y+i].B
		sumA += image[x+i][y+i].A

		// down-right
		sumR += image[x + xWidth-i][y + yHeight-i].R
		sumG += image[x + xWidth-i][y + yHeight-i].G
		sumB += image[x + xWidth-i][y + yHeight-i].B
		sumA += image[x + xWidth-i][y + yHeight-i].A

		// UP - RIGHT
		sumR += image[x + i][y + yHeight-i].R
		sumG += image[x + i][y + yHeight-i].G
		sumB += image[x + i][y + yHeight-i].B
		sumA += image[x + i][y + yHeight-i].A

		// DOWN-LEFT
		sumR += image[x + xWidth-i][y + i].R
		sumG += image[x + xWidth-i][y + i].G
		sumB += image[x + xWidth-i][y + i].B
		sumA += image[x + xWidth-i][y + i].A
	}*/

	// for i := x; i < xWidth; i++ {
	// 	for j := y; j < yHeight; j++ {
	// 		sumR += image[i][j].R
	// 		sumG += image[i][j].G
	// 		sumB += image[i][j].B
	// 		sumA += image[i][j].A
	// 	}
	// }

	return Pixel{sumR / count, sumB / count, sumG / count, sumG / count}
}

func computeMeanSlow(width int, height int, x int, y int, image [][]Pixel) Pixel {

	var sumR uint32 = 0
	var sumG uint32 = 0
	var sumB uint32 = 0
	var sumA uint32 = 0

	xWidth := x + width
	yHeight := y + height

	if xWidth > len(image) {
		xWidth = len(image) - 1
	}

	if yHeight > len(image[0]) {
		yHeight = len(image[0]) - 1
	}

	for i := x; i < xWidth; i++ {
		for j := y; j < yHeight; j++ {
			sumR += image[i][j].R
			sumG += image[i][j].G
			sumB += image[i][j].B
			sumA += image[i][j].A
		}
	}

	return Pixel{sumR / uint32(width*height), sumG / uint32(width*height), sumB / uint32(width*height), sumA / uint32(width*height)}
}
