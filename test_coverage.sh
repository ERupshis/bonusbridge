#!/bin/bash

go test -tags ignoretests -coverprofile=coverage.out ./...
go tool cover -html=coverage.out