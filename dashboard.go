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
	DsType      string     `json:"dsType"`
	Query       string     `json:"query"`
	Selects     [][]Select `json:"select"`
	Measurement string     `json:"measurement"`
	Tags        []Tag      `json:"tags"`
	GroupBys    []GroupBy  `json:"groupBy"`
}

type Select struct {
	Params []string `json:"params"`
	Type   string   `json:"type"`
}

type YAxis struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

type Tag struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type GroupBy struct {
	Params []string `json:"params"`
	Type   string   `json:"type"`
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
