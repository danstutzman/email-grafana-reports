package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type Dashboard struct {
	Title string `json:"title"`
	Rows  []Row  `json:"rows"`
}

type Row struct {
	Panels []Panel `json:"panels"`
}

type Panel struct {
	Title      string   `json:"title"`
	Targets    []Target `json:"targets"`
	DataSource string   `json:"datasource"`
	YAxes      []YAxis  `json:"yaxes"`
}

type Target struct {
	DsType      string `json:"dsType"`
	Query       string `json:"query"`
	Measurement string `json:"measurement"`
}

type YAxis struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

func parseDashboardsJson(reader io.Reader) []Dashboard {
	dashboards := []Dashboard{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		dashboard := Dashboard{}
		err := json.Unmarshal(scanner.Bytes(), &dashboard)
		if err != nil {
			log.Fatalf("Error from Unmarshal: %s", err)
		}
		dashboards = append(dashboards, dashboard)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error from scanner.Err(): %s", err)
	}

	return dashboards
}
