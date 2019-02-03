#!/bin/bash -e

echo "Building testing binary and running tests..."
#Get into the right directory
cd $(dirname $0)

export NETGEAR_EXPORTER_URL="https://192.168.0.1"
export NETGEAR_EXPORTER_INSECURE="true"
export NETGEAR_EXPORTER_PASSWORD=`cat exporter_password`

export GOOS=""
export GOARCH=""

#Add this directory to PATH
export PATH="$PATH:`pwd`"

go build -o "netgear_exporter" ../

echo "Running tests..."
cd ../

go test
