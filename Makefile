
.PHONY: test, tidy

tidy:
	go mod tidy

test:
	go test -v --race ./...
