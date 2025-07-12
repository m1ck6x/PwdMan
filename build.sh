#!/bin/bash

echo "Building..."

cd src
go build -o ../release/PwdMan .

echo "Build done!"
