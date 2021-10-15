.PHONY: build clean deploy

build:
	go get ./...
	go mod vendor
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/hello hello/main.go
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/world world/main.go

clean:
	rm -rf ./bin ./vendor

deploy: clean build
	sls deploy --verbose
