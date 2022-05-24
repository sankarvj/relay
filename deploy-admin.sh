#!/bin/bash
echo "Entering project relay-admin..."
cd cmd/relay-admin

echo "Building linux compatible project..."
GOOS=linux GOARCH=amd64 go build -o bin/application

echo "Zipping all contents.."
today=$(date +"%Y-%m-%d")
zipFileName="relay-admin-${today}.zip"

mkdir configurations
cp ../../go.mod .
cp ../../config/prod/* configurations
cp ../../private.pem configurations
cp -r ../../templates .

zip -r $zipFileName bin .ebextensions configurations templates go.mod

echo "Deleting copied contents"
rm -rf go.mod
rm -rf configurations
rm -rf bin
rm -rf templates


cd ../../

echo "Your admin build is ready!"
