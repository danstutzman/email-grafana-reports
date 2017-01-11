#!/bin/bash -ex
go install -race .

rm -f out.png
$GOPATH/bin/prometheus-email-reports -pngPath out.png -prometheusHostPort localhost:9090
open out.png
