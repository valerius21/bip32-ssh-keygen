.PHONY: build test lint clean release-dry-run

build:
	CGO_ENABLED=0 go build -o bip32-ssh-keygen .

test:
	CGO_ENABLED=0 go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

clean:
	rm -f bip32-ssh-keygen coverage.out

release-dry-run:
	goreleaser release --snapshot --clean