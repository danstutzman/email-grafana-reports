package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type Dashboard struct {
	Title string `json:"title"`
}

func parseDashboardsJson(reader io.Reader) []Dashboard {
	dashboards := []Dashboard{}
	scanner := bufio.NewScanner(reader)
	dashboard := Dashboard{}
	for scanner.Scan() {
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
