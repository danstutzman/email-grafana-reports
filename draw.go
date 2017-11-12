package main

import (
	"image"
	"log"
	"math"
	"strconv"
	"time"

	chart "github.com/wcharczuk/go-chart"
)

func drawChart(points [][]Point, yAxisTitle string,
	xMin, xMax time.Time,
	yMin, yMax string) image.Image {

	minXValue := float64(xMin.UnixNano())
	maxXValue := float64(xMax.UnixNano())
	minYValue := math.MaxFloat64
	maxYValue := -math.MaxFloat64
	serieses := []chart.Series{}

	for _, seriesPoints := range points {
		xvalues := make([]float64, len(seriesPoints))
		yvalues := make([]float64, len(seriesPoints))
		for i, point := range seriesPoints {
			xvalue := float64(point.Time.UnixNano())
			xvalues[i] = xvalue

			yvalue := point.Value
			yvalues[i] = yvalue
			if yvalue < minYValue {
				minYValue = yvalue
			}
			if yvalue > maxYValue {
				maxYValue = yvalue
			}
		}
		if minYValue == maxYValue {
			minYValue = 0
		}
		series := chart.ContinuousSeries{XValues: xvalues, YValues: yvalues}
		serieses = append(serieses, series)
	}

	if yMin != "" {
		var err error
		minYValue, err = strconv.ParseFloat(yMin, 64)
		if err != nil {
			log.Fatalf("Error from ParseFloat for yMin '%s'", yMin)
		}
	}

	if yMax != "" {
		var err error
		maxYValue, err = strconv.ParseFloat(yMax, 64)
		if err != nil {
			log.Fatalf("Error from ParseFloat for yMax '%s'", yMax)
		}
	}

	graph := chart.Chart{
		Title:      yAxisTitle,
		TitleStyle: chart.StyleShow(),
		Width:      300,
		Height:     200,
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			Range:          &chart.ContinuousRange{Min: minXValue, Max: maxXValue},
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:      yAxisTitle,
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			Range:     &chart.ContinuousRange{Min: minYValue, Max: maxYValue},
		},
		Series: serieses,
	}

	imageWriter := &chart.ImageWriter{}
	err := graph.Render(chart.PNG, imageWriter)
	if err != nil {
		log.Fatalf("Error from graph.Render: %s", err)
	}

	chartImage, err := imageWriter.Image()
	if err != nil {
		log.Fatalf("Error from imageWriter.Image(): %s", err)
	}
	return chartImage
}
