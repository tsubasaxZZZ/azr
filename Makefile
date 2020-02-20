export GO111MODULE=on

.PHONY: build
build:
	go build -v .
	go mod tidy
	GOOS=windows GOARCH=amd64 go build -v .

.PHONY: test
test: ## go test
	go get github.com/rakyll/statik
	go generate
	go mod tidy
	go test -v -cover .

.PHONY: clean
clean: ## go clean
	go mod tidy
	go clean -cache -testcache

.PHONY: analyze
analyze: ## do static code analysis
	goimports -l -w .
	go vet ./...
	golint ./...

.PHONY: all
all: test build
