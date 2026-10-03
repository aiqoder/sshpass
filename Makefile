TARGETS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: build clean test fmt

build:
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/sshpass-linux-amd64 ./cmd/sshpass
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/sshpass-linux-arm64 ./cmd/sshpass
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/sshpass-darwin-amd64 ./cmd/sshpass
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o build/sshpass-darwin-arm64 ./cmd/sshpass
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/sshpass-windows-amd64.exe ./cmd/sshpass
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o build/sshpass-windows-arm64.exe ./cmd/sshpass

clean:
	rm -rf build/

test:
	CGO_ENABLED=0 go test ./...

fmt:
	gofmt -w .
