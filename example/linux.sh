export GO111MODULE=off
cd .. && go build -ldflags="-s -w" -o example/autogo && cd example && ./autogo "$@"