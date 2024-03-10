dev: test

test:
	go test ./... -v

build: test

	go build -o ./build/pass-man ./cmd/main.go
