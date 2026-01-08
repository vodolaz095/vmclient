tools:
	@which go
	@go version

# https://go.dev/blog/govulncheck
# install it by go install golang.org/x/vuln/cmd/govulncheck@latest
vuln:
	which govulncheck
	govulncheck ./...

deps:
	go mod download
	go mod verify
	go mod tidy

test:
	go test -v ./...

include make/*.mk

push:
	go run examples/push/main.go

instant:
	go run examples/instant/main.go

range:
	go run examples/range/main.go
