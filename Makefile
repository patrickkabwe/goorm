build:
	go build -v ./...

test:
	go test -v -cover .
format:
	go fmt .