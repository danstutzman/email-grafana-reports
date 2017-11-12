package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/fogleman/gg"
)

const MARGIN_Y = 10
const STICK_TO_LEFT = 0.0
const STICK_TO_TOP = 1.0

type MultiChart struct {
	bigImage   *image.RGBA
	nextChartX int
	nextChartY int
}

func NewMultiChart() *MultiChart {
	bigImage := image.NewRGBA(image.Rect(0, 0, 600, 600))

	// set background to white
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(bigImage, bigImage.Bounds(), &image.Uniform{white},
		image.ZP, draw.Src)

	return &MultiChart{
		bigImage:   bigImage,
		nextChartX: 0,
		nextChartY: 0,
	}
}

func (multichart *MultiChart) WriteHeader(headerText string) {
	context := gg.NewContextForRGBA(multichart.bigImage)
	context.SetRGB(0, 0, 0)
	if err := context.LoadFontFace("/Library/Fonts/Arial.ttf", 30); err != nil {
		panic(err)
	}
	context.DrawStringAnchored(headerText,
		0, float64(multichart.nextChartY), STICK_TO_LEFT, STICK_TO_TOP)

	multichart.nextChartY += 30
}

func (multichart *MultiChart) CopyChart(chartImage image.Image) {
	// Draw chartImage starting at Point{nextChartX,nextChartY}
	draw.Draw(multichart.bigImage, chartImage.Bounds().Add(
		image.Point{multichart.nextChartX, multichart.nextChartY}), chartImage,
		chartImage.Bounds().Min, draw.Src)

	// Increase Y by height of chart and margin
	multichart.nextChartY += chartImage.Bounds().Max.Y + MARGIN_Y
}

func (multichart *MultiChart) SaveToPng(path string) {
	outfile, err := os.Create(path)
	if err != nil {
		log.Fatalf("Error from os.Create('%s'): %s", path, err)
	}
	defer outfile.Close()
	err = png.Encode(outfile, multichart.bigImage)
	if err != nil {
		log.Fatalf("Error from png.Encode: %s", err)
	}
}
