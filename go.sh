#!/bin/bash -ex
go vet .
go install .
rm -f out.png
$GOPATH/bin/prometheus-email-reports
open out.png
