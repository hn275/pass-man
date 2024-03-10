dev: test
	go run ./cmd/main.go new usernametest sitetest

test:
	go test ./... -v

build: test
	go build -o ./build/pass-man ./cmd/main.go
