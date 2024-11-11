#! /bin/bash

cd frontend
npm install
npm run build

go build main.go

cd ..
