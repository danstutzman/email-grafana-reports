package main

import (
	"encoding/json"
	"log"
	"time"

	clientPkg "github.com/influxdata/influxdb/client/v2"
)

func query(client clientPkg.Client, databaseName, command string) []Point {
	q := clientPkg.Query{
		Command:   command,
		Database:  databaseName,
		Precision: "ns",
	}
	response, err := client.Query(q)
	if err != nil {
		log.Fatalf("Error from Query: %s", err)
	}

	if response.Error() != nil {
		log.Fatalf("Error from query: %s", response.Error())
	}

	if len(response.Results) != 1 {
		log.Fatalf("Expected len(Results) to be 1, but was %d", len(response.Results))
	}
	result := response.Results[0]

	if len(result.Messages) > 0 {
		log.Fatalf("Unexpected messages in result: %v", result.Messages)
	}
	if len(result.Err) > 0 {
		log.Fatalf("Unexpected Err in result: %v", result.Err)
	}

	if len(result.Series) != 1 {
		log.Fatalf("Expected len(Series) to be 1, but was %d", len(result.Series))
	}
	series := result.Series[0]

	if len(series.Columns) != 2 {
		log.Fatalf("Expected len(Columns) to be 2, but was %d", len(series.Columns))
	}
	if series.Columns[0] != "time" {
		log.Fatalf("Expected Columns[0] to be 'time', but was %s", series.Columns[0])
	}

	points := []Point{}
	for _, row := range series.Values {
		timeNanos, err := row[0].(json.Number).Int64()
		if err != nil {
			log.Fatalf("Error from Float64 of %v", row[0])
		}

		value, err := row[1].(json.Number).Float64()
		if err != nil {
			log.Fatalf("Error from Float64 of %v", row[1])
		}

		point := Point{
			Time:  time.Unix(0, timeNanos).UTC(),
			Value: value,
		}
		points = append(points, point)
	}

	return points
}
