BINARY_NAME=ptlm
MAIN_PATH=cmd/ptlm/main.go
BUILD_DIR=build
VERSION=1.0.1
LDFLAGS=-ldflags "-s -w -X printcode2llm/internal/version.Version=$(VERSION)"

ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    RM := cmd /C del /Q /F
    RMDIR := cmd /C rmdir /S /Q
    MKDIR := cmd /C mkdir
    EXE := .exe
    SEP := \\
else
    DETECTED_OS := $(shell uname -s)
    RM := rm -f
    RMDIR := rm -rf
    MKDIR := mkdir -p
    EXE :=
    SEP := /
endif

.PHONY: all clean install help test run \
        windows linux darwin \
        windows-386 windows-amd64 windows-arm64 \
        linux-386 linux-amd64 linux-arm linux-arm64 linux-loong64 \
        darwin-amd64 darwin-arm64

all: windows linux darwin
	@echo All platforms compiled successfully!
	@echo Output directory: $(BUILD_DIR)$(SEP)

windows: windows-386 windows-amd64 windows-arm64
linux: linux-386 linux-amd64 linux-arm linux-arm64 linux-loong64
darwin: darwin-amd64 darwin-arm64

ifeq ($(DETECTED_OS),Windows)

windows-386:
	@echo Building Windows 32-bit...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=windows&& set GOARCH=386&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-windows-386.exe $(MAIN_PATH)

windows-amd64:
	@echo Building Windows 64-bit...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=windows&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

windows-arm64:
	@echo Building Windows ARM64...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=windows&& set GOARCH=arm64&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-windows-arm64.exe $(MAIN_PATH)

linux-386:
	@echo Building Linux 32-bit...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=linux&& set GOARCH=386&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-linux-386 $(MAIN_PATH)

linux-amd64:
	@echo Building Linux 64-bit...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=linux&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

linux-arm:
	@echo Building Linux ARM...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=linux&& set GOARCH=arm&& set GOARM=7&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-linux-arm $(MAIN_PATH)

linux-arm64:
	@echo Building Linux ARM64...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=linux&& set GOARCH=arm64&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

linux-loong64:
	@echo Building Linux LoongArch64...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=linux&& set GOARCH=loong64&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-linux-loong64 $(MAIN_PATH)

darwin-amd64:
	@echo Building macOS Intel...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=darwin&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

darwin-arm64:
	@echo Building macOS Apple Silicon...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@set CGO_ENABLED=0&& set GOOS=darwin&& set GOARCH=arm64&& go build $(LDFLAGS) -o $(BUILD_DIR)$(SEP)$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

else

windows-386:
	@echo "Building Windows 32-bit..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=windows GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-386.exe $(MAIN_PATH)

windows-amd64:
	@echo "Building Windows 64-bit..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

windows-arm64:
	@echo "Building Windows ARM64..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe $(MAIN_PATH)

linux-386:
	@echo "Building Linux 32-bit..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386 $(MAIN_PATH)

linux-amd64:
	@echo "Building Linux 64-bit..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

linux-arm:
	@echo "Building Linux ARM..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm $(MAIN_PATH)

linux-arm64:
	@echo "Building Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

linux-loong64:
	@echo "Building Linux LoongArch64..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=loong64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-loong64 $(MAIN_PATH)

darwin-amd64:
	@echo "Building macOS Intel..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

darwin-arm64:
	@echo "Building macOS Apple Silicon..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

endif

build-local:
	@echo "Building for current platform..."
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME)$(EXE) $(MAIN_PATH)
	@echo "Build complete: ./$(BINARY_NAME)$(EXE)"

install: build-local
	@echo "Installing to system..."
	@./$(BINARY_NAME)$(EXE) install

ifeq ($(DETECTED_OS),Windows)
clean:
	@echo Cleaning build files...
	@if exist $(BUILD_DIR) rmdir /S /Q $(BUILD_DIR)
	@if exist $(BINARY_NAME).exe del /Q /F $(BINARY_NAME).exe
	@if exist LLM_CODE*.md del /Q /F LLM_CODE*.md
	@echo Clean complete
else
clean:
	@echo "Cleaning build files..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f LLM_CODE*.md
	@echo "Clean complete"
endif

test:
	@echo "Running tests..."
	@go test -v ./...

run:
	@CGO_ENABLED=0 go run $(MAIN_PATH) $(ARGS)

help:
	@echo PrintCode2LLM Build Tool
	@echo.
	@echo Usage:
	@echo   make              - Build all platforms
	@echo   make build-local  - Build current platform
	@echo   make windows      - Build all Windows versions
	@echo   make linux        - Build all Linux versions
	@echo   make darwin       - Build all macOS versions
	@echo   make install      - Build and install to system
	@echo   make clean        - Clean build files
	@echo   make test         - Run tests
	@echo   make run ARGS=.   - Run directly
	@echo   make help         - Show this help
	@echo.
	@echo Current OS: $(DETECTED_OS)