@echo off
rem ----- Batch script to build & install go-gb on windows without gnu -----

set SOURCE_FILES=gb.go build.go deps.go gofmt.go goinstall.go make.go pkg.go runext.go

if exist %GOBIN%/8g.exe %GOBIN%/8g.exe %SOURCE_FILES% & %GOBIN%/8l -o gb.exe gb.8
if exist %GOBIN%/6g.exe %GOBIN%/6g.exe %SOURCE_FILES% & %GOBIN%/6l -o gb.exe gb.6

move gb.exe %GOBIN%
