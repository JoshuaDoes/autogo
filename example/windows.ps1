$ErrorActionPreference = "Stop"
cd ..
go build -ldflags="-s -w" -o example\autogo.exe
cd example
.\autogo.exe $args
