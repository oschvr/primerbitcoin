NAME=primerbitcoin

build:
	mkdir build

	GOARCH=amd64 GOOS=linux go build -o build/${NAME}-linux-amd64 cmd/primerbitcoin/main.go
	
	GOARCH=arm64 GOOS=darwin go build -o build/${NAME}-darwin-amd64 cmd/primerbitcoin/main.go

clean:
	go clean
	rm -rf build