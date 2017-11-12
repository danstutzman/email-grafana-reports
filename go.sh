#!/bin/bash -ex

cd $GOPATH/src/github.com/danielstutzman/email-grafana-reports
go vet github.com/danielstutzman/email-grafana-reports
go install -v -race .

rm -f out.png
$GOPATH/bin/email-grafana-reports \
  -pngPath out.png \
  -influxdbUsername admin \
  -influxdbPassword `cat INFLUXDB_PASSWORD`
open out.png
