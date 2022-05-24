clear
export GO111MODULE=off
export AUTOGOEX=$PWD
cd ../cmd/autogo && go build -ldflags="-s -w" -o $AUTOGOEX/autogo && cd $AUTOGOEX && ./autogo "$@"
