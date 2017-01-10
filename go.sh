#!/bin/bash -ex
go install .

rm -f out.png
$GOPATH/bin/prometheus-email-reports -pngPath out.png -prometheusHostPort localhost:9090
open out.png
