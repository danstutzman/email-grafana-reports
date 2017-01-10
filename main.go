package main

import (
	"bytes"
	"flag"
	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
	chart "github.com/wcharczuk/go-chart"
	"golang.org/x/net/context"
	"image"
	"image/color/palette"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"regexp"
	"time"
)

const UNIX_MILLIS_TO_UNIX_NANOS = 1000 * 1000
const HOST_AND_PORT_REGEXP = "^[a-z0-9-]+:[0-9]+)$"

type Config struct {
	pngPath            string
	prometheusHostPort string
	emailFrom          string
	emailTo            string
	emailSubject       string
	smtpHostPort       string
	doSendEmail        bool
}

func getConfigFromFlags() Config {
	config := Config{}
	flag.StringVar(&config.pngPath, "pngPath", "", "Path to save .png image to")
	flag.StringVar(&config.prometheusHostPort, "prometheusHostPort", "",
		"Hostname and port for Prometheus server (e.g. localhost:9090)")
	flag.StringVar(&config.emailFrom, "emailFrom", "", "Email address to send report from; e.g. Reports <reports@monitoring.danstutzman.com>")
	flag.StringVar(&config.emailTo, "emailTo", "", "Email address to send report to")
	flag.StringVar(&config.emailSubject, "emailSubject", "", "Subject for email report")
	flag.StringVar(&config.smtpHostPort, "smtpHostPort", "",
		"Hostname and port for SMTP server; e.g. localhost:25")
	flag.Parse()

	if config.pngPath == "" {
		log.Fatalf("You must specify -pngPath; try ./out.png")
	}
	if config.prometheusHostPort == "" {
		log.Fatalf("You must specify -prometheusHostPort; try localhost:9090")
	}
	if matched, _ := regexp.Match(HOST_AND_PORT_REGEXP,
		[]byte(config.prometheusHostPort)); matched {
		log.Fatalf("-prometheusHostPort value must match " + HOST_AND_PORT_REGEXP)
	}
	if config.emailFrom == "" &&
		config.emailTo == "" &&
		config.emailSubject == "" &&
		config.smtpHostPort == "" {
		config.doSendEmail = false
	} else if config.emailFrom != "" &&
		config.emailTo != "" &&
		config.emailSubject != "" &&
		config.smtpHostPort != "" {
		config.doSendEmail = true
	} else {
		log.Fatalf("Please supply values for all of -emailFrom, -emailTo, -emailSubject, and -smtpHostPort or none of them")
	}

	return config
}

func queryCloudfrontVisits(api prometheus.QueryAPI) model.Matrix {
	value, err := api.QueryRange(context.TODO(),
		`cloudfront_visits{site_name="vocabincontext.com",status="200"}`,
		prometheus.Range{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
			Step:  20 * time.Minute,
		},
	)
	if err != nil {
		log.Fatalf("Error from api.QueryRange: %s", err)
	}
	if value.Type() != model.ValMatrix {
		log.Fatalf("Expected value.Type() == ValMatrix but got %d", value.Type())
	}
	return value.(model.Matrix)
}

func writeRgbPngToPalettedPng(buffer *bytes.Buffer, outputPngPath string) {
	src, err := png.Decode(buffer)
	if err != nil {
		log.Fatalf("Error from png.Decode: %s", err)
	}

	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	paletted := image.NewPaletted(bounds, palette.Plan9)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			oldColor := src.At(x, y)
			paletted.Set(x, y, oldColor)
		}
	}

	outfile, err := os.Create(outputPngPath)
	if err != nil {
		log.Fatalf("Error from os.Create('%s'): %s", outputPngPath, err)
	}
	defer outfile.Close()
	encoder := &png.Encoder{CompressionLevel: png.BestCompression}
	err = encoder.Encode(outfile, paletted)
	if err != nil {
		log.Fatalf("Error from png.Encode: %s", err)
	}
}

func draw1SeriesChart(matrix model.Matrix, yAxisTitle string) image.Image {
	if len(matrix) != 1 {
		log.Fatalf("Expected only one series but got len(matrix) == %d", len(matrix))
	}
	numValues := len(matrix[0].Values)

	xvalues := make([]float64, numValues)
	yvalues := make([]float64, numValues)
	minYValue := math.MaxFloat64
	maxYValue := -math.MaxFloat64
	for i, samplePair := range matrix[0].Values {
		xvalue := float64(int64(samplePair.Timestamp) * UNIX_MILLIS_TO_UNIX_NANOS)
		xvalues[i] = xvalue

		yvalue := float64(samplePair.Value)
		yvalues[i] = yvalue
		if yvalue < minYValue {
			minYValue = yvalue
		}
		if yvalue > maxYValue {
			maxYValue = yvalue
		}
	}

	graph := chart.Chart{
		Width:  300,
		Height: 200,
		XAxis: chart.XAxis{
			Style: chart.StyleShow(),
			Range: &chart.ContinuousRange{
				Min: xvalues[0],
				Max: xvalues[numValues-1],
			},
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:      yAxisTitle,
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			Range:     &chart.ContinuousRange{Min: minYValue, Max: maxYValue},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{XValues: xvalues, YValues: yvalues},
		},
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

func saveChartImagesAsPng(images []image.Image, pngPath string) {
	paletted := image.NewPaletted(image.Rect(0, 0, 600, 600), palette.Plan9)

	MARGIN_Y := 10
	destX := 100 // sort of centered
	destY := MARGIN_Y
	for _, chartImage := range images {
		// Draw chartImage starting at Point{destX,destY}
		draw.Draw(paletted, chartImage.Bounds().Add(image.Point{destX, destY}), chartImage,
			chartImage.Bounds().Min, draw.Src)

		// Increase Y by height of chart and margin
		destY += chartImage.Bounds().Max.Y + MARGIN_Y
	}

	outfile, err := os.Create(pngPath)
	if err != nil {
		log.Fatalf("Error from os.Create('%s'): %s", pngPath, err)
	}
	defer outfile.Close()
	err = png.Encode(outfile, paletted)
	if err != nil {
		log.Fatalf("Error from png.Encode: %s", err)
	}
}

func main() {
	config := getConfigFromFlags()

	client, err := prometheus.New(prometheus.Config{
		Address: "http://" + config.prometheusHostPort,
	})
	if err != nil {
		log.Fatalf("Error from prometheus.New: %s", err)
	}
	prometheusApi := prometheus.NewQueryAPI(client)

	log.Printf("Querying Prometheus at http://%s...", config.prometheusHostPort)
	chartImages := []image.Image{}
	prometheusTimeout := time.Duration(1) * time.Second
	c := make(chan model.Matrix, 1)
	go func() {
		c <- queryCloudfrontVisits(prometheusApi)
	}()
	select {
	case cloudfrontVisitsMatrix := <-c:
		chartImages = append(chartImages,
			draw1SeriesChart(cloudfrontVisitsMatrix, "Cloudfront Visits"))
	case <-time.After(prometheusTimeout):
		log.Fatalf("Prometheus timeout after %v: %v",
			prometheusTimeout, config.prometheusHostPort)
	}

	saveChartImagesAsPng(chartImages, config.pngPath)

	if config.doSendEmail {
		sendMail(config.smtpHostPort, config.emailFrom,
			config.emailTo, config.emailSubject, "(see attached image)",
			config.pngPath)
	}
}
