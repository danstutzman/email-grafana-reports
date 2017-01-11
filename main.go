package main

import (
	"flag"
	"github.com/prometheus/client_golang/api/prometheus"
	"golang.org/x/net/context"
	"log"
	"regexp"
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
	log.Printf("Querying Prometheus at http://%s...", config.prometheusHostPort)
	queryToMatrix := doQueries(context.Background(), queries, prometheusApi,
		NUM_CHART_QUERIES_AT_ONCE, PROMETHEUS_TIMEOUT_MILLIS*time.Millisecond)

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
