package main

import (
	"encoding/json"
	"log"
	"time"

	clientPkg "github.com/influxdata/influxdb/client/v2"
)

func query(client clientPkg.Client, databaseName, command string) []Point {
	log.Printf("Query is %s", command)

	q := clientPkg.Query{
		Command:   command,
		Database:  databaseName,
		Precision: "ns",
	}
	response, err := client.Query(q)
	if err != nil {
		log.Fatalf("Error from Query with command %s: %s", command, err)
	}

	if response.Error() != nil {
		log.Fatalf("Error from Error with command %s: %s", command, response.Error())
	}

	if len(response.Results) != 1 {
		log.Fatalf("Expected len(Results) to be 1, but was %d in command %s", len(response.Results), command)
	}
	result := response.Results[0]

	if len(result.Messages) > 0 {
		log.Fatalf("Unexpected messages in result for command %s: %v", command, result.Messages)
	}
	if len(result.Err) > 0 {
		log.Fatalf("Unexpected Err in result for command %s: %v", command, result.Err)
	}

	points := []Point{}
	// you get multiple series if you union multiple tags
	for _, series := range result.Series {
		if len(series.Columns) != 2 {
			log.Fatalf("Expected len(Columns) to be 2, but was %d in command %s", len(series.Columns), command)
		}
		if series.Columns[0] != "time" {
			log.Fatalf("Expected Columns[0] to be 'time', but was %s in command %s", series.Columns[0], command)
		}

		for _, row := range series.Values {
			timeNanos, err := row[0].(json.Number).Int64()
			if err != nil {
				log.Fatalf("Error from Float64 of %v", row[0])
			}

			if row[1] != nil {
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
		}
	}

	return points
}
