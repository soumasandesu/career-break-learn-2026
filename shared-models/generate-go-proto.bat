@echo off
REM Generate Go code from proto files
set PROTO_DIR=src\main\proto
set OUTPUT_DIR=..\user-service-golang\proto

REM Create output directory if it doesn't exist
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM Generate Go code for each proto file
protoc --go_out=%OUTPUT_DIR% --go_opt=paths=source_relative --proto_path=%PROTO_DIR% %PROTO_DIR%\*.proto

echo Go proto files generated in %OUTPUT_DIR%
