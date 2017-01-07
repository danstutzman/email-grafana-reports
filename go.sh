#!/bin/bash -ex
pushd $GOPATH/src/github.com/danielstutzman/prometheus-email-reports
go vet .
pushd

go install .
rm -f out.png
$GOPATH/bin/prometheus-email-reports
open out.png
