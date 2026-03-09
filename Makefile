BINARY_NAME=golauncher
PREFIX=/usr/local

build:
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) main.go

install: build
	@echo "Installing to $(PREFIX)/bin..."
	@sudo install -m 755 $(BINARY_NAME) $(PREFIX)/bin/

uninstall:
	@echo "Removing $(BINARY_NAME) from $(PREFIX)/bin..."
	@sudo rm $(PREFIX)/bin/$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)