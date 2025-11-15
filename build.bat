@echo off
chcp 65001 >nul
setlocal

set BINARY_NAME=ptlm
set MAIN_PATH=cmd\ptlm\main.go
set BUILD_DIR=build
set VERSION=1.0.1
set LDFLAGS=-ldflags "-s -w -X printcode2llm/internal/version.Version=%VERSION%"

if "%1"=="" goto all
if /i "%1"=="all" goto all
if /i "%1"=="local" goto local
if /i "%1"=="windows" goto windows
if /i "%1"=="linux" goto linux
if /i "%1"=="darwin" goto darwin
if /i "%1"=="clean" goto clean
if /i "%1"=="help" goto help
goto help

:all
echo 开始编译所有平台...
echo.
call :windows
call :linux
call :darwin
echo.
echo 所有平台编译完成！
goto end

:windows
echo 编译 Windows 平台...
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
set GOOS=windows& set GOARCH=386& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-windows-386.exe %MAIN_PATH%
set GOOS=windows& set GOARCH=amd64& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-windows-amd64.exe %MAIN_PATH%
set GOOS=windows& set GOARCH=arm64& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-windows-arm64.exe %MAIN_PATH%
goto end

:linux
echo 编译 Linux 平台...
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
set GOOS=linux& set GOARCH=386& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-linux-386 %MAIN_PATH%
set GOOS=linux& set GOARCH=amd64& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-linux-amd64 %MAIN_PATH%
set GOOS=linux& set GOARCH=arm& set GOARM=7& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-linux-arm %MAIN_PATH%
set GOOS=linux& set GOARCH=arm64& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-linux-arm64 %MAIN_PATH%
set GOOS=linux& set GOARCH=loong64& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-linux-loong64 %MAIN_PATH%
goto end

:darwin
echo 编译 macOS 平台...
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
set GOOS=darwin& set GOARCH=amd64& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-darwin-amd64 %MAIN_PATH%
set GOOS=darwin& set GOARCH=arm64& go build %LDFLAGS% -o %BUILD_DIR%\%BINARY_NAME%-darwin-arm64 %MAIN_PATH%
goto end

:local
echo 编译当前平台...
go build %LDFLAGS% -o %BINARY_NAME%.exe %MAIN_PATH%
echo 编译完成: %BINARY_NAME%.exe
goto end

:clean
echo 清理构建文件...
if exist %BUILD_DIR% rmdir /S /Q %BUILD_DIR%
if exist %BINARY_NAME%.exe del /Q /F %BINARY_NAME%.exe
if exist LLM_CODE*.md del /Q /F LLM_CODE*.md
echo 清理完成
goto end

:help
echo PrintCode2LLM 构建工具
echo.
echo 用法:
echo   build.bat           - 编译所有平台
echo   build.bat local     - 编译当前平台
echo   build.bat windows   - 编译所有 Windows 版本
echo   build.bat linux     - 编译所有 Linux 版本
echo   build.bat darwin    - 编译所有 macOS 版本
echo   build.bat clean     - 清理构建文件
echo   build.bat help      - 显示此帮助
goto end

:end
endlocal