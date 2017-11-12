package main

import (
	"flag"
	"log"
	"strings"
	"time"

	clientPkg "github.com/influxdata/influxdb/client/v2"
)

const UNIX_MILLIS_TO_UNIX_NANOS = 1000 * 1000
const NUM_CHART_QUERIES_AT_ONCE = 3
const PROMETHEUS_TIMEOUT_MILLIS = 1000

type Config struct {
	pngPath          string
	influxdbHostname string
	influxdbPort     string
	influxdbUsername string
	influxdbPassword string
	emailFrom        string
	emailTo          string
	emailSubject     string
	smtpHostPort     string
	doSendEmail      bool
}

type Point struct {
	Time  time.Time
	Value float64
}

func getConfigFromFlags() Config {
	config := Config{}
	flag.StringVar(&config.pngPath, "pngPath", "", "Path to save .png image to")
	flag.StringVar(&config.influxdbHostname, "influxdbHostname", "localhost", "Hostname for InfluxDB")
	flag.StringVar(&config.influxdbPort, "influxdbPort", "8086", "Port for InfluxDB")
	flag.StringVar(&config.influxdbUsername, "influxdbUsername", "admin", "Username for InfluxDB, e.g. admin")
	flag.StringVar(&config.influxdbPassword, "influxdbPassword", "", "Password for InfluxDB")
	flag.StringVar(&config.emailFrom, "emailFrom", "", "Email address to send report from; e.g. Reports <reports@monitoring.danstutzman.com>")
	flag.StringVar(&config.emailTo, "emailTo", "", "Email address to send report to")
	flag.StringVar(&config.emailSubject, "emailSubject", "", "Subject for email report")
	flag.StringVar(&config.smtpHostPort, "smtpHostPort", "",
		"Hostname and port for SMTP server; e.g. localhost:25")
	flag.Parse()

	if config.pngPath == "" {
		log.Fatalf("You must specify -pngPath; try ./out.png")
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

	client, err := clientPkg.NewHTTPClient(clientPkg.HTTPConfig{
		Addr:     "http://" + config.influxdbHostname + ":" + config.influxdbPort,
		Username: config.influxdbUsername,
		Password: config.influxdbPassword,
	})
	if err != nil {
		log.Fatalf("Error from NewHTTPClient: %s", err)
	}

	command := "SELECT count(status) FROM \"belugacdn_logs\" WHERE time > now() - 1d GROUP BY time(1h) fill(null);"
	command = strings.Replace(command, "$timeFilter", "time > now() - 10m", 1)
	command = strings.Replace(command, "$__interval", "1m", 1)

	points := query(client, "mydb", command)

	multichart := NewMultiChart()
	multichart.CopyChart(drawChart(points, "BelugaCDN logs", false))

	log.Printf("Writing %s", config.pngPath)
	multichart.SaveToPng(config.pngPath)

	if config.doSendEmail {
		sendMail(config.smtpHostPort, config.emailFrom,
			config.emailTo, config.emailSubject, "(see attached image)",
			config.pngPath)
	}
}
