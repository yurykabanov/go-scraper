.PHONY: all clean

VERSION          = $(shell git describe --tags 2>/dev/null || echo "unknown")
BUILD            = $(shell git rev-parse HEAD 2>/dev/null)
LDFLAGS          = -X main.Version=$(VERSION) -X main.Build=$(BUILD)
BUILD_DIR        = build

all:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/parser cmd/parser/main.go

clean:
	-rm -f $(BUILD_DIR)/*
