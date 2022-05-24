cls
$env:GO111MODULE = "off"
$env:AUTOGOEX = "$PWD"
cd ..\cmd\autogo && (go build -ldflags="-s -w" -o $env:AUTOGOEX\autogo.exe || (cd $env:AUTOGOEX && Write-Error 'Failed to build autogo.' -ErrorAction Stop)) && cd $env:AUTOGOEX && .\autogo.exe $args
