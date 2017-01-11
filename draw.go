package main

import (
	"github.com/prometheus/common/model"
	chart "github.com/wcharczuk/go-chart"
	"image"
	"log"
	"math"
)

func drawChart(matrix model.Matrix, yAxisTitle string,
	setYRangeTo01 bool) image.Image {

	numValues := len(matrix[0].Values)
	for i := range matrix {
		if len(matrix[i].Values) != numValues {
			log.Fatalf("len(matrix[0]) was %d but len(matrix[%d] is %d",
				numValues, i, len(matrix[i].Values))
		}
	}

	minXValue := math.MaxFloat64
	maxXValue := -math.MaxFloat64
	minYValue := math.MaxFloat64
	maxYValue := -math.MaxFloat64
	serieses := []chart.Series{}
	for _, sampleStream := range matrix {
		xvalues := make([]float64, numValues)
		yvalues := make([]float64, numValues)
		for i, samplePair := range sampleStream.Values {
			xvalue := float64(int64(samplePair.Timestamp) * UNIX_MILLIS_TO_UNIX_NANOS)
			xvalues[i] = xvalue
			if xvalue < minXValue {
				minXValue = xvalue
			}
			if xvalue > maxXValue {
				maxXValue = xvalue
			}

			yvalue := float64(samplePair.Value)
			yvalues[i] = yvalue
			if yvalue < minYValue {
				minYValue = yvalue
			}
			if yvalue > maxYValue {
				maxYValue = yvalue
			}
		}
		series := chart.ContinuousSeries{XValues: xvalues, YValues: yvalues}
		serieses = append(serieses, series)
	}

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
