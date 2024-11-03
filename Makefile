build:
	go build -v ./...

test:
	POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/goorm-test?sslmode=disable" go test -v -cover .
format:
	go fmt .