.PHONY: build lint test cover clean

build:
	go build -v ./...

lint:
	golangci-lint run

test:
	go test -v -race ./...

cover:
	go test -v -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html

clean:
	rm -f coverage.txt coverage.html
