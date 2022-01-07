$env:GO111MODULE = "off"
cd .. && (go build -ldflags="-s -w" -o example\autogo.exe || (cd example && Write-Error 'Failed to build autogo.' -ErrorAction Stop)) && cd example && .\autogo.exe $args