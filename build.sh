#!/bin/bash

BINARY_NAME=ptlm
MAIN_PATH=cmd/ptlm/main.go
BUILD_DIR=build
VERSION=1.0.1
LDFLAGS="-ldflags -s -w -X printcode2llm/internal/version.Version=$VERSION"

build_windows() {
    echo "编译 Windows 平台..."
    mkdir -p $BUILD_DIR
    GOOS=windows GOARCH=386 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-windows-386.exe $MAIN_PATH
    GOOS=windows GOARCH=amd64 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-windows-amd64.exe $MAIN_PATH
    GOOS=windows GOARCH=arm64 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-windows-arm64.exe $MAIN_PATH
}

build_linux() {
    echo "编译 Linux 平台..."
    mkdir -p $BUILD_DIR
    GOOS=linux GOARCH=386 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-linux-386 $MAIN_PATH
    GOOS=linux GOARCH=amd64 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-linux-amd64 $MAIN_PATH
    GOOS=linux GOARCH=arm GOARM=7 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-linux-arm $MAIN_PATH
    GOOS=linux GOARCH=arm64 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-linux-arm64 $MAIN_PATH
    GOOS=linux GOARCH=loong64 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-linux-loong64 $MAIN_PATH
}

build_darwin() {
    echo "编译 macOS 平台..."
    mkdir -p $BUILD_DIR
    GOOS=darwin GOARCH=amd64 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-darwin-amd64 $MAIN_PATH
    GOOS=darwin GOARCH=arm64 go build $LDFLAGS -o $BUILD_DIR/$BINARY_NAME-darwin-arm64 $MAIN_PATH
}

case "$1" in
    all|"")
        echo "开始编译所有平台..."
        build_windows
        build_linux
        build_darwin
        echo "所有平台编译完成！"
        ;;
    windows)
        build_windows
        ;;
    linux)
        build_linux
        ;;
    darwin|macos)
        build_darwin
        ;;
    local)
        echo "编译当前平台..."
        go build $LDFLAGS -o $BINARY_NAME $MAIN_PATH
        echo "编译完成: ./$BINARY_NAME"
        ;;
    clean)
        echo "清理构建文件..."
        rm -rf $BUILD_DIR
        rm -f $BINARY_NAME
        rm -f LLM_CODE*.md
        echo "清理完成"
        ;;
    *)
        echo "用法: $0 [all|local|windows|linux|darwin|clean]"
        ;;
esac