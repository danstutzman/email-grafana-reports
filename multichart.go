package main

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/png"
	"log"
	"os"
)

const MARGIN_Y = 10

type MultiChart struct {
	bigImage   *image.Paletted
	nextChartX int
	nextChartY int
}

func NewMultiChart() *MultiChart {
	bigImage := image.NewPaletted(image.Rect(0, 0, 600, 600), palette.Plan9)

	// set background to white
	white := uint8(bigImage.Palette.Index(color.White))
	pix := bigImage.Pix
	for i := range pix {
		pix[i] = white
	}

	return &MultiChart{
		bigImage:   bigImage,
		nextChartX: 0,
		nextChartY: 0,
	}
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
