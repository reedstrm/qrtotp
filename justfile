# Build the CLI binary
build:
	go build -o auth ./cmd/auth

# Run the CLI on a test image
run IMAGE:
	./auth {{IMAGE}}

# Run tests with verbose output
test:
	go test -v ./cmd/auth

# Format code using goimports
fmt:
	goimports -w .

# Remove build artifacts and test QR images
clean:
	rm -f auth test_*.png

# Run everything for a sanity check
check:
	just fmt
	just test
	just build
