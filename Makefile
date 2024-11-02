# Set the output binary name
BINARY_NAME=narr

# Default target
.PHONY: all
all: vet fmt build

# Format the Go code
.PHONY: fmt
fmt:
	go fmt ./...

# Run static analysis (vet)
.PHONY: vet
vet:
	go vet ./...
	golint ./...
	staticcheck ./...

# Run vuln check
.PHONY: vulncheck
vulncheck:
	govulncheck ./...	


# Run tests
.PHONY: test
test:
	go test ./...

# Build the Go project
.PHONY: build
build:
	go build -o $(BINARY_NAME) .

# Run the Go project
.PHONY: run
run: build
	./$(BINARY_NAME)

# Clean up the binary
.PHONY: clean
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f demo/*.m4b
	rm -f demo/filelist.txt
	rm -rf demo/m4a
	rm -rf demo/*.chapters.txt

# Run the Go project (shortcut)
.PHONY: dev
dev: fmt vet run

.PHONY: generate-test-files
generate-test-files:
	go run ./bin test.m4a
