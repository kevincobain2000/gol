#! /bin/bash

cd frontend
npm install
npm run build
cd ..

go build main.go
