lint:
	golangci-lint run -v

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.txt ./... && go tool cover -html=coverage.txt