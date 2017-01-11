package main

import (
	"flag"
	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
	chart "github.com/wcharczuk/go-chart"
	"golang.org/x/net/context"
	"image"
	"log"
	"math"
	"regexp"
	"sync"
	"time"
)

const UNIX_MILLIS_TO_UNIX_NANOS = 1000 * 1000
const HOST_AND_PORT_REGEXP = "^[a-z0-9-]+:[0-9]+)$"
const NUM_CHART_QUERIES_AT_ONCE = 3
const PROMETHEUS_TIMEOUT_MILLIS = 1000

type Config struct {
	pngPath            string
	prometheusHostPort string
	emailFrom          string
	emailTo            string
	emailSubject       string
	smtpHostPort       string
	doSendEmail        bool
}

type Query struct {
	expression    string
	yAxisTitle    string
	setYRangeTo01 bool
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

func queryOrFatal(api prometheus.QueryAPI, expression string) model.Matrix {
	c := make(chan model.Matrix, 1)
	go func() {
		value, err := api.QueryRange(context.TODO(), expression, prometheus.Range{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
			Step:  20 * time.Minute,
		})
		if err != nil {
			log.Fatalf("Error from api.QueryRange: %s", err)
		}
		if value.Type() != model.ValMatrix {
			log.Fatalf("Expected value.Type() == ValMatrix but got %d", value.Type())
		}
		c <- value.(model.Matrix)
	}()

	timeout := PROMETHEUS_TIMEOUT_MILLIS * time.Millisecond
	select {
	case matrix := <-c:
		return matrix
	case <-time.After(timeout):
		log.Fatalf("Prometheus timeout after %v nanoseconds", timeout)
		return nil
	}
}

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

func main() {
	config := getConfigFromFlags()

	client, err := prometheus.New(prometheus.Config{
		Address: "http://" + config.prometheusHostPort,
	})
	if err != nil {
		log.Fatalf("Error from prometheus.New: %s", err)
	}
	prometheusApi := prometheus.NewQueryAPI(client)

	queries := []Query{{
		expression:    `cloudfront_visits{site_name="vocabincontext.com",status="200"}`,
		yAxisTitle:    "CloudFront Visits",
		setYRangeTo01: false,
	}, {
		expression:    `1 - irate(node_cpu{mode="idle"}[5m])`,
		yAxisTitle:    "CPU",
		setYRangeTo01: true,
	}}

	queriesChan := make(chan Query, NUM_CHART_QUERIES_AT_ONCE)
	go func() {
		for _, query := range queries {
			queriesChan <- query
		}
		close(queriesChan)
	}()

	log.Printf("Querying Prometheus at http://%s...", config.prometheusHostPort)
	queryToMatrix := map[Query]model.Matrix{}
	var wg sync.WaitGroup
	wg.Add(len(queries))
	for query := range queriesChan {
		go func(query Query) {
			matrix := queryOrFatal(prometheusApi, query.expression)
			queryToMatrix[query] = matrix
			wg.Done()
		}(query)
	}
	wg.Wait()

	multichart := NewMultiChart()
	for _, query := range queries {
		matrix := queryToMatrix[query]
		multichart.CopyChart(drawChart(matrix, query.yAxisTitle, query.setYRangeTo01))
	}

	log.Printf("Writing %s", config.pngPath)
	multichart.SaveToPng(config.pngPath)

	if config.doSendEmail {
		sendMail(config.smtpHostPort, config.emailFrom,
			config.emailTo, config.emailSubject, "(see attached image)",
			config.pngPath)
	}
}
