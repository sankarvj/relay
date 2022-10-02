#!/bin/bash
echo "Entering project relay-events..."
cd cmd/relay-events

echo "Building linux compatible project..."
GOOS=linux GOARCH=amd64 go build main.go

mkdir configurations
cp ../../config/prod/* configurations
cp ../../private.pem configurations

echo "Zipping all contents..."
zip -r function.zip main configurations 


echo "Clean contents"
rm -rf main 
rm -rf configurations

cd ../../
echo "Your events function is ready!"