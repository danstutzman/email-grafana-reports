package main

import (
	"image"
	"log"
	"math"

	chart "github.com/wcharczuk/go-chart"
)

func drawChart(points []Point, yAxisTitle string, setYRangeTo01 bool) image.Image {
	minXValue := math.MaxFloat64
	maxXValue := -math.MaxFloat64
	minYValue := math.MaxFloat64
	maxYValue := -math.MaxFloat64
	serieses := []chart.Series{}

	xvalues := make([]float64, len(points))
	yvalues := make([]float64, len(points))
	for i, point := range points {
		xvalue := float64(point.Time.UnixNano())
		xvalues[i] = xvalue
		if xvalue < minXValue {
			minXValue = xvalue
		}
		if xvalue > maxXValue {
			maxXValue = xvalue
		}

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

	if setYRangeTo01 {
		minYValue = 0.0
		maxYValue = 1.0
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
