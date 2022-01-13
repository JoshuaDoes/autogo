$env:GO111MODULE = "off"
$env:AUTOGOEX = "$PWD"
cd ..\cmd\autogo && (go1.18beta1 build -ldflags="-s -w" -o $env:AUTOGOEX\autogo.exe || (cd ..\..\example && Write-Error 'Failed to build autogo.' -ErrorAction Stop)) && cd ..\..\example && .\autogo.exe $args
