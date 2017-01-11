package main

import (
	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
	"golang.org/x/net/context"
	"log"
	"sync"
	"time"
)

type Query struct {
	expression    string
	yAxisTitle    string
	setYRangeTo01 bool
}

func queryOrFatal(ctx context.Context, api prometheus.QueryAPI, expression string,
	timeoutDuration time.Duration) model.Matrix {

	c := make(chan model.Matrix, 1)
	go func() {
		value, err := api.QueryRange(ctx, expression, prometheus.Range{
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

	select {
	case matrix := <-c:
		return matrix
	case <-time.After(timeoutDuration):
		log.Fatalf("Prometheus timeout after %v", timeoutDuration)
		return nil
	}
}

func doQueries(ctx context.Context, queries []Query, prometheusApi prometheus.QueryAPI,
	numQueriesAtOnce int, timeoutDuration time.Duration) map[Query]model.Matrix {

	queriesChan := make(chan Query, numQueriesAtOnce)
	go func() {
		for _, query := range queries {
			queriesChan <- query
		}
		close(queriesChan)
	}()

	queryToMatrix := map[Query]model.Matrix{}
	var wg sync.WaitGroup
	wg.Add(len(queries))
	for query := range queriesChan {
		go func(query Query) {
			matrix := queryOrFatal(ctx, prometheusApi, query.expression, timeoutDuration)
			queryToMatrix[query] = matrix
			wg.Done()
		}(query)
	}
	wg.Wait()

	return queryToMatrix
}
