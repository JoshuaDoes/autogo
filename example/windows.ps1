$env:GO111MODULE = "off"
del autogo.exe
cd ..\cmd\autogo && (go build -ldflags="-s -w" -o autogo.exe || (cd ..\..\example && Write-Error 'Failed to build autogo.' -ErrorAction Stop)) && mv autogo.exe ..\..\example\ && cd ..\..\example && .\autogo.exe $args
