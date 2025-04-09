# Output binary name
BIN := auth

# Main source package
CMD := ./cmd/auth

# Default target: build
all: build

# Build the binary
build:
	go build -o $(BIN) $(CMD)

# Run the app on an image
run:
	./$(BIN) qr.png

# Run all tests (verbose)
test:
	go test -v ./...

# Format code with goimports
fmt:
	goimports -w .

# Remove build artifacts and test-generated files
clean:
	rm -f $(BIN) test_*.png *.test *.cover

# Run all checks
check: fmt test build

.PHONY: all build run test fmt clean check
